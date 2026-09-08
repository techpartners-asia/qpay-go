package qpay_wechat

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
	var reader io.Reader
	if payload != nil {
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, url, reader)
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
	authObj, err := q.authQPayV2()
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
	response, status, err := doRequest(q.client, api.Method, url, payload, map[string]string{
		"Content-Type":  utils.HttpContent,
		"Authorization": "Bearer " + authObj.AccessToken,
	}, nil)
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("%s-QPay response error (Status: %d): %s",
			time.Now().Format(utils.TimeFormatYYYYMMDDHHMMSS), status, utils.TruncateForError(string(response)))
	}

	return response, nil
}

// AuthQPayV2 [Login to qpay]
func (q *qpay_auth) authQPayV2() (qpayLoginResponse, error) {
	q.mu.RLock()
	if q.loginObject != nil && tokenStillValid(int64(q.loginObject.ExpiresIn)) {
		cached := *q.loginObject
		q.mu.RUnlock()
		return cached, nil
	}
	q.mu.RUnlock()

	return authQPayV2(q.client, q.username, q.password, q.endpoint)
}

func authQPayV2(client *http.Client, username, password, endpoint string) (qpayLoginResponse, error) {
	var authRes qpayLoginResponse

	body, status, err := doRequest(client, QPayAuthToken.Method, endpoint+QPayAuthToken.Url, nil, map[string]string{
		"Content-Type": utils.HttpContent,
	}, &basicAuth{username: username, password: password})
	if err != nil {
		return authRes, err
	}

	if status != http.StatusOK {
		return authRes, fmt.Errorf("%s-QPay auth response (Status: %d)",
			time.Now().Format(utils.TimeFormatYYYYMMDDHHMMSS), status)
	}

	if err := json.Unmarshal(body, &authRes); err != nil {
		return qpayLoginResponse{}, fmt.Errorf("qpay: decode auth response: %w", err)
	}
	if authRes.AccessToken == "" {
		return qpayLoginResponse{}, errors.New("qpay: auth response contained no access token")
	}

	return authRes, nil
}

func (q *qpay_auth) refreshToken() (qpayLoginResponse, error) {
	var authRes qpayLoginResponse

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

	body, status, err := doRequest(q.client, QPayAuthRefresh.Method, q.endpoint+QPayAuthRefresh.Url, nil, map[string]string{
		"Content-Type":  utils.HttpContent,
		"Authorization": "Bearer " + refresh,
	}, nil)
	if err != nil {
		return authRes, err
	}

	if status != http.StatusOK {
		return authRes, fmt.Errorf("%s-QPay token refresh response (Status: %d)",
			time.Now().Format(utils.TimeFormatYYYYMMDDHHMMSS), status)
	}

	if err := json.Unmarshal(body, &authRes); err != nil {
		return qpayLoginResponse{}, fmt.Errorf("qpay: decode refresh response: %w", err)
	}

	return authRes, nil
}

// tokenStillValid reports whether a qPay expiry field is comfortably in the
// future. qPay documents expires_in as a Unix timestamp, but sends a plain
// duration in seconds on some deployments; values too small to be a
// timestamp are therefore treated as a duration and never cached.
func tokenStillValid(expiresIn int64) bool {
	if expiresIn <= 0 {
		return false
	}

	const minPlausibleTimestamp = 1_000_000_000 // 2001-09-09
	if expiresIn < minPlausibleTimestamp {
		return false
	}

	return time.Now().Before(time.Unix(expiresIn, 0).Add(-1 * time.Minute))
}
