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
	refreshMu   sync.Mutex // Serializes re-auth calls when mu is unlocked
	client      *resty.Client
}

// QPay [QPay V2 SDK Interface / Интерфэйс]
type QPay interface {
	// CreateInvoice [Төлбөрийн нэхэмжлэл үүсгэх]
	// See: https://developer.qpay.mn/#invoice-Create
	CreateInvoice(input QPayCreateInvoiceInput) (QPaySimpleInvoiceResponse, error)

	// GetInvoice [Үүсгэсэн нэхэмжлэлийн мэдээлэл харах]
	// See: https://developer.qpay.mn/#invoice-Get
	GetInvoice(invoiceId string) (QpayInvoiceGetResponse, error)

	// CancelInvoice [Нэхэмжлэх цуцлах]
	// See: https://developer.qpay.mn/#invoice-Cancel
	CancelInvoice(invoiceId string) (QpayGeneralResponse, error)

	// GetPayment [Төлбөрийн мэдээлэл татах]
	// See: https://developer.qpay.mn/#payment-Get
	GetPayment(paymentId string) (QpayTransaction, error)

	// CheckPayment [Төлбөр төлөгдсөн эсэхийг шалгах]
	// See: https://developer.qpay.mn/#payment-check
	CheckPayment(invoiceId string, pageLimit, pageNumber int64) (QpayPaymentCheckResponse, error)

	// CancelPayment [Төлөгдсөн төлбөрийг цуцлах]
	// See: https://developer.qpay.mn/#payment-cancel
	CancelPayment(invoiceId, paymentId string) (QpayGeneralResponse, error)

	// RefundPayment [Төлбөр буцаах]
	// See: https://developer.qpay.mn/#payment-refund
	RefundPayment(invoiceId, paymentId string) (QpayGeneralResponse, error)

	// GetPaymentList [Төлбөрийн жагсаалт авах]
	// See: https://developer.qpay.mn/#payment-list
	GetPaymentList(input QPayPaymentListInput) (QpayPaymentListResponse, error)
}

// Option defines an option for qpay initialization.
type Option func(*qpay)

// WithClient [Custom resty.Client ашиглах]
// This is useful for injecting a client with custom timeouts, certificates, etc.
func WithClient(client *resty.Client) Option {
	return func(q *qpay) {
		if client != nil {
			q.client = client
		}
	}
}

// New [QPay V2 SDK-ийг шинээр үүсгэх]
// username: qPay-ээс өгсөн хэрэглэгчийн нэр (client_id)
// password: qPay-ээс өгсөн нууц үг (client_secret)
// endpoint: Sandbox эсвэл Production хаяг
// callback: Төлбөр төлөгдсөний дараа дуудагдах URL
// invoiceCode: qPay нэхэмжлэхийн код
// merchantId: Байгууллагын ID
func New(username, password, endpoint, callback, invoiceCode, merchantId string, options ...Option) QPay {
	q := &qpay{
		endpoint:    endpoint,
		password:    password,
		username:    username,
		callback:    callback,
		invoiceCode: invoiceCode,
		merchantId:  merchantId,
		client:      resty.New().SetTimeout(60 * time.Second),
	}

	for _, opt := range options {
		opt(q)
	}

	// Attempt login in background to warm the token cache.
	// If it fails (network down or bad config), authQPayV2 will retry
	// transparently on the first real API call.
	go q.authQPayV2() //nolint:errcheck

	return q
}

// CreateInvoice [Нэхэмжлэх үүсгэх]
func (q *qpay) CreateInvoice(input QPayCreateInvoiceInput) (QPaySimpleInvoiceResponse, error) {
	vals := url.Values{}
	for k, v := range input.CallbackParam {
		vals.Add(k, v)
	}

	callbackUrl := q.callback
	if len(vals) > 0 {
		callbackUrl = fmt.Sprintf("%s?%s", q.callback, vals.Encode())
	}

	var minAmt *int64
	if input.MinimumAmount > 0 {
		minAmt = &input.MinimumAmount
	}
	var maxAmt *int64
	if input.MaximumAmount > 0 {
		maxAmt = &input.MaximumAmount
	}

	request := QPaySimpleInvoiceRequest{
		InvoiceCode:         q.invoiceCode,
		SenderInvoiceNo:     input.SenderCode,
		SenderBranchCode:    input.SenderBranchCode,
		SenderBranchData:    input.SenderBranchData,
		SenderTerminalCode:  input.SenderTerminalCode,
		SenderTerminalData:  input.SenderTerminalData,
		SenderStaffCode:     input.SenderStaffCode,
		SenderStaffData:     input.SenderStaffData,
		InvoiceReceiverCode: input.ReceiverCode,
		InvoiceReceiverData: input.ReceiverData,
		InvoiceDescription:  input.Description,
		Amount:              input.Amount,
		CallbackUrl:         callbackUrl,
		InvoiceDueDate:      input.InvoiceDueDate,
		ExpiryDate:          input.ExpiryDate,
		EnableExpiry:        input.EnableExpiry,
		AllowPartial:        input.AllowPartial,
		MinimumAmount:       minAmt,
		AllowExceed:         input.AllowExceed,
		MaximumAmount:       maxAmt,
		CalculateVat:        input.CalculateVat,
		Lines:               input.Lines,
		Note:                input.Note,
		TaxCustomerCode:     input.TaxCustomerCode,
		LineTaxCode:         input.LineTaxCode,
		Transactions:        input.Transactions,
	}

	var response QPaySimpleInvoiceResponse
	err := q.httpRequestQPay(request, &response, QPayInvoiceCreate, "")
	if err != nil {
		return QPaySimpleInvoiceResponse{}, err
	}

	return response, nil
}

// GetInvoice [Нэхэмжлэхийн мэдээлэл авах]
func (q *qpay) GetInvoice(invoiceId string) (QpayInvoiceGetResponse, error) {
	var response QpayInvoiceGetResponse
	err := q.httpRequestQPay(nil, &response, QPayInvoiceGet, invoiceId)
	if err != nil {
		return QpayInvoiceGetResponse{}, err
	}

	return response, nil
}

// CancelInvoice [Үүсгэсэн нэхэмжлэлийг цуцлах]
func (q *qpay) CancelInvoice(invoiceId string) (QpayGeneralResponse, error) {
	var response QpayGeneralResponse
	err := q.httpRequestQPay(nil, &response, QPayInvoiceCancel, invoiceId)
	if err != nil {
		return QpayGeneralResponse{}, err
	}

	return response, nil
}

// GetPayment [Төлбөрийн мэдээлэл татах]
func (q *qpay) GetPayment(paymentId string) (QpayTransaction, error) {
	var response QpayTransaction
	err := q.httpRequestQPay(nil, &response, QPayPaymentGet, paymentId)
	if err != nil {
		return QpayTransaction{}, err
	}

	return response, nil
}

// CheckPayment [Нэхэмжлэлийн төлбөрийг шалгах]
func (q *qpay) CheckPayment(invoiceId string, pageLimit, pageNumber int64) (QpayPaymentCheckResponse, error) {
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
		return response, err
	}

	return response, nil
}

// CancelPayment [Төлөгдсөн төлбөрийг цуцлах]
func (q *qpay) CancelPayment(invoiceId, paymentId string) (QpayGeneralResponse, error) {
	req := QpayPaymentCancelRequest{
		CallbackUrl: q.callback,
		Note:        "Cancel payment for invoice: " + invoiceId,
	}

	var response QpayGeneralResponse
	err := q.httpRequestQPay(req, &response, QPayPaymentCancel, paymentId)
	if err != nil {
		return QpayGeneralResponse{}, err
	}

	return response, nil
}

// RefundPayment [Төлбөр буцаалт хийх]
func (q *qpay) RefundPayment(invoiceId, paymentId string) (QpayGeneralResponse, error) {
	req := QpayPaymentCancelRequest{
		CallbackUrl: q.callback,
		Note:        "Refund payment for invoice: " + invoiceId,
	}

	var response QpayGeneralResponse
	err := q.httpRequestQPay(req, &response, QPayPaymentRefund, paymentId)
	if err != nil {
		return QpayGeneralResponse{}, err
	}

	return response, nil
}

// GetPaymentList [Төлбөр төлөлтийн жагсаалт авах]
func (q *qpay) GetPaymentList(input QPayPaymentListInput) (QpayPaymentListResponse, error) {
	// Default to MERCHANT if not specified
	objType := input.ObjectType
	if objType == "" {
		objType = "MERCHANT"
	}

	// Default to q.merchantId if not specified
	objID := input.ObjectID
	if objID == "" {
		objID = q.merchantId
	}

	req := QpayPaymentListRequest{
		ObjectType:           objType,
		ObjectID:             objID,
		MerchantBranchCode:   input.BranchCode,
		MerchantTerminalCode: input.TerminalCode,
		MerchantStaffCode:    input.StaffCode,
		Offset: QpayOffset{
			PageLimit:  input.PageLimit,
			PageNumber: input.PageNumber,
		},
	}

	var response QpayPaymentListResponse
	err := q.httpRequestQPay(req, &response, QPayPaymentList, "")
	if err != nil {
		return QpayPaymentListResponse{}, err
	}

	return response, nil
}
