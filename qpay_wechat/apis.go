package qpay_wechat

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
		Url:    "/payment/check",
		Method: http.MethodPost,
	}
	QPayPaymentCancel = utils.API{
		Url:    "/payment/cancel",
		Method: http.MethodDelete,
	}
	QPayPaymentRefund = utils.API{
		Url:    "/payment/refund/",
		Method: http.MethodDelete,
	}
	QPayPaymentList = utils.API{
		Url:    "/payment/url",
		Method: http.MethodPost,
	}
	QPayInvoiceCreate = utils.API{
		Url:    "/invoice",
		Method: http.MethodPost,
	}
	QPayInvoiceGet = utils.API{
		Url:    "/invoice/",
		Method: http.MethodGet,
	}
	QPayInvoiceCancel = utils.API{
		Url:    "/invoice/",
		Method: http.MethodDelete,
	}
)

// doRequest performs a single JSON request and returns the (bounded) body.
// Transport errors, request-construction errors and body closing are all
// handled here; previously each of them was either ignored or panicked.
func doRequest(client *http.Client, method, url string, payload []byte, headers map[string]string, basic *basicAuth) ([]byte, int, error) {
	return doRequestCtx(context.Background(), client, method, url, payload, headers, basic)
}

// doRequestCtx is doRequest bound to a context, for the auth calls the caller
// drives explicitly.
func doRequestCtx(ctx context.Context, client *http.Client, method, url string, payload []byte, headers map[string]string, basic *basicAuth) ([]byte, int, error) {
	var reader io.Reader
	if payload != nil {
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, 0, fmt.Errorf("qpay: build request: %w", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if basic != nil {
		req.SetBasicAuth(basic.username, basic.password)
	}

	res, err := client.Do(req)
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

// basicAuth carries credentials for the auth endpoint.
type basicAuth struct {
	username string
	password string
}

func (q *qpay_auth) httpRequest(body interface{}, api utils.API, urlExt string) ([]byte, error) {
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
	response, status, err := doRequest(q.client, api.Method, url, payload, map[string]string{
		"Content-Type":  utils.HttpContent,
		"Authorization": "Bearer " + token.AccessToken,
	}, nil)
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
// the caller's to hold, store and install with [QPayAuth.SetToken].
func (q *qpay_auth) Login(ctx context.Context) (Token, error) {
	body, status, err := doRequestCtx(ctx, q.client, QPayAuthToken.Method, q.endpoint+QPayAuthToken.Url, nil, map[string]string{
		"Content-Type": utils.HttpContent,
	}, &basicAuth{username: q.username, password: q.password})
	if err != nil {
		return Token{}, err
	}

	// Rejected credentials are not a provider outage: the caller can tell the
	// two apart and avoid counting a configuration mistake against qPay.
	if status == http.StatusUnauthorized || status == http.StatusForbidden {
		return Token{}, fmt.Errorf("%w (Status: %d): %s", ErrUnauthorized,
			status, utils.TruncateForError(string(body)))
	}
	if status != http.StatusOK {
		return Token{}, fmt.Errorf("%s-QPay auth response (Status: %d)",
			time.Now().Format(utils.TimeFormatYYYYMMDDHHMMSS), status)
	}

	var authRes qpayLoginResponse
	if err := json.Unmarshal(body, &authRes); err != nil {
		return Token{}, fmt.Errorf("qpay: decode auth response: %w", err)
	}
	if authRes.AccessToken == "" {
		return Token{}, errors.New("qpay: auth response contained no access token")
	}

	return tokenFrom(authRes), nil
}

// Refresh [Refresh token ашиглан access token шинэчлэх]
//
// Refresh is a single request, like [QPayAuth.Login], and does not fall back to
// a full login when the refresh token is rejected: the caller decides.
func (q *qpay_auth) Refresh(ctx context.Context, refreshToken string) (Token, error) {
	if refreshToken == "" {
		return Token{}, errors.New("qpay: refresh token is required")
	}

	body, status, err := doRequestCtx(ctx, q.client, QPayAuthRefresh.Method, q.endpoint+QPayAuthRefresh.Url, nil, map[string]string{
		"Content-Type":  utils.HttpContent,
		"Authorization": "Bearer " + refreshToken,
	}, nil)
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

	var authRes qpayLoginResponse
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
func (q *qpay_auth) SetToken(token Token) {
	q.mu.Lock()
	q.token = token
	q.mu.Unlock()
}

// Token returns the installed token.
func (q *qpay_auth) Token() Token {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.token
}
