package qpay_v1

import (
	"bytes"
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
	var reader io.Reader
	if payload != nil {
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, url, reader)
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
	authObj, err := q.authQPayV1()
	if err != nil {
		return nil, err
	}

	q.mu.Lock()
	q.loginObject = &authObj
	q.mu.Unlock()

	var payload []byte
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("qpay: encode request: %w", err)
		}
	}

	url := q.endpoint + api.Url + utils.EscapePathSegment(urlExt)
	response, status, err := q.doRequest(api.Method, url, payload, map[string]string{
		"Content-Type":  utils.HttpContent,
		"Authorization": "Bearer " + authObj.AccessToken,
	})
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("%s-QPay response error (Status: %d): %s",
			time.Now().Format(utils.TimeFormatYYYYMMDDHHMMSS), status, utils.TruncateForError(string(response)))
	}

	return response, nil
}

// authQPayV1 [Login to qpay]
func (q *qpay) authQPayV1() (QpayLoginResponse, error) {
	q.mu.RLock()
	if q.loginObject != nil && tokenStillValid(int64(q.loginObject.ExpiresIn)) {
		cached := *q.loginObject
		q.mu.RUnlock()
		return cached, nil
	}
	q.mu.RUnlock()

	var authRes QpayLoginResponse

	payload, err := json.Marshal(&QpayLogin{
		ClientID:     q.client_id,
		ClientSecret: q.client_secret,
		GrantType:    q.grant_type,
		RefreshToken: q.refresh_token,
	})
	if err != nil {
		return authRes, fmt.Errorf("qpay: encode auth request: %w", err)
	}

	body, status, err := q.doRequest(QPayAuthToken.Method, q.endpoint+QPayAuthToken.Url, payload, map[string]string{
		"Content-Type": utils.HttpContent,
	})
	if err != nil {
		return authRes, err
	}

	if status != http.StatusOK {
		return authRes, fmt.Errorf("%s-QPay auth response (Status: %d)",
			time.Now().Format(utils.TimeFormatYYYYMMDDHHMMSS), status)
	}

	if err := json.Unmarshal(body, &authRes); err != nil {
		return QpayLoginResponse{}, fmt.Errorf("qpay: decode auth response: %w", err)
	}
	if authRes.AccessToken == "" {
		// A 200 with no token would otherwise be cached and every later call
		// would silently go out unauthenticated.
		return QpayLoginResponse{}, errors.New("qpay: auth response contained no access token")
	}

	return authRes, nil
}

func (q *qpay) refreshToken() (QpayLoginResponse, error) {
	var authRes QpayLoginResponse

	q.mu.RLock()
	if q.loginObject == nil {
		q.mu.RUnlock()
		return authRes, errors.New("qpay: no session to refresh")
	}
	refresh := q.loginObject.RefreshToken
	q.mu.RUnlock()

	if refresh == "" {
		return authRes, errors.New("qpay: no refresh token available")
	}

	body, status, err := q.doRequest(QPayAuthRefresh.Method, q.endpoint+QPayAuthRefresh.Url, nil, map[string]string{
		"Content-Type":  utils.HttpContent,
		"Authorization": "Bearer " + refresh,
	})
	if err != nil {
		return authRes, err
	}

	if status != http.StatusOK {
		return authRes, fmt.Errorf("%s-QPay token refresh response (Status: %d)",
			time.Now().Format(utils.TimeFormatYYYYMMDDHHMMSS), status)
	}

	if err := json.Unmarshal(body, &authRes); err != nil {
		return QpayLoginResponse{}, fmt.Errorf("qpay: decode refresh response: %w", err)
	}

	return authRes, nil
}

// tokenStillValid reports whether a qPay expiry field is comfortably in the
// future. qPay documents expires_in as a Unix timestamp, but sends a plain
// duration in seconds on some deployments; values too small to be a
// timestamp are therefore treated as a duration from now.
func tokenStillValid(expiresIn int64) bool {
	if expiresIn <= 0 {
		return false
	}

	const minPlausibleTimestamp = 1_000_000_000 // 2001-09-09
	if expiresIn < minPlausibleTimestamp {
		// Duration semantics: we cannot tell when it was issued, so do not cache.
		return false
	}

	return time.Now().Before(time.Unix(expiresIn, 0).Add(-1 * time.Minute))
}
