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
	QPayAuthToken = utils.API{
		Url:    "/auth/token",
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

// func (q *qpay) ExpireTokenForce() {
// 	q.loginObject.ExpiresIn = 0
// }

func (q *qpay) httpRequestQPay(body interface{}, result interface{}, api utils.API, urlExt string) error {

	authObj, authErr := q.authQPayV2()
	if authErr != nil {
		return authErr
	}
	q.loginObject = &authObj

	req := q.client.R().
		SetHeader("Content-Type", utils.HttpContent).
		SetAuthToken(q.loginObject.AccessToken).
		SetResult(result)

	if body != nil {
		req.SetBody(body)
	}

	res, err := req.Execute(api.Method, q.endpoint+api.Url+urlExt)
	if err != nil {
		return err
	}
	
	if res.IsError() {
		return errors.New(res.String())
	}

	return nil
}

// AuthQPayV2 [Login to qpay]
func (q *qpay) authQPayV2() (authRes qpayLoginResponse, err error) {
	// check loginOnject is valid
	if q.loginObject != nil {
		expireInA := time.Unix(int64(q.loginObject.ExpiresIn), 0)
		expireInB := expireInA.Add(time.Duration(-12) * time.Hour)
		now := time.Now()
		if now.Before(expireInB) {
			authRes = *q.loginObject
			err = nil
			return
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
		return authRes, fmt.Errorf("%s-QPay auth response: %s", time.Now().Format(utils.TimeFormatYYYYMMDDHHMMSS), res.Status())
	}

	return authRes, nil
}

func authQPayV2(username, password, endpoint, callback, invoiceCode, merchantId string) (authRes qpayLoginResponse, err error) {
	client := resty.New()
	url := endpoint + QPayAuthToken.Url

	res, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBasicAuth(username, password).
		SetResult(&authRes).
		Post(url)

	if err != nil {
		return authRes, err
	}

	if res.IsError() {
		return authRes, fmt.Errorf("%s-QPay auth response: %s", time.Now().Format(utils.TimeFormatYYYYMMDDHHMMSS), res.Status())
	}

	return authRes, nil
}

