package qpay_wechat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/techpartners-asia/qpay-go/utils"
)

type qpay_auth struct {
	endpoint    string
	password    string
	username    string
	callback    string
	invoiceCode string
	merchantId  string

	mu          sync.RWMutex
	loginObject *qpayLoginResponse
	client      *http.Client
}

// Wechat not supported

type QPayAuth interface {
	// CreateInvoice(input QPayCreateInvoiceInput) (QPaySimpleInvoiceResponse, QPayAuth, error)
	// GetInvoice(invoiceId string) (QpayInvoiceGetResponse, QPayAuth, error)
	// CancelInvoice(invoiceId string) (interface{}, QPayAuth, error)
	// GetPayment(invoiceId string) (interface{}, QPayAuth, error)
	// CheckPayment(invoiceId string, pageLimit, pageNumber int64) (QpayPaymentCheckResponse, QPayAuth, error)
	// CancelPayment(invoiceId, paymentUUID string) (QpayPaymentCheckResponse, QPayAuth, error)
	// RefundPayment(invoiceId, paymentUUID string) (interface{}, QPayAuth, error)
	// GetPaymentList()
}

func New(username, password, endpoint, callback, invoiceCode, merchantId string) QPayAuth {
	q := &qpay_auth{
		endpoint:    endpoint,
		password:    password,
		username:    username,
		callback:    callback,
		invoiceCode: invoiceCode,
		merchantId:  merchantId,
		client:      utils.NewHTTPClient(),
	}

	// Warm the token cache. A failure here is not fatal: authQPayV2 retries
	// on the first API call. Storing an empty login object on failure (as this
	// used to) would have made every later request go out unauthenticated.
	if authObj, err := authQPayV2(q.client, username, password, endpoint); err == nil {
		q.loginObject = &authObj
	}

	return q
}

func (q *qpay_auth) CreateInvoice(input QPayCreateInvoiceInput) (QPaySimpleInvoiceResponse, QPayAuth, error) {
	amountInt := int64(input.Amount)
	request := QPaySimpleInvoiceRequest{
		InvoiceCode:         q.invoiceCode,
		SenderInvoiceCode:   input.SenderInvoiceNo,
		SenderBranchCode:    input.SenderBranchCode,
		InvoiceReceiverCode: input.InvoiceReceiverCode,
		InvoiceDescription:  input.InvoiceDescription,
		Amount:              amountInt,
		CallbackUrl:         utils.BuildCallbackURL(q.callback, input.CallbackParam),
	}

	res, err := q.httpRequest(request, QPayInvoiceCreate, "")
	if err != nil {
		return QPaySimpleInvoiceResponse{}, q, err
	}

	var response QPaySimpleInvoiceResponse
	if err := json.Unmarshal(res, &response); err != nil {
		return QPaySimpleInvoiceResponse{}, q, fmt.Errorf("qpay: decode invoice create response: %w", err)
	}

	return response, q, nil
}
func (q *qpay_auth) GetInvoice(invoiceId string) (QpayInvoiceGetResponse, QPayAuth, error) {
	res, err := q.httpRequest(nil, QPayInvoiceGet, invoiceId)
	if err != nil {
		return QpayInvoiceGetResponse{}, q, err
	}

	var response QpayInvoiceGetResponse
	if err := json.Unmarshal(res, &response); err != nil {
		return QpayInvoiceGetResponse{}, q, fmt.Errorf("qpay: decode invoice get response: %w", err)
	}

	return response, q, nil
}
func (q *qpay_auth) CancelInvoice(invoiceId string) (interface{}, QPayAuth, error) {
	res, err := q.httpRequest(nil, QPayInvoiceCancel, invoiceId)
	if err != nil {
		return nil, q, err
	}

	var response interface{}
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, q, fmt.Errorf("qpay: decode response: %w", err)
	}

	return response, q, nil
}

func (q *qpay_auth) GetPayment(invoiceId string) (interface{}, QPayAuth, error) {
	res, err := q.httpRequest(nil, QPayPaymentGet, invoiceId)
	if err != nil {
		return nil, q, err
	}

	var response interface{}
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, q, fmt.Errorf("qpay: decode response: %w", err)
	}

	return response, q, nil
}

func (q *qpay_auth) CheckPayment(invoiceId string, pageLimit, pageNumber int64) (QpayPaymentCheckResponse, QPayAuth, error) {
	req := QpayPaymentCheckRequest{}
	req.ObjectID = invoiceId
	req.ObjectType = "INVOICE"
	req.Offset.PageLimit = pageLimit
	req.Offset.PageNumber = pageNumber

	var response QpayPaymentCheckResponse

	res, err := q.httpRequest(req, QPayPaymentCheck, "")
	if err != nil {
		return response, q, err
	}

	if err := json.Unmarshal(res, &response); err != nil {
		return QpayPaymentCheckResponse{}, q, fmt.Errorf("qpay: decode payment check response: %w", err)
	}

	return response, q, nil
}

func (q *qpay_auth) CancelPayment(invoiceId, paymentUUID string) (QpayPaymentCheckResponse, QPayAuth, error) {
	var req QpayPaymentCancelRequest

	req.CallbackUrl = q.callback + paymentUUID
	req.Note = "Cancel payment - " + invoiceId

	var response QpayPaymentCheckResponse

	res, err := q.httpRequest(req, QPayPaymentCancel, invoiceId)
	// ret := func() QPayAuth {
	// 	return &qpay_auth{
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

	if err := json.Unmarshal(res, &response); err != nil {
		return QpayPaymentCheckResponse{}, q, fmt.Errorf("qpay: decode payment cancel response: %w", err)
	}

	return response, q, nil
}

func (q *qpay_auth) RefundPayment(invoiceId, paymentUUID string) (interface{}, QPayAuth, error) {
	var req QpayPaymentCancelRequest

	req.CallbackUrl = q.callback + paymentUUID
	req.Note = "Cancel payment - " + invoiceId

	var response interface{}

	res, err := q.httpRequest(req, QPayPaymentRefund, invoiceId)
	if err != nil {
		return response, q, err
	}

	if err := json.Unmarshal(res, &response); err != nil {
		return nil, q, fmt.Errorf("qpay: decode payment refund response: %w", err)
	}

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
