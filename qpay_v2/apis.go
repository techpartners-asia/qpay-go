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
// See: https://developer.qpay.mn/#auth-token
func (q *qpay) authQPayV2() (authRes qpayLoginResponse, err error) {
	// 1. Fast path: Read-lock check
	q.mu.RLock()
	if q.loginObject != nil {
		expiryTime := q.loginTime.Add(time.Duration(q.loginObject.ExpiresIn) * time.Second)
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

	// 3. Double-check token state with Read-lock after acquiring refreshMu
	// (Another goroutine might have refreshed it while we were waiting on refreshMu)
	q.mu.RLock()
	if q.loginObject != nil {
		expiryTime := q.loginTime.Add(time.Duration(q.loginObject.ExpiresIn) * time.Second)
		if time.Now().Before(expiryTime.Add(-1 * time.Minute)) {
			authRes = *q.loginObject
			q.mu.RUnlock()
			return authRes, nil
		}
	}
	q.mu.RUnlock()

	// 4. Perform the actual network refresh (outside the main 'mu' to keep it responsive)
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

	// 5. Update shared state under Write-lock
	q.mu.Lock()
	q.loginObject = &authRes
	q.loginTime = time.Now()
	q.mu.Unlock()

	return authRes, nil
}
