package qpay_v2

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/techpartners-asia/qpay-go/utils"
	"resty.dev/v3"
)

var (
	// QPayAuthToken [Access Token авах]
	QPayAuthToken = utils.API{
		Url:    "/auth/token",
		Method: http.MethodPost,
	}
	// QPayAuthRefresh [Access Token шинэчлэх]
	QPayAuthRefresh = utils.API{
		Url:    "/auth/refresh",
		Method: http.MethodPost,
	}
	// QPayInvoiceCreate [Нэхэмжлэх үүсгэх]
	QPayInvoiceCreate = utils.API{
		Url:    "/invoice",
		Method: http.MethodPost,
	}
	// QPayInvoiceGet [Нэхэмжлэх харах]
	QPayInvoiceGet = utils.API{
		Url:    "/invoice/",
		Method: http.MethodGet,
	}
	// QPayInvoiceCancel [Нэхэмжлэх цуцлах]
	QPayInvoiceCancel = utils.API{
		Url:    "/invoice/",
		Method: http.MethodDelete,
	}
	// QPayPaymentGet [Төлбөр харах]
	QPayPaymentGet = utils.API{
		Url:    "/payment/",
		Method: http.MethodGet,
	}
	// QPayPaymentCheck [Төлбөр шалгах]
	QPayPaymentCheck = utils.API{
		Url:    "/payment/check",
		Method: http.MethodPost,
	}
	// QPayPaymentCancel [Төлбөр цуцлах]
	QPayPaymentCancel = utils.API{
		Url:    "/payment/cancel/",
		Method: http.MethodDelete,
	}
	// QPayPaymentRefund [Төлбөр буцаах]
	QPayPaymentRefund = utils.API{
		Url:    "/payment/refund/",
		Method: http.MethodDelete,
	}
	// QPayPaymentList [Төлбөрийн жагсаалт]
	QPayPaymentList = utils.API{
		Url:    "/payment/list",
		Method: http.MethodPost,
	}
	// QPayEbarimtCreate [И-баримт үүсгэх]
	QPayEbarimtCreate = utils.API{
		Url:    "/ebarimt_v3/create",
		Method: http.MethodPost,
	}
	// QPayEbarimtCancel [И-баримт цуцлах]
	QPayEbarimtCancel = utils.API{
		Url:    "/ebarimt_v3/",
		Method: http.MethodDelete,
	}
)

// httpRequestQPay [Internal: QPay API-руу HTTP хүсэлт илгээх туслах функц]
// body: Хүсэлтийн бие (POST/PUT үед)
// result: Хариуг задлах бүтэц (struct pointer)
// api: utils.API төрлийн эндпоинт тохиргоо
// urlExt: URL-д залгагдах нэмэлт ID (invoice_id, payment_id г.м)
func (q *qpay) httpRequestQPay(body interface{}, result interface{}, api utils.API, urlExt string) error {

	_, authErr := q.authQPayV2()
	if authErr != nil {
		return authErr
	}

	// Ensure thread safety for token fetch
	q.mu.RLock()
	token := ""
	if q.loginObject != nil {
		token = q.loginObject.AccessToken
	}
	q.mu.RUnlock()

	url := q.endpoint + api.Url + utils.EscapePathSegment(urlExt)
	req := q.client.R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(token).
		SetResult(result)

	// Standard guard: avoid sending identity bodies on non-mutation requests
	if body != nil {
		req.SetBody(body)
	}

	res, err := req.Execute(api.Method, url)
	if err != nil {
		return err
	}
	defer closeBody(res)

	// Anything outside 2xx is an error. Checking only for >= 400 let 3xx
	// responses through with `result` left at its zero value, which reads
	// downstream as a successful-but-empty payment.
	if !res.IsStatusSuccess() {
		return fmt.Errorf("%s-QPay response error: %s (Status: %d)",
			time.Now().Format("2006-01-02 15:04:05"),
			utils.TruncateForError(res.String()),
			res.StatusCode())
	}

	return nil
}

// closeBody releases a response body. Resty drains and closes it while
// decoding 2xx and >=400 responses, but not for 204 or 3xx, and the body's
// Close is what cancels the per-request timeout context.
func closeBody(res *resty.Response) {
	if res != nil && res.Body != nil {
		_ = res.Body.Close()
	}
}

// authQPayV2 [Internal: qPay-ээс Access Token авах/шинэчлэх]
// Simple: check token → if valid return cached → if expired, one goroutine auths via singleflight.
// See: https://developer.qpay.mn/#auth-token
func (q *qpay) authQPayV2() (qpayLoginResponse, error) {
	// Fast path: token still valid
	q.mu.RLock()
	if q.loginObject != nil && q.tokenValid() {
		res := *q.loginObject
		q.mu.RUnlock()
		return res, nil
	}
	q.mu.RUnlock()

	// Slow path: singleflight deduplicates concurrent auth calls
	v, err, _ := q.authGroup.Do("auth", func() (any, error) {
		// Double-check after waiting
		q.mu.RLock()
		if q.loginObject != nil && q.tokenValid() {
			res := *q.loginObject
			q.mu.RUnlock()
			return res, nil
		}

		// Try refresh token first, fallback to full auth
		canRefresh := q.loginObject != nil && q.loginObject.RefreshToken != "" && q.refreshTokenValid()
		var refreshToken string
		if canRefresh {
			refreshToken = q.loginObject.RefreshToken
		}
		q.mu.RUnlock()

		var res qpayLoginResponse
		var authErr error
		if canRefresh {
			res, authErr = q.doRefresh(refreshToken)
			if authErr != nil {
				// Refresh failed — fallback to full auth
				res, authErr = q.doAuth()
			}
		} else {
			res, authErr = q.doAuth()
		}
		if authErr != nil {
			return res, authErr
		}

		q.mu.Lock()
		q.loginObject = &res
		q.mu.Unlock()
		return res, nil
	})
	if err != nil {
		return qpayLoginResponse{}, err
	}
	return v.(qpayLoginResponse), nil
}

// tokenValid checks if access token is still valid (must hold mu.RLock)
func (q *qpay) tokenValid() bool {
	return tokenStillValid(q.loginObject.ExpiresIn)
}

// refreshTokenValid checks if refresh token is still valid (must hold mu.RLock)
func (q *qpay) refreshTokenValid() bool {
	return tokenStillValid(q.loginObject.RefreshExpiresIn)
}

// tokenStillValid reports whether a qPay expiry field is comfortably in the
// future. qPay documents expires_in as a Unix timestamp; a value too small to
// be one is a relative duration whose issue time we do not know, so it is
// never treated as cacheable.
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

// doAuth [Full auth: username/password]
func (q *qpay) doAuth() (qpayLoginResponse, error) {
	var authRes qpayLoginResponse
	res, err := q.client.R().
		SetHeader("Content-Type", "application/json").
		SetBasicAuth(q.username, q.password).
		SetResult(&authRes).
		Post(q.endpoint + QPayAuthToken.Url)
	if err != nil {
		return authRes, err
	}
	defer closeBody(res)
	if !res.IsStatusSuccess() {
		return authRes, fmt.Errorf("%s-QPay auth failed: %s (Status: %d)",
			time.Now().Format("2006-01-02 15:04:05"), utils.TruncateForError(res.String()), res.StatusCode())
	}
	if authRes.AccessToken == "" {
		// Caching a tokenless response would send every later request with an
		// empty bearer and re-authenticate on each one.
		return qpayLoginResponse{}, errors.New("qpay: auth response contained no access token")
	}
	return authRes, nil
}

// doRefresh [Refresh token ашиглан access token шинэчлэх]
func (q *qpay) doRefresh(refreshToken string) (qpayLoginResponse, error) {
	var authRes qpayLoginResponse
	res, err := q.client.R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(refreshToken).
		SetResult(&authRes).
		Post(q.endpoint + QPayAuthRefresh.Url)
	if err != nil {
		return authRes, err
	}
	defer closeBody(res)
	if !res.IsStatusSuccess() {
		return authRes, fmt.Errorf("%s-QPay refresh failed: %s (Status: %d)",
			time.Now().Format("2006-01-02 15:04:05"), utils.TruncateForError(res.String()), res.StatusCode())
	}
	if authRes.AccessToken == "" {
		return qpayLoginResponse{}, errors.New("qpay: refresh response contained no access token")
	}
	return authRes, nil
}
