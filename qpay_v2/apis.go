package qpay_v2

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

// func (q *qpay) ExpireTokenForce() {
// 	q.loginObject.ExpiresIn = 0
// }

func (q *qpay) httpRequestQPay(body interface{}, api utils.API, urlExt string) (response []byte, err error) {

	authObj, authErr := q.authQPayV2()
	if authErr != nil {
		err = authErr
		return
	}
	q.loginObject = &authObj

	var requestByte []byte
	var requestBody *bytes.Reader
	if body == nil {
		requestBody = bytes.NewReader(nil)
	} else {
		requestByte, _ = json.Marshal(body)
		requestBody = bytes.NewReader(requestByte)
	}

	req, _ := http.NewRequest(api.Method, q.endpoint+api.Url+urlExt, requestBody)
	req.Header.Add("Content-Type", utils.HttpContent)
	req.Header.Add("Authorization", "Bearer "+q.loginObject.AccessToken)

	res, err := http.DefaultClient.Do(req)

	response, _ = io.ReadAll(res.Body)
	if res.StatusCode != 200 {
		return nil, errors.New(string(response))
	}

	return
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
	req, err := http.NewRequest(QPayAuthToken.Method, url, nil)
	if err != nil {
		fmt.Println(err.Error())
	}
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(q.username, q.password)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	if res.StatusCode != http.StatusOK {
		return authRes, fmt.Errorf("%s-QPay auth response: %s", time.Now().Format(utils.TimeFormatYYYYMMDDHHMMSS), res.Status)
	}

	body, _ := io.ReadAll(res.Body)
	json.Unmarshal(body, &authRes)

	defer res.Body.Close()
	return authRes, nil
}

func authQPayV2(username, password, endpoint, callback, invoiceCode, merchantId string) (authRes qpayLoginResponse, err error) {
	url := endpoint + QPayAuthToken.Url
	req, err := http.NewRequest(QPayAuthToken.Method, url, nil)
	if err != nil {
		fmt.Println(err.Error())
	}
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(username, password)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	if res.StatusCode != http.StatusOK {
		return authRes, fmt.Errorf("%s-QPay auth response: %s", time.Now().Format(utils.TimeFormatYYYYMMDDHHMMSS), res.Status)
	}

	body, _ := io.ReadAll(res.Body)
	json.Unmarshal(body, &authRes)

	defer res.Body.Close()
	return authRes, nil
}

func (q *qpay) refreshToken() (authRes qpayLoginResponse, err error) {
	url := q.endpoint + QPayAuthRefresh.Url
	req, err := http.NewRequest(QPayAuthRefresh.Method, url, nil)
	if err != nil {
		fmt.Println(err.Error())
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+q.loginObject.RefreshToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	if res.StatusCode != 200 {
		return authRes, errors.New(time.Now().Format(utils.TimeFormatYYYYMMDDHHMMSS) + "-QPay token refresh response: " + res.Status)
	}

	body, _ := io.ReadAll(res.Body)
	json.Unmarshal(body, &authRes)

	defer res.Body.Close()
	return
}
