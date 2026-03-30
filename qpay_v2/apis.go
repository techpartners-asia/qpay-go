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
// See: https://developer.qpay.mn/#auth-token
func (q *qpay) authQPayV2() (authRes qpayLoginResponse, err error) {
	// 1. Fast path: Read-lock check — access token still valid
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

	// 2. Slow path: Acquire refresh lock (serializes the network call)
	q.refreshMu.Lock()
	defer q.refreshMu.Unlock()

	// 3. Double-check after acquiring refreshMu
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

	// 4. Determine whether to use refresh token or full auth
	q.mu.RLock()
	canRefresh := false
	var refreshToken string
	if q.loginObject != nil && q.loginObject.RefreshToken != "" {
		refreshExpiry := time.Unix(int64(q.loginObject.RefreshExpiresIn), 0)
		if time.Now().Before(refreshExpiry.Add(-1 * time.Minute)) {
			canRefresh = true
			refreshToken = q.loginObject.RefreshToken
		}
	}
	q.mu.RUnlock()

	if canRefresh {
		// 4a. Use refresh token — lightweight, no username/password needed
		authRes, err = q.refreshAccessToken(refreshToken)
	} else {
		// 4b. Full auth — first login or refresh token expired
		authRes, err = q.fullAuth()
	}
	if err != nil {
		return authRes, err
	}

	// 5. Update shared state under Write-lock
	q.mu.Lock()
	q.loginObject = &authRes
	q.loginTime = time.Now()
	q.mu.Unlock()

	return authRes, nil
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
