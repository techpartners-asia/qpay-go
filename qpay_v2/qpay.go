package qpay_v2

import (
	"fmt"
	"net/url"
	"sync"
	"time"

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
	loginTime   time.Time
	mu          sync.RWMutex
	client      *resty.Client
}

// QPay [QPay V2 SDK Interface / Интерфэйс]
type QPay interface {
	// CreateInvoice [Төлбөрийн нэхэмжлэл үүсгэх]
	// See: https://developer.qpay.mn/#invoice-Create
	CreateInvoice(input QPayCreateInvoiceInput) (QPaySimpleInvoiceResponse, QPay, error)

	// GetInvoice [Үүсгэсэн нэхэмжлэлийн мэдээлэл харах]
	// See: https://developer.qpay.mn/#invoice-Get
	GetInvoice(invoiceId string) (QpayInvoiceGetResponse, QPay, error)

	// CancelInvoice [Нэхэмжлэх цуцлах]
	// See: https://developer.qpay.mn/#invoice-Cancel
	CancelInvoice(invoiceId string) (interface{}, QPay, error)

	// GetPayment [Төлбөрийн мэдээлэл татах]
	// See: https://developer.qpay.mn/#payment-Get
	GetPayment(paymentId string) (interface{}, QPay, error)

	// CheckPayment [Төлбөр төлөгдсөн эсэхийг шалгах]
	// See: https://developer.qpay.mn/#payment-check
	CheckPayment(invoiceId string, pageLimit, pageNumber int64) (QpayPaymentCheckResponse, QPay, error)

	// CancelPayment [Төлөгдсөн төлбөрийг цуцлах]
	// See: https://developer.qpay.mn/#payment-cancel
	CancelPayment(invoiceId, paymentId string) (QpayPaymentCheckResponse, QPay, error)

	// RefundPayment [Төлбөр буцаах]
	// See: https://developer.qpay.mn/#payment-refund
	RefundPayment(invoiceId, paymentId string) (interface{}, QPay, error)

	// GetPaymentList [Төлбөрийн жагсаалт авах]
	// See: https://developer.qpay.mn/#payment-list
	GetPaymentList(pageLimit, pageNumber int64) (interface{}, QPay, error)
}

// Option defines an option for qpay initialization.
type Option func(*qpay)

// WithClient [Custom resty.Client ашиглах]
// This is useful for injecting a client with custom timeouts, certificates, etc.
func WithClient(client *resty.Client) Option {
	return func(q *qpay) {
		q.client = client
	}
}

// New [QPay V2 SDK-ийг шинээр үүсгэх]
// username: qPay-ээс өгсөн хэрэглэгчийн нэр (client_id)
// password: qPay-ээс өгсөн нууц үг (client_secret)
// endpoint: Sandbox эсвэл Production хаяг
// callback: Төлбөр төлөгдсөний дараа дуудагдах URL
// invoiceCode: qPay нэхэмжлэхийн код
// merchantId: Байгууллагын ID
func New(username, password, endpoint, callback, invoiceCode, merchantId string, options ...Option) (QPay, error) {
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

	// Initial authentication to verify credentials and warm up the cache
	_, authErr := q.authQPayV2()
	if authErr != nil {
		return nil, authErr
	}

	return q, nil
}

// SetClient [Гаднаас resty.Client тохируулах]
func (q *qpay) SetClient(client *resty.Client) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.client = client
}

// CreateInvoice [Нэхэмжлэх үүсгэх]
func (q *qpay) CreateInvoice(input QPayCreateInvoiceInput) (QPaySimpleInvoiceResponse, QPay, error) {
	vals := url.Values{}
	for k, v := range input.CallbackParam {
		vals.Add(k, v)
	}

	callbackUrl := q.callback
	if len(vals) > 0 {
		callbackUrl = fmt.Sprintf("%s?%s", q.callback, vals.Encode())
	}

	request := QPaySimpleInvoiceRequest{
		InvoiceCode:         q.invoiceCode,
		SenderInvoiceCode:   input.SenderCode,
		SenderBranchCode:    input.SenderBranchCode,
		InvoiceReceiverCode: input.ReceiverCode,
		InvoiceReceiverData: input.ReceiverData,
		InvoiceDescription:  input.Description,
		Amount:              input.Amount,
		CallbackUrl:         callbackUrl,
		Lines:               input.Lines,
		Note:                input.Note,
	}

	var response QPaySimpleInvoiceResponse
	err := q.httpRequestQPay(request, &response, QPayInvoiceCreate, "")
	if err != nil {
		return QPaySimpleInvoiceResponse{}, q, err
	}

	return response, q, nil
}

// GetInvoice [Нэхэмжлэхийн мэдээлэл авах]
func (q *qpay) GetInvoice(invoiceId string) (QpayInvoiceGetResponse, QPay, error) {
	var response QpayInvoiceGetResponse
	err := q.httpRequestQPay(nil, &response, QPayInvoiceGet, invoiceId)
	if err != nil {
		return QpayInvoiceGetResponse{}, q, err
	}

	return response, q, nil
}

// CancelInvoice [Үүсгэсэн нэхэмжлэлийг цуцлах]
func (q *qpay) CancelInvoice(invoiceId string) (interface{}, QPay, error) {
	var response interface{}
	err := q.httpRequestQPay(nil, &response, QPayInvoiceCancel, invoiceId)
	if err != nil {
		return nil, q, err
	}

	return response, q, nil
}

// GetPayment [Төлбөрийн мэдээлэл татах]
func (q *qpay) GetPayment(paymentId string) (interface{}, QPay, error) {
	var response interface{}
	err := q.httpRequestQPay(nil, &response, QPayPaymentGet, paymentId)
	if err != nil {
		return nil, q, err
	}

	return response, q, nil
}

// CheckPayment [Нэхэмжлэлийн төлбөрийг шалгах]
func (q *qpay) CheckPayment(invoiceId string, pageLimit, pageNumber int64) (QpayPaymentCheckResponse, QPay, error) {
	req := QpayPaymentCheckRequest{
		ObjectType: "INVOICE",
		ObjectID:   invoiceId,
		Offset: QpayOffset{
			PageLimit:  pageLimit,
			PageNumber: pageNumber,
		},
	}

	var response QpayPaymentCheckResponse
	err := q.httpRequestQPay(req, &response, QPayPaymentCheck, "")
	if err != nil {
		return response, q, err
	}

	return response, q, nil
}

// CancelPayment [Төлөгдсөн төлбөрийг цуцлах]
func (q *qpay) CancelPayment(invoiceId, paymentId string) (QpayPaymentCheckResponse, QPay, error) {
	req := QpayPaymentCancelRequest{
		CallbackUrl: q.callback,
		Note:        "Cancel payment for invoice: " + invoiceId,
	}

	var response QpayPaymentCheckResponse
	err := q.httpRequestQPay(req, &response, QPayPaymentCancel, paymentId)
	if err != nil {
		return response, q, err
	}

	return response, q, nil
}

// RefundPayment [Төлбөр буцаалт хийх]
func (q *qpay) RefundPayment(invoiceId, paymentId string) (interface{}, QPay, error) {
	req := QpayPaymentCancelRequest{
		CallbackUrl: q.callback,
		Note:        "Refund payment for invoice: " + invoiceId,
	}

	var response interface{}
	err := q.httpRequestQPay(req, &response, QPayPaymentRefund, paymentId)
	if err != nil {
		return response, q, err
	}

	return response, q, nil
}

// GetPaymentList [Төлбөр төлөлтийн жагсаалт авах]
func (q *qpay) GetPaymentList(pageLimit, pageNumber int64) (interface{}, QPay, error) {
	req := QpayPaymentListRequest{
		MerchantID: q.merchantId,
		Offset: QpayOffset{
			PageLimit:  pageLimit,
			PageNumber: pageNumber,
		},
	}

	var response interface{}
	err := q.httpRequestQPay(req, &response, QPayPaymentList, "")
	if err != nil {
		return response, q, err
	}

	return response, q, nil
}
