package qpay_v2

import (
	"fmt"
	"net/http"
	"time"

	"github.com/techpartners-asia/qpay-go/utils"
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

	url := q.endpoint + api.Url + urlExt
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

	if res.IsError() {
		return fmt.Errorf("%s-QPay response error: %s (Status: %d)",
			time.Now().Format("2006-01-02 15:04:05"),
			res.String(),
			res.StatusCode())
	}

	return nil
}

// authQPayV2 [Internal: qPay-ээс Access Token авах/шинэчлэх]
// Энэ функц нь токен дуусах хугацааг шалгаж, шаардлагатай бол автоматаар шинэчилнэ.
// Access token хугацаа дуусвал refresh token ашиглан шинэчилнэ.
// Refresh token хугацаа дуусвал бүрэн дахин нэвтэрнэ.
// Uses singleflight: 50 concurrent calls = 1 network request, all 50 get the same result.
// See: https://developer.qpay.mn/#auth-token
func (q *qpay) authQPayV2() (authRes qpayLoginResponse, err error) {
	// Fast path: access token still valid
	q.mu.RLock()
	if q.loginObject != nil {
		expiryTime := time.Unix(int64(q.loginObject.ExpiresIn), 0)
		if time.Now().Before(expiryTime.Add(-1 * time.Minute)) {
			authRes = *q.loginObject
			q.mu.RUnlock()
			return authRes, nil
		}
	}
	q.mu.RUnlock()

	// Slow path: singleflight ensures only one auth call at a time.
	// All concurrent callers share the same result.
	v, err, _ := q.authGroup.Do("auth", func() (interface{}, error) {
		// Double-check: another goroutine may have refreshed while we waited
		canRefresh := false
		var refreshToken string
		q.mu.RLock()
		if q.loginObject != nil {
			expiryTime := time.Unix(int64(q.loginObject.ExpiresIn), 0)
			if time.Now().Before(expiryTime.Add(-1 * time.Minute)) {
				res := *q.loginObject
				q.mu.RUnlock()
				return res, nil
			}
			// Token expired — check if we can use refresh token
			if q.loginObject.RefreshToken != "" {
				refreshExpiry := time.Unix(int64(q.loginObject.RefreshExpiresIn), 0)
				if time.Now().Before(refreshExpiry.Add(-1 * time.Minute)) {
					canRefresh = true
					refreshToken = q.loginObject.RefreshToken
				}
			}
		}
		q.mu.RUnlock()

		var res qpayLoginResponse
		var authErr error
		if canRefresh {
			res, authErr = q.refreshAccessToken(refreshToken)
		} else {
			res, authErr = q.fullAuth()
		}
		if authErr != nil {
			return res, authErr
		}

		// Update shared state
		q.mu.Lock()
		q.loginObject = &res
		q.loginTime = time.Now()
		q.mu.Unlock()

		return res, nil
	})
	if err != nil {
		return authRes, err
	}

	return v.(qpayLoginResponse), nil
}

// refreshAccessToken [Internal: Refresh token ашиглан access token шинэчлэх]
func (q *qpay) refreshAccessToken(refreshToken string) (qpayLoginResponse, error) {
	var authRes qpayLoginResponse
	url := q.endpoint + QPayAuthRefresh.Url
	res, err := q.client.R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(refreshToken).
		SetResult(&authRes).
		Post(url)

	if err != nil {
		return authRes, err
	}

	if res.IsError() {
		return authRes, fmt.Errorf("%s-QPay refresh failed: %s (Status: %d)",
			time.Now().Format("2006-01-02 15:04:05"),
			res.String(),
			res.StatusCode())
	}

	return authRes, nil
}

// fullAuth [Internal: Username/password ашиглан бүрэн нэвтрэх]
func (q *qpay) fullAuth() (qpayLoginResponse, error) {
	var authRes qpayLoginResponse
	url := q.endpoint + QPayAuthToken.Url
	res, err := q.client.R().
		SetHeader("Content-Type", "application/json").
		SetBasicAuth(q.username, q.password).
		SetResult(&authRes).
		Post(url)

	if err != nil {
		return authRes, err
	}

	if res.IsError() {
		return authRes, fmt.Errorf("%s-QPay auth failed: %s (Status: %d)",
			time.Now().Format("2006-01-02 15:04:05"),
			res.String(),
			res.StatusCode())
	}

	return authRes, nil
}
