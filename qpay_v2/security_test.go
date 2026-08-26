package qpay_v2

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"resty.dev/v3"
)

func authHandler(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(qpayLoginResponse{
		TokenType:        "bearer",
		AccessToken:      "test-access-token",
		RefreshToken:     "test-refresh-token",
		ExpiresIn:        time.Now().Add(24 * time.Hour).Unix(),
		RefreshExpiresIn: time.Now().Add(48 * time.Hour).Unix(),
	})
}

// Invoice and payment ids often come from end users. They must stay inside
// their own path segment instead of retargeting the request.
func TestURLExtCannotEscapeItsPathSegment(t *testing.T) {
	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/token" {
			authHandler(w)
			return
		}
		gotPath = r.URL.EscapedPath()
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"invoice_id":"x"}`))
	}))
	defer srv.Close()

	q := newTestQPay(srv.URL)

	if _, err := q.GetInvoice("../payment/list"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(gotPath, "/payment/") {
		t.Fatalf("path traversal reached another endpoint: %s", gotPath)
	}
	if !strings.HasPrefix(gotPath, "/invoice/") {
		t.Fatalf("request left its endpoint: %s", gotPath)
	}

	if _, err := q.GetInvoice("abc?merchant_id=other"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotQuery != "" {
		t.Fatalf("id injected a query string: %s", gotQuery)
	}
}

// The id must survive escaping so legitimate lookups still work.
func TestNormalInvoiceIDIsUnchanged(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/token" {
			authHandler(w)
			return
		}
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"invoice_id":"x"}`))
	}))
	defer srv.Close()

	q := newTestQPay(srv.URL)
	const id = "c565079c-11e6-45c2-a802-fc0f2f784993"
	if _, err := q.GetInvoice(id); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/invoice/"+id {
		t.Fatalf("unexpected path: %s", gotPath)
	}
}

// A 3xx that resty surfaces used to pass the >=400 check and return a
// zero-valued response with a nil error.
func TestNonSuccessStatusIsAnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/token" {
			authHandler(w)
			return
		}
		http.Redirect(w, r, "/invoice/elsewhere", http.StatusFound)
	}))
	defer srv.Close()

	q := &qpay{
		endpoint: srv.URL,
		username: "u",
		password: "p",
		// Do not follow the redirect, so the 3xx reaches our status check.
		client: resty.New().
			SetTimeout(5 * time.Second).
			SetRedirectPolicy(resty.RedirectNoPolicy()),
	}

	if _, err := q.GetInvoice("abc"); err == nil {
		t.Fatal("expected an error for a 302 response")
	}
}

func TestAuthWithoutAccessTokenIsRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token_type":"bearer","expires_in":253402300799}`))
	}))
	defer srv.Close()

	q := newUnauthedQPay(srv.URL)
	if _, err := q.Login(context.Background()); err == nil {
		t.Fatal("expected an error when auth returns no access token")
	}
}

// Callback params must not corrupt a callback URL that already has a query.
func TestCreateInvoiceCallbackURLKeepsExistingQuery(t *testing.T) {
	var payload map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/token" {
			authHandler(w)
			return
		}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"invoice_id":"x"}`))
	}))
	defer srv.Close()

	q := newTestQPay(srv.URL)
	q.callback = "https://shop.mn/cb?order=7"

	if _, err := q.CreateInvoice(QPayCreateInvoiceInput{
		SenderInvoiceNo: "1",
		Amount:          100,
		CallbackParam:   map[string]string{"payment_id": "42"},
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cb, _ := payload["callback_url"].(string)
	u, err := url.Parse(cb)
	if err != nil {
		t.Fatalf("callback_url is not a valid URL: %q", cb)
	}
	if strings.Count(cb, "?") != 1 {
		t.Fatalf("callback_url has a duplicated '?': %q", cb)
	}
	if u.Query().Get("order") != "7" || u.Query().Get("payment_id") != "42" {
		t.Fatalf("callback params lost: %q", cb)
	}
}

// qPay is Keycloak-backed and may send not-before-policy as a JSON number.
// A rigid field type there discards an otherwise valid access token and
// takes down every API call.
func TestNumericNotBeforePolicyDoesNotBreakAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token_type":"bearer","access_token":"tok",` +
			`"refresh_token":"rt","expires_in":253402300799,` +
			`"refresh_expires_in":253402300799,"not-before-policy":0,"session_state":"s"}`))
	}))
	defer srv.Close()

	q := newUnauthedQPay(srv.URL)
	res, err := q.Login(context.Background())
	if err != nil {
		t.Fatalf("auth failed on numeric not-before-policy: %v", err)
	}
	if res.AccessToken != "tok" {
		t.Fatalf("access token not decoded: %q", res.AccessToken)
	}
}
