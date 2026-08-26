package qpay_v1

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func writeAuth(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(QpayLoginResponse{
		TokenType:        "bearer",
		AccessToken:      "tok",
		RefreshToken:     "rtok",
		ExpiresIn:        int(time.Now().Add(24 * time.Hour).Unix()),
		RefreshExpiresIn: int(time.Now().Add(48 * time.Hour).Unix()),
	})
}

// newAuthed returns a client with a token already installed, which is the
// state every API-level test needs: the SDK no longer logs in on its own.
func newAuthed(endpoint string) QPay {
	q := New("id", "secret", endpoint, "cb", "m", "tpl", "b", "p")
	q.SetToken(Token{AccessToken: "tok", ExpiresAt: time.Now().Add(time.Hour)})
	return q
}

// A transport failure after a successful auth used to dereference a nil
// *http.Response and crash the caller's process.
func TestTransportErrorReturnsErrorNotPanic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/token" {
			writeAuth(w)
			return
		}
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Errorf("server does not support hijacking")
			return
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			t.Errorf("hijack: %v", err)
			return
		}
		conn.Close() // drop the connection with no response
	}))
	defer srv.Close()

	q := newAuthed(srv.URL)
	if _, err := q.GetInvoice("abc"); err == nil {
		t.Fatal("expected an error from the dropped connection, got nil")
	}
}

// An id carrying a control character made http.NewRequest fail; the error was
// discarded and the nil request was then dereferenced.
func TestMalformedInvoiceIDReturnsErrorNotPanic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/token" {
			writeAuth(w)
			return
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	q := newAuthed(srv.URL)
	// Escaped, this is a single harmless path segment rather than a crash.
	if _, err := q.GetInvoice("abc\ndef"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Ids must stay inside their own path segment.
func TestInvoiceIDCannotRetargetTheRequest(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/token" {
			writeAuth(w)
			return
		}
		gotPath = r.URL.EscapedPath()
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	q := newAuthed(srv.URL)
	if _, err := q.GetInvoice("../payment/check/victim"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(gotPath, "/invoice/") {
		t.Fatalf("request escaped its endpoint: %s", gotPath)
	}
	if strings.Contains(gotPath, "/payment/") {
		t.Fatalf("path traversal reached another endpoint: %s", gotPath)
	}
}

func TestNon200ResponseIsAnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/token" {
			writeAuth(w)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"boom"}`))
	}))
	defer srv.Close()

	q := newAuthed(srv.URL)
	if _, err := q.GetInvoice("abc"); err == nil {
		t.Fatal("expected an error for HTTP 500")
	}
}

func TestGarbageBodyIsAnErrorNotAnEmptyResult(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/token" {
			writeAuth(w)
			return
		}
		_, _ = w.Write([]byte(`<html>gateway error</html>`))
	}))
	defer srv.Close()

	q := newAuthed(srv.URL)
	// Previously this returned a zero-valued response and a nil error, which a
	// caller reads as "payment exists but is unpaid".
	if _, err := q.CheckPayment("pay-1"); err == nil {
		t.Fatal("expected a decode error for a non-JSON body")
	}
}

func TestAuthWithoutAccessTokenIsRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token_type":"bearer"}`))
	}))
	defer srv.Close()

	q := New("id", "secret", srv.URL, "cb", "m", "tpl", "b", "p")
	if _, err := q.Login(context.Background()); err == nil {
		t.Fatal("expected an error when auth returns no access token")
	}
}

func TestConcurrentUseIsRaceFree(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/token" {
			writeAuth(w)
			return
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	q := newAuthed(srv.URL)

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = q.GetInvoice("abc")
		}()
	}
	wg.Wait()
}

func TestRequestWithoutTokenFailsBeforeReachingQPay(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
	}))
	defer srv.Close()

	q := New("id", "secret", srv.URL, "cb", "m", "tpl", "b", "p")
	if _, err := q.GetInvoice("abc"); !errors.Is(err, ErrNoToken) {
		t.Fatalf("expected ErrNoToken, got %v", err)
	}
	if calls != 0 {
		t.Fatalf("expected no request to qPay, got %d", calls)
	}
}

func TestLoginReturnsTokenAndInstallsNothing(t *testing.T) {
	var authCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/token" {
			authCalls++
			writeAuth(w)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	q := New("id", "secret", srv.URL, "cb", "m", "tpl", "b", "p")
	tok, err := q.Login(context.Background())
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if tok.AccessToken != "tok" {
		t.Fatalf("unexpected access token %q", tok.AccessToken)
	}
	// Login must not install the token: the caller decides what to keep.
	if !q.Token().IsZero() {
		t.Fatal("Login installed a token")
	}

	// No cache, so a second Login is a second request.
	if _, err := q.Login(context.Background()); err != nil {
		t.Fatalf("second login failed: %v", err)
	}
	if authCalls != 2 {
		t.Fatalf("expected 2 auth calls, got %d", authCalls)
	}
}

func TestRejectedTokenIsErrUnauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_token"}`))
	}))
	defer srv.Close()

	if _, err := newAuthed(srv.URL).GetInvoice("abc"); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}
