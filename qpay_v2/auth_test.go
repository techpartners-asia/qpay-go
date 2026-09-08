package qpay_v2

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/techpartners-asia/qpay-go/utils"
	"resty.dev/v3"
)

// newMockServer creates a test server that returns configurable auth responses.
func newMockServer(t *testing.T, authCalls *atomic.Int32, refreshCalls *atomic.Int32, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	if handler != nil {
		return httptest.NewServer(handler)
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := qpayLoginResponse{
			TokenType:        "bearer",
			AccessToken:      "test-access-token",
			RefreshToken:     "test-refresh-token",
			ExpiresIn:        time.Now().Add(24 * time.Hour).Unix(),
			RefreshExpiresIn: time.Now().Add(48 * time.Hour).Unix(),
			Scope:            "get_token",
			SessionState:     "test",
		}

		switch r.URL.Path {
		case "/auth/token":
			authCalls.Add(1)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		case "/auth/refresh":
			refreshCalls.Add(1)
			resp.AccessToken = "refreshed-access-token"
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

// newTestQPay returns a client with a token already installed, which is the
// state every API-level test needs: the SDK no longer logs in on its own.
func newTestQPay(endpoint string) *qpay {
	q := newUnauthedQPay(endpoint)
	q.SetToken(Token{AccessToken: "test-access-token", ExpiresAt: time.Now().Add(time.Hour)})
	return q
}

// newUnauthedQPay returns a client with no token installed.
func newUnauthedQPay(endpoint string) *qpay {
	return &qpay{
		endpoint:    endpoint,
		username:    "test-user",
		password:    "test-pass",
		invoiceCode: "TEST",
		client:      resty.New().SetTimeout(5 * time.Second),
	}
}

func TestLogin_ReturnsTokenAndCachesNothing(t *testing.T) {
	var authCalls, refreshCalls atomic.Int32
	srv := newMockServer(t, &authCalls, &refreshCalls, nil)
	defer srv.Close()

	q := newUnauthedQPay(srv.URL)

	tok, err := q.Login(context.Background())
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if tok.AccessToken != "test-access-token" {
		t.Fatalf("unexpected access token %q", tok.AccessToken)
	}
	if tok.RefreshToken != "test-refresh-token" {
		t.Fatalf("unexpected refresh token %q", tok.RefreshToken)
	}
	if tok.ExpiresAt.IsZero() || tok.RefreshExpiresAt.IsZero() {
		t.Fatalf("expiries not resolved: %+v", tok)
	}

	// Login must not install the token: the caller decides what to keep.
	if got := q.Token(); !got.IsZero() {
		t.Fatalf("Login installed a token: %+v", got)
	}

	// A second Login is a second request — no cache to short-circuit it.
	if _, err := q.Login(context.Background()); err != nil {
		t.Fatalf("second login failed: %v", err)
	}
	if authCalls.Load() != 2 {
		t.Fatalf("expected 2 auth calls, got %d", authCalls.Load())
	}
}

func TestLogin_ConcurrentCallsAreNotCoalesced(t *testing.T) {
	var authCalls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authCalls.Add(1)
		time.Sleep(20 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(qpayLoginResponse{
			AccessToken: "test-access-token",
			ExpiresIn:   time.Now().Add(24 * time.Hour).Unix(),
		})
	}))
	defer srv.Close()

	q := newUnauthedQPay(srv.URL)

	// Deduplication moved out of the SDK on purpose, so this documents the
	// contract the caller has to honour: ten Logins are ten requests.
	const n = 10
	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := q.Login(context.Background()); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent login failed: %v", err)
	}

	if authCalls.Load() != n {
		t.Fatalf("expected %d auth calls, got %d", n, authCalls.Load())
	}
}

func TestLogin_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := newUnauthedQPay(srv.URL).Login(ctx); err == nil {
		t.Fatal("expected error for a cancelled context, got nil")
	}
}

func TestLogin_NoAccessTokenIsAnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"token_type":"bearer"}`))
	}))
	defer srv.Close()

	if _, err := newUnauthedQPay(srv.URL).Login(context.Background()); err == nil {
		t.Fatal("expected error for a tokenless 200, got nil")
	}
}

func TestAuthServerDown_ReturnsError(t *testing.T) {
	// Point to a closed server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()

	if _, err := newUnauthedQPay(srv.URL).Login(context.Background()); err == nil {
		t.Fatal("expected error when server is down, got nil")
	}
}

func TestRefresh_UsesRefreshEndpoint(t *testing.T) {
	var authCalls, refreshCalls atomic.Int32
	srv := newMockServer(t, &authCalls, &refreshCalls, nil)
	defer srv.Close()

	tok, err := newUnauthedQPay(srv.URL).Refresh(context.Background(), "test-refresh-token")
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if tok.AccessToken != "refreshed-access-token" {
		t.Fatalf("unexpected access token %q", tok.AccessToken)
	}
	if refreshCalls.Load() != 1 || authCalls.Load() != 0 {
		t.Fatalf("expected 1 refresh and 0 auth calls, got %d and %d",
			refreshCalls.Load(), authCalls.Load())
	}
}

func TestRefresh_RejectedTokenIsErrUnauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"invalid_token"}`))
	}))
	defer srv.Close()

	// A rejected refresh must NOT silently fall back to a full login: the
	// caller decides, because it may share the token with other processes.
	_, err := newUnauthedQPay(srv.URL).Refresh(context.Background(), "stale")
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestRefresh_EmptyTokenRejectedWithoutRequest(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
	}))
	defer srv.Close()

	if _, err := newUnauthedQPay(srv.URL).Refresh(context.Background(), ""); err == nil {
		t.Fatal("expected error for an empty refresh token, got nil")
	}
	if calls.Load() != 0 {
		t.Fatalf("expected no request, got %d", calls.Load())
	}
}

func TestRequest_WithoutTokenFailsBeforeReachingQPay(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
	}))
	defer srv.Close()

	_, err := newUnauthedQPay(srv.URL).GetInvoice("inv-1")
	if !errors.Is(err, ErrNoToken) {
		t.Fatalf("expected ErrNoToken, got %v", err)
	}
	if calls.Load() != 0 {
		t.Fatalf("expected no request to qPay, got %d", calls.Load())
	}
}

func TestRequest_SendsInstalledTokenAndSetTokenReplacesIt(t *testing.T) {
	var seen []string
	var mu sync.Mutex
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		seen = append(seen, r.Header.Get("Authorization"))
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	q := newUnauthedQPay(srv.URL)
	q.SetToken(Token{AccessToken: "first"})
	if _, err := q.GetInvoice("inv-1"); err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	q.SetToken(Token{AccessToken: "second"})
	if _, err := q.GetInvoice("inv-1"); err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	want := []string{"Bearer first", "Bearer second"}
	mu.Lock()
	defer mu.Unlock()
	if len(seen) != len(want) {
		t.Fatalf("expected %d requests, got %d", len(want), len(seen))
	}
	for i := range want {
		if seen[i] != want[i] {
			t.Fatalf("request %d carried %q, want %q", i, seen[i], want[i])
		}
	}
}

func TestRequest_RejectedTokenIsErrUnauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":"forbidden"}`))
	}))
	defer srv.Close()

	// The SDK does not retry: it reports that the token was rejected so the
	// owner of the token can replace it and decide whether to retry.
	_, err := newTestQPay(srv.URL).GetInvoice("inv-1")
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestSetToken_ZeroValueClearsTheToken(t *testing.T) {
	q := newTestQPay("https://qpay.test")
	q.SetToken(Token{})

	if _, err := q.GetInvoice("inv-1"); !errors.Is(err, ErrNoToken) {
		t.Fatalf("expected ErrNoToken after clearing, got %v", err)
	}
}

func TestExpiresAt_UnanchoredExpiryYieldsZeroTime(t *testing.T) {
	// qPay sends expires_in as a Unix timestamp, but some deployments send a
	// plain duration. A duration cannot be anchored, so it must not be turned
	// into a confident expiry the caller would then trust.
	if got := utils.TokenExpiresAt(3600); !got.IsZero() {
		t.Fatalf("expected zero time for a duration-shaped expiry, got %v", got)
	}
	if got := utils.TokenExpiresAt(0); !got.IsZero() {
		t.Fatalf("expected zero time for a missing expiry, got %v", got)
	}

	ts := time.Now().Add(time.Hour).Unix()
	if got := utils.TokenExpiresAt(ts); !got.Equal(time.Unix(ts, 0)) {
		t.Fatalf("expected %v, got %v", time.Unix(ts, 0), got)
	}
}

func TestWithToken_InstallsAtConstruction(t *testing.T) {
	q := New("u", "p", "https://qpay.test", "", "TEST", "m",
		WithToken(Token{AccessToken: "preinstalled"}))

	if got := q.Token().AccessToken; got != "preinstalled" {
		t.Fatalf("expected the preinstalled token, got %q", got)
	}
}

func TestNew_PerformsNoNetworkIO(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
	}))
	defer srv.Close()

	New("u", "p", srv.URL, "", "TEST", "m")

	// Construction used to warm a token in a background goroutine, which made
	// New() a network call in disguise and logged in once per client built.
	time.Sleep(50 * time.Millisecond)
	if calls.Load() != 0 {
		t.Fatalf("New performed %d requests, want 0", calls.Load())
	}
}
