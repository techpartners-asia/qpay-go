package utils

import (
	"errors"
	"time"
)

// Token is a qPay access token together with everything a caller needs to
// decide when to replace it.
//
// The qPay SDKs do not cache tokens: Login and Refresh each perform exactly
// one HTTP call and return what the gateway said, and SetToken installs the
// token the next request will carry. Whoever owns the client owns the token's
// lifetime — expiry tracking, refresh scheduling and deduplication of
// concurrent logins are all the caller's, because only the caller knows
// whether the token is shared beyond this process.
type Token struct {
	// AccessToken is the bearer sent as Authorization on every API call.
	AccessToken string

	// RefreshToken exchanges for a new access token via the client's Refresh.
	RefreshToken string

	// TokenType is qPay's token_type, normally "Bearer".
	TokenType string

	// ExpiresAt is when AccessToken stops being accepted. It is the ZERO time
	// when qPay reported an expiry we cannot anchor — see [TokenExpiresAt] —
	// and a caller must then treat the token as good for this one use only.
	ExpiresAt time.Time

	// RefreshExpiresAt is when RefreshToken stops being accepted, under the
	// same zero-time caveat as ExpiresAt.
	RefreshExpiresAt time.Time

	// Scope and SessionState are returned verbatim for diagnostics.
	Scope        string
	SessionState string
}

// IsZero reports whether the token carries no credential at all.
func (t Token) IsZero() bool { return t.AccessToken == "" }

var (
	// ErrNoToken is returned when an API call is attempted before a token has
	// been installed with SetToken. The SDK never authenticates on its own, so
	// this is a wiring mistake in the caller, not a gateway failure.
	ErrNoToken = errors.New("qpay: no access token installed: call Login then SetToken")

	// ErrUnauthorized is returned when qPay rejects the installed token with
	// 401 or 403. The caller should discard the token, obtain a new one and
	// retry — the rejected request was never processed by qPay, so retrying
	// cannot double-create an invoice.
	ErrUnauthorized = errors.New("qpay: access token rejected")
)

// TokenExpiresAt turns a qPay expiry field into an absolute time.
//
// qPay documents expires_in as a Unix timestamp. A value too small to be one
// is a relative duration whose issue time we do not know, so it yields the
// zero time rather than a guess: a caller reading a zero ExpiresAt knows not
// to reuse the token, which is what the SDKs' own caches used to do with it.
func TokenExpiresAt(expiresIn int64) time.Time {
	const minPlausibleTimestamp = 1_000_000_000 // 2001-09-09

	if expiresIn < minPlausibleTimestamp {
		return time.Time{}
	}
	return time.Unix(expiresIn, 0)
}
