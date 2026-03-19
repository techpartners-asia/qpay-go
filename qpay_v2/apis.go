package qpay_v2

import (
	"fmt"
	"time"

	"github.com/techpartners-asia/qpay-go/utils"
)

var (
	// QPayAuthToken [Access Token авах]
	QPayAuthToken = utils.API{
		Url:    "/auth/token",
		Method: "POST",
	}
	// QPayInvoiceCreate [Нэхэмжлэх үүсгэх]
	QPayInvoiceCreate = utils.API{
		Url:    "/invoice",
		Method: "POST",
	}
	// QPayInvoiceGet [Нэхэмжлэх харах]
	QPayInvoiceGet = utils.API{
		Url:    "/invoice/",
		Method: "GET",
	}
	// QPayInvoiceCancel [Нэхэмжлэх цуцлах]
	QPayInvoiceCancel = utils.API{
		Url:    "/invoice/",
		Method: "DELETE",
	}
	// QPayPaymentGet [Төлбөр харах]
	QPayPaymentGet = utils.API{
		Url:    "/payment/",
		Method: "GET",
	}
	// QPayPaymentCheck [Төлбөр шалгах]
	QPayPaymentCheck = utils.API{
		Url:    "/payment/check",
		Method: "POST",
	}
	// QPayPaymentCancel [Төлбөр цуцлах]
	QPayPaymentCancel = utils.API{
		Url:    "/payment/cancel/",
		Method: "DELETE",
	}
	// QPayPaymentRefund [Төлбөр буцаах]
	QPayPaymentRefund = utils.API{
		Url:    "/payment/refund/",
		Method: "DELETE",
	}
	// QPayPaymentList [Төлбөрийн жагсаалт]
	QPayPaymentList = utils.API{
		Url:    "/payment/list",
		Method: "POST",
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
	res, err := q.client.R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(token).
		SetBody(body).
		SetResult(result).
		Execute(api.Method, url)

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
	q.mu.RLock()
	// Check if existing token is still valid (with 1 minute buffer)
	if q.loginObject != nil {
		expiryTime := q.loginTime.Add(time.Duration(q.loginObject.ExpiresIn) * time.Second)
		if time.Now().Before(expiryTime.Add(-1 * time.Minute)) {
			authRes = *q.loginObject
			q.mu.RUnlock()
			return authRes, nil
		}
	}
	q.mu.RUnlock()

	// Perform auth with lock to prevent multiple concurrent login requests
	q.mu.Lock()
	defer q.mu.Unlock()

	// Double check after acquiring lock (Double-Checked Locking pattern)
	if q.loginObject != nil {
		expiryTime := q.loginTime.Add(time.Duration(q.loginObject.ExpiresIn) * time.Second)
		if time.Now().Before(expiryTime.Add(-1 * time.Minute)) {
			return *q.loginObject, nil
		}
	}

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

	// Persist the new token and the time it was received
	q.loginObject = &authRes
	q.loginTime = time.Now()

	return authRes, nil
}
