package qpay_v1

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/techpartners-asia/qpay-go/utils"
)

var (
	QPayAuthToken = utils.API{
		Url:    "/auth/token",
		Method: http.MethodPost,
	}
	QPayAuthRefresh = utils.API{
		Url:    "/auth/refresh",
		Method: http.MethodPost,
	}
	QPayPaymentGet = utils.API{
		Url:    "/payment/get/",
		Method: http.MethodGet,
	}
	QPayPaymentCheck = utils.API{
		Url:    "/payment/check/",
		Method: http.MethodGet,
	}

	QPayInvoiceCreate = utils.API{
		Url:    "/bill/create",
		Method: http.MethodPost,
	}
	QPayInvoiceGet = utils.API{
		Url:    "/invoice/",
		Method: http.MethodGet,
	}
)

// doRequest performs a single JSON request and returns the (bounded) body.
// It centralises the error handling that was previously missing: transport
// errors, request-construction errors and body closing.
func (q *qpay) doRequest(method, url string, payload []byte, headers map[string]string) ([]byte, int, error) {
	return q.doRequestCtx(context.Background(), method, url, payload, headers)
}

// doRequestCtx is doRequest bound to a context, for the auth calls the caller
// drives explicitly.
func (q *qpay) doRequestCtx(ctx context.Context, method, url string, payload []byte, headers map[string]string) ([]byte, int, error) {
	var reader io.Reader
	if payload != nil {
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		// Reachable with an id containing a control character or space.
		return nil, 0, fmt.Errorf("qpay: build request: %w", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	res, err := q.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("qpay: request failed: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(io.LimitReader(res.Body, utils.MaxResponseBytes))
	if err != nil {
		return nil, res.StatusCode, fmt.Errorf("qpay: read response: %w", err)
	}

	return body, res.StatusCode, nil
}

func (q *qpay) httpRequestQPay(body interface{}, api utils.API, urlExt string) ([]byte, error) {
	token := q.Token()
	if token.IsZero() {
		// The SDK no longer authenticates behind the caller's back; an absent
		// token is a wiring mistake, not a gateway failure.
		return nil, ErrNoToken
	}

	var payload []byte
	var err error
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("qpay: encode request: %w", err)
		}
	}

	url := q.endpoint + api.Url + utils.EscapePathSegment(urlExt)
	response, status, err := q.doRequest(api.Method, url, payload, map[string]string{
		"Content-Type":  utils.HttpContent,
		"Authorization": "Bearer " + token.AccessToken,
	})
	if err != nil {
		return nil, err
	}

	// A rejected token is its own error so the caller can tell "replace the
	// token and retry" apart from "qPay refused this request".
	if status == http.StatusUnauthorized || status == http.StatusForbidden {
		return nil, fmt.Errorf("%w (Status: %d): %s", ErrUnauthorized,
			status, utils.TruncateForError(string(response)))
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("%s-QPay response error (Status: %d): %s",
			time.Now().Format(utils.TimeFormatYYYYMMDDHHMMSS), status, utils.TruncateForError(string(response)))
	}

	return response, nil
}

// Login [qPay-ээс Access Token авах]
//
// Login performs exactly one request and caches nothing: the returned token is
// the caller's to hold, store and install with [QPay.SetToken].
func (q *qpay) Login(ctx context.Context) (Token, error) {
	payload, err := json.Marshal(&QpayLogin{
		ClientID:     q.client_id,
		ClientSecret: q.client_secret,
		GrantType:    q.grant_type,
	})
	if err != nil {
		return Token{}, fmt.Errorf("qpay: encode auth request: %w", err)
	}

	body, status, err := q.doRequestCtx(ctx, QPayAuthToken.Method, q.endpoint+QPayAuthToken.Url, payload, map[string]string{
		"Content-Type": utils.HttpContent,
	})
	if err != nil {
		return Token{}, err
	}

	if status != http.StatusOK {
		return Token{}, fmt.Errorf("%s-QPay auth response (Status: %d)",
			time.Now().Format(utils.TimeFormatYYYYMMDDHHMMSS), status)
	}

	var authRes QpayLoginResponse
	if err := json.Unmarshal(body, &authRes); err != nil {
		return Token{}, fmt.Errorf("qpay: decode auth response: %w", err)
	}
	if authRes.AccessToken == "" {
		// Returning a tokenless response as success would send every later
		// call out unauthenticated.
		return Token{}, errors.New("qpay: auth response contained no access token")
	}

	return tokenFrom(authRes), nil
}

// Refresh [Refresh token ашиглан access token шинэчлэх]
//
// Refresh is a single request, like [QPay.Login], and does not fall back to a
// full login when the refresh token is rejected: the caller decides.
func (q *qpay) Refresh(ctx context.Context, refreshToken string) (Token, error) {
	if refreshToken == "" {
		return Token{}, errors.New("qpay: refresh token is required")
	}

	body, status, err := q.doRequestCtx(ctx, QPayAuthRefresh.Method, q.endpoint+QPayAuthRefresh.Url, nil, map[string]string{
		"Content-Type":  utils.HttpContent,
		"Authorization": "Bearer " + refreshToken,
	})
	if err != nil {
		return Token{}, err
	}

	if status == http.StatusUnauthorized || status == http.StatusForbidden {
		return Token{}, fmt.Errorf("%w (Status: %d): %s", ErrUnauthorized,
			status, utils.TruncateForError(string(body)))
	}
	if status != http.StatusOK {
		return Token{}, fmt.Errorf("%s-QPay token refresh response (Status: %d)",
			time.Now().Format(utils.TimeFormatYYYYMMDDHHMMSS), status)
	}

	var authRes QpayLoginResponse
	if err := json.Unmarshal(body, &authRes); err != nil {
		return Token{}, fmt.Errorf("qpay: decode refresh response: %w", err)
	}
	if authRes.AccessToken == "" {
		return Token{}, errors.New("qpay: refresh response contained no access token")
	}

	return tokenFrom(authRes), nil
}

// SetToken installs the token subsequent calls will carry. Passing the zero
// Token clears it, which makes the next call fail with [ErrNoToken] rather
// than reach qPay unauthenticated.
func (q *qpay) SetToken(token Token) {
	q.mu.Lock()
	q.token = token
	q.mu.Unlock()
}

// Token returns the installed token.
func (q *qpay) Token() Token {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.token
}
