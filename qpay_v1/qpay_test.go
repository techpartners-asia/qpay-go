package qpay_v1

import (
	"encoding/json"
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

	q := New("id", "secret", srv.URL, "cb", "m", "tpl", "b", "p")
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

	q := New("id", "secret", srv.URL, "cb", "m", "tpl", "b", "p")
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

	q := New("id", "secret", srv.URL, "cb", "m", "tpl", "b", "p")
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

	q := New("id", "secret", srv.URL, "cb", "m", "tpl", "b", "p")
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

	q := New("id", "secret", srv.URL, "cb", "m", "tpl", "b", "p")
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
	if _, err := q.GetInvoice("abc"); err == nil {
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

	q := New("id", "secret", srv.URL, "cb", "m", "tpl", "b", "p")

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
