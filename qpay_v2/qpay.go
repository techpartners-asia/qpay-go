package qpay_v2

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type qpay struct {
	endpoint    string
	password    string
	username    string
	callback    string
	invoiceCode string
	merchantId  string
	loginObject *qpayLoginResponse
}

type QPay interface {
	CreateInvoice(input QPayCreateInvoiceInput) (QPaySimpleInvoiceResponse, QPay, error)
	GetInvoice(invoiceId string) (QpayInvoiceGetResponse, QPay, error)
	CancelInvoice(invoiceId string) (interface{}, QPay, error)
	GetPayment(invoiceId string) (interface{}, QPay, error)
	CheckPayment(invoiceId string, pageLimit, pageNumber int64) (QpayPaymentCheckResponse, QPay, error)
	CancelPayment(invoiceId, paymentUUID string) (QpayPaymentCheckResponse, QPay, error)
	RefundPayment(invoiceId, paymentUUID string) (interface{}, QPay, error)
	// GetPaymentList()
}

func New(username, password, endpoint, callback, invoiceCode, merchantId string) QPay {
	return &qpay{
		endpoint:    endpoint,
		password:    password,
		username:    username,
		callback:    callback,
		invoiceCode: invoiceCode,
		merchantId:  merchantId,
		loginObject: func() *qpayLoginResponse {
			authObj, authErr := authQPayV2(username, password, endpoint, callback, invoiceCode, merchantId)
			if authErr != nil {
				// err = authErr
				return &qpayLoginResponse{}
			}
			return &authObj
		}(),
	}
}

func (q *qpay) CreateInvoice(input QPayCreateInvoiceInput) (QPaySimpleInvoiceResponse, QPay, error) {
	vals := url.Values{}
	for k, v := range input.CallbackParam {
		vals.Add(k, v)
	}

	amountInt := int64(input.Amount)
	request := QPaySimpleInvoiceRequest{
		InvoiceCode:         q.invoiceCode,
		SenderInvoiceCode:   input.SenderCode,
		SenderBranchCode:    input.SenderBranchCode,
		InvoiceReceiverCode: input.ReceiverCode,
		InvoiceDescription:  input.Description,
		Amount:              amountInt,
		CallbackUrl:         fmt.Sprintf("%s?%s", q.callback, vals.Encode()),
	}

	res, err := q.httpRequestQPay(request, QPayInvoiceCreate, "")
	if err != nil {
		return QPaySimpleInvoiceResponse{}, q, err
	}

	var response QPaySimpleInvoiceResponse
	json.Unmarshal(res, &response)

	return response, q, nil
}
func (q *qpay) GetInvoice(invoiceId string) (QpayInvoiceGetResponse, QPay, error) {
	res, err := q.httpRequestQPay(nil, QPayInvoiceGet, invoiceId)
	if err != nil {
		return QpayInvoiceGetResponse{}, q, err
	}

	var response QpayInvoiceGetResponse
	json.Unmarshal(res, &response)

	return response, q, nil
}
func (q *qpay) CancelInvoice(invoiceId string) (interface{}, QPay, error) {
	res, err := q.httpRequestQPay(nil, QPayInvoiceCancel, invoiceId)
	if err != nil {
		return nil, q, err
	}

	var response interface{}
	json.Unmarshal(res, &response)

	return response, q, nil
}

func (q *qpay) GetPayment(invoiceId string) (interface{}, QPay, error) {
	res, err := q.httpRequestQPay(nil, QPayPaymentGet, invoiceId)
	if err != nil {
		return nil, q, err
	}

	var response interface{}
	json.Unmarshal(res, &response)

	return response, q, nil
}

func (q *qpay) CheckPayment(invoiceId string, pageLimit, pageNumber int64) (QpayPaymentCheckResponse, QPay, error) {
	req := QpayPaymentCheckRequest{}
	req.ObjectID = invoiceId
	req.ObjectType = "INVOICE"
	req.Offset.PageLimit = pageLimit
	req.Offset.PageNumber = pageNumber

	var response QpayPaymentCheckResponse

	res, err := q.httpRequestQPay(req, QPayPaymentCheck, "")
	if err != nil {
		return response, q, err
	}

	json.Unmarshal(res, &response)

	return response, q, nil
}

func (q *qpay) CancelPayment(invoiceId, paymentUUID string) (QpayPaymentCheckResponse, QPay, error) {
	var req QpayPaymentCancelRequest

	req.CallbackUrl = q.callback + paymentUUID
	req.Note = "Cancel payment - " + invoiceId

	var response QpayPaymentCheckResponse

	res, err := q.httpRequestQPay(req, QPayPaymentCancel, invoiceId)
	// ret := func() QPay {
	// 	return &qpay{
	// 		endpoint:    q.endpoint,
	// 		password:    q.password,
	// 		username:    q.username,
	// 		callback:    q.callback,
	// 		invoiceCode: q.invoiceCode,
	// 		merchantId:  q.merchantId,
	// 		loginObject: q.loginObject,
	// 	}
	// }()
	if err != nil {
		return response, q, err
	}

	json.Unmarshal(res, &response)

	return response, q, nil
}

func (q *qpay) RefundPayment(invoiceId, paymentUUID string) (interface{}, QPay, error) {
	var req QpayPaymentCancelRequest

	req.CallbackUrl = q.callback + paymentUUID
	req.Note = "Cancel payment - " + invoiceId

	var response interface{}

	res, err := q.httpRequestQPay(req, QPayPaymentRefund, invoiceId)
	if err != nil {
		return response, q, err
	}

	json.Unmarshal(res, &response)

	return response, q, nil
}

// func (q *qpay) GetPaymentList() (QpayPaymentListRequest, error) {
// 	var req QpayPaymentListRequest
// 	req.MerchantID = q.merchantId

// 	res, err := utils.HttpRequestQpay(list, helper.QPayPaymentList, "")
// 	if err != nil {
// 		return res, err
// 	}
// }
