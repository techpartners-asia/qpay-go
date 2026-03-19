package qpay_v2

import (
	"fmt"
	"net/url"

	"resty.dev/v3"
)

type qpay struct {
	endpoint    string
	password    string
	username    string
	callback    string
	invoiceCode string
	merchantId  string
	loginObject *qpayLoginResponse
	client      *resty.Client
}

type QPay interface {
	CreateInvoice(input QPayCreateInvoiceInput) (QPaySimpleInvoiceResponse, QPay, error)
	GetInvoice(invoiceId string) (QpayInvoiceGetResponse, QPay, error)
	CancelInvoice(invoiceId string) (interface{}, QPay, error)
	GetPayment(invoiceId string) (interface{}, QPay, error)
	CheckPayment(invoiceId string, pageLimit, pageNumber int64) (QpayPaymentCheckResponse, QPay, error)
	CancelPayment(invoiceId, paymentUUID string) (QpayPaymentCheckResponse, QPay, error)
	RefundPayment(invoiceId, paymentUUID string) (interface{}, QPay, error)
	SetClient(client *resty.Client)
	// GetPaymentList()
}

// Option defines an option for qpay initialization.
type Option func(*qpay)

// WithClient allows providing a custom resty.Client instance.
// This is useful for injecting a client with custom timeouts,
// certificates, middlewares, or connection pools.
func WithClient(client *resty.Client) Option {
	return func(q *qpay) {
		q.client = client
	}
}

func New(username, password, endpoint, callback, invoiceCode, merchantId string, options ...Option) QPay {
	q := &qpay{
		endpoint:    endpoint,
		password:    password,
		username:    username,
		callback:    callback,
		invoiceCode: invoiceCode,
		merchantId:  merchantId,
		client:      resty.New(),
	}

	for _, opt := range options {
		opt(q)
	}

	// Login right after setting configuration/client
	q.loginObject = func() *qpayLoginResponse {
		authObj, authErr := authQPayV2(username, password, endpoint, callback, invoiceCode, merchantId)
		if authErr != nil {
			return &qpayLoginResponse{}
		}
		return &authObj
	}()

	return q
}

func (q *qpay) SetClient(client *resty.Client) {
	q.client = client
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

	var response QPaySimpleInvoiceResponse
	err := q.httpRequestQPay(request, &response, QPayInvoiceCreate, "")
	if err != nil {
		return QPaySimpleInvoiceResponse{}, q, err
	}

	return response, q, nil
}
func (q *qpay) GetInvoice(invoiceId string) (QpayInvoiceGetResponse, QPay, error) {
	var response QpayInvoiceGetResponse
	err := q.httpRequestQPay(nil, &response, QPayInvoiceGet, invoiceId)
	if err != nil {
		return QpayInvoiceGetResponse{}, q, err
	}

	return response, q, nil
}
func (q *qpay) CancelInvoice(invoiceId string) (interface{}, QPay, error) {
	var response interface{}
	err := q.httpRequestQPay(nil, &response, QPayInvoiceCancel, invoiceId)
	if err != nil {
		return nil, q, err
	}

	return response, q, nil
}

func (q *qpay) GetPayment(invoiceId string) (interface{}, QPay, error) {
	var response interface{}
	err := q.httpRequestQPay(nil, &response, QPayPaymentGet, invoiceId)
	if err != nil {
		return nil, q, err
	}

	return response, q, nil
}

func (q *qpay) CheckPayment(invoiceId string, pageLimit, pageNumber int64) (QpayPaymentCheckResponse, QPay, error) {
	req := QpayPaymentCheckRequest{}
	req.ObjectID = invoiceId
	req.ObjectType = "INVOICE"
	req.Offset.PageLimit = pageLimit
	req.Offset.PageNumber = pageNumber

	var response QpayPaymentCheckResponse
	err := q.httpRequestQPay(req, &response, QPayPaymentCheck, "")
	if err != nil {
		return response, q, err
	}

	return response, q, nil
}

func (q *qpay) CancelPayment(invoiceId, paymentUUID string) (QpayPaymentCheckResponse, QPay, error) {
	var req QpayPaymentCancelRequest

	req.CallbackUrl = q.callback + paymentUUID
	req.Note = "Cancel payment - " + invoiceId

	var response QpayPaymentCheckResponse
	err := q.httpRequestQPay(req, &response, QPayPaymentCancel, invoiceId)
	if err != nil {
		return response, q, err
	}

	return response, q, nil
}

func (q *qpay) RefundPayment(invoiceId, paymentUUID string) (interface{}, QPay, error) {
	var req QpayPaymentCancelRequest

	req.CallbackUrl = q.callback + paymentUUID
	req.Note = "Cancel payment - " + invoiceId

	var response interface{}
	err := q.httpRequestQPay(req, &response, QPayPaymentRefund, invoiceId)
	if err != nil {
		return response, q, err
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
