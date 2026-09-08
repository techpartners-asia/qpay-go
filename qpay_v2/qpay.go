package qpay_v2

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/techpartners-asia/qpay-go/utils"
	"resty.dev/v3"
)

type qpay struct {
	endpoint    string
	password    string
	username    string
	callback    string
	invoiceCode string
	merchantId  string

	// token is the credential installed by SetToken. The SDK reads it and
	// never populates it on its own: expiry tracking, refresh scheduling and
	// deduplication of concurrent logins all belong to the caller, which is
	// the only layer that knows whether the token is shared beyond this
	// process. See Token.
	mu    sync.RWMutex
	token Token

	client *resty.Client
}

// QPay [QPay V2 SDK Interface / Интерфэйс]
//
// # Authentication
//
// This SDK does not manage tokens. Obtain one with [QPay.Login] (or
// [QPay.Refresh]), install it with [QPay.SetToken], and every call below
// carries it. A call made with no token installed fails with [ErrNoToken]; a
// call whose token qPay rejects fails with [ErrUnauthorized], which is the
// signal to obtain a fresh token and retry.
type QPay interface {
	// Login [Access Token авах] — one request, no caching.
	Login(ctx context.Context) (Token, error)

	// Refresh [Access Token шинэчлэх] — one request, no caching.
	Refresh(ctx context.Context, refreshToken string) (Token, error)

	// SetToken installs the token subsequent calls carry.
	SetToken(token Token)

	// Token returns the installed token.
	Token() Token

	// CreateInvoice [Төлбөрийн нэхэмжлэл үүсгэх]
	// See: https://developer.qpay.mn/#invoice-Create
	CreateInvoice(input QPayCreateInvoiceInput) (QPaySimpleInvoiceResponse, error)

	// CreateEbarimtInvoice [И-баримт 3.0 мэдээлэлтэй нэхэмжлэх үүсгэх]
	CreateEbarimtInvoice(input QPayCreateEbarimtInvoiceInput) (QPaySimpleInvoiceResponse, error)

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

	// CreateEbarimt [Төлбөр төлөгдсөний дараа и-баримт 3.0 үүсгэх]
	CreateEbarimt(input QPayEbarimtCreateInput) (QPayEbarimtResponse, error)

	// CancelEbarimt [И-баримт 3.0 цуцлах]
	CancelEbarimt(paymentId string) (QPayEbarimtResponse, error)
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

// WithToken [Токеныг эхлүүлэхдээ шингээх]
// Installs a token at construction time, for a caller that already holds a
// valid one — from a shared cache, say — and wants the first call to go out
// authenticated without a login round trip.
func WithToken(token Token) Option {
	return func(q *qpay) {
		q.token = token
	}
}

// New [QPay V2 SDK-ийг шинээр үүсгэх]
// username: qPay-ээс өгсөн хэрэглэгчийн нэр (client_id)
// password: qPay-ээс өгсөн нууц үг (client_secret)
// endpoint: Sandbox эсвэл Production хаяг
// callback: Төлбөр төлөгдсөний дараа дуудагдах URL
// invoiceCode: qPay нэхэмжлэхийн код
// merchantId: Байгууллагын ID
//
// New performs no network I/O. The returned client has no token until one is
// installed with [QPay.SetToken] or [WithToken]; see [QPay] on authentication.
func New(username, password, endpoint, callback, invoiceCode, merchantId string, options ...Option) QPay {
	q := &qpay{
		endpoint:    endpoint,
		password:    password,
		username:    username,
		callback:    callback,
		invoiceCode: invoiceCode,
		merchantId:  merchantId,
		client:      resty.New().SetTransport(newTransport()).SetTimeout(60 * time.Second),
	}

	for _, opt := range options {
		opt(q)
	}

	return q
}

// CreateInvoice [Нэхэмжлэх үүсгэх]
func (q *qpay) CreateInvoice(input QPayCreateInvoiceInput) (QPaySimpleInvoiceResponse, error) {
	callbackUrl := q.callback
	if input.CallbackUrl != nil {
		callbackUrl = *input.CallbackUrl
	}
	// Merge params into any query string the callback URL already carries;
	// blindly appending "?" produced a second one and silently dropped them.
	callbackUrl = utils.BuildCallbackURL(callbackUrl, input.CallbackParam)

	var minAmt *int64
	if input.MinimumAmount > 0 {
		minAmt = &input.MinimumAmount
	}
	var maxAmt *int64
	if input.MaximumAmount > 0 {
		maxAmt = &input.MaximumAmount
	}

	request := QPaySimpleInvoiceRequest{
		InvoiceCode:          q.invoiceCode,
		SenderInvoiceNo:      input.SenderInvoiceNo,
		SenderBranchCode:     input.SenderBranchCode,
		SenderBranchData:     input.SenderBranchData,
		SenderTerminalCode:   input.SenderTerminalCode,
		SenderTerminalData:   input.SenderTerminalData,
		SenderStaffCode:      input.SenderStaffCode,
		SenderStaffData:      input.SenderStaffData,
		InvoiceReceiverCode:  input.InvoiceReceiverCode,
		InvoiceReceiverData:  input.InvoiceReceiverData,
		InvoiceDescription:   input.InvoiceDescription,
		Amount:               input.Amount,
		CallbackUrl:          callbackUrl,
		InvoiceDueDate:       input.InvoiceDueDate,
		ExpiryDate:           input.ExpiryDate,
		EnableExpiry:         input.EnableExpiry,
		AllowPartial:         input.AllowPartial,
		MinimumAmount:        minAmt,
		AllowExceed:          input.AllowExceed,
		MaximumAmount:        maxAmt,
		CalculateVat:         input.CalculateVat,
		Lines:                input.Lines,
		Note:                 input.Note,
		TaxCustomerCode:      input.TaxCustomerCode,
		LineTaxCode:          input.LineTaxCode,
		Transactions:         input.Transactions,
		AllowSubscribe:       input.AllowSubscribe,
		SubscriptionInterval: input.SubscriptionInterval,
		SubscriptionWebhook:  input.SubscriptionWebhook,
		TaxType:              string(input.TaxType),
		DistrictCode:         input.DistrictCode,
		Lottery:              input.Lottery,
	}

	var response QPaySimpleInvoiceResponse
	err := q.httpRequestQPay(request, &response, QPayInvoiceCreate, "")
	if err != nil {
		return QPaySimpleInvoiceResponse{}, err
	}

	return response, nil
}

// CreateEbarimtInvoice [И-баримт 3.0 мэдээлэлтэй нэхэмжлэх үүсгэх]
func (q *qpay) CreateEbarimtInvoice(input QPayCreateEbarimtInvoiceInput) (QPaySimpleInvoiceResponse, error) {
	request := q.newEbarimtInvoiceRequest(input)

	var response QPaySimpleInvoiceResponse
	err := q.httpRequestQPay(request, &response, QPayInvoiceCreate, "")
	if err != nil {
		return QPaySimpleInvoiceResponse{}, err
	}

	return response, nil
}

func (q *qpay) newEbarimtInvoiceRequest(input QPayCreateEbarimtInvoiceInput) QPayEbarimtInvoiceRequest {
	callbackURL := input.CallbackURL
	if callbackURL == "" {
		callbackURL = q.callback
	}
	callbackURL = utils.BuildCallbackURL(callbackURL, input.CallbackParam)

	invoiceCode := input.InvoiceCode
	if invoiceCode == "" {
		invoiceCode = q.invoiceCode
	}

	calculateVat := input.CalculateVat
	if calculateVat == nil && (input.TaxType == QPayTaxTypeNoVAT || input.TaxType == QPayTaxTypeVATExempt) {
		value := false
		calculateVat = &value
	}

	return QPayEbarimtInvoiceRequest{
		InvoiceCode:         invoiceCode,
		SenderInvoiceNo:     input.SenderInvoiceNo,
		SenderBranchCode:    input.SenderBranchCode,
		SenderStaffCode:     input.SenderStaffCode,
		SenderStaffData:     input.SenderStaffData,
		InvoiceReceiverCode: input.InvoiceReceiverCode,
		InvoiceReceiverData: input.InvoiceReceiverData,
		InvoiceDescription:  input.InvoiceDescription,
		TaxType:             input.TaxType,
		DistrictCode:        input.DistrictCode,
		CallbackUrl:         callbackURL,
		CalculateVat:        calculateVat,
		Lines:               input.Lines,
	}
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
		// Return the zero value like every other method: a half-filled
		// payment-check result alongside an error invites the caller to
		// read Rows/PaidAmount as if the check had succeeded.
		return QpayPaymentCheckResponse{}, err
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

// newTransport creates an http.Transport with sensible defaults.
func newTransport() *http.Transport {
	return &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12},
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		MaxConnsPerHost:       20,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ForceAttemptHTTP2:     true,
	}
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
		StartDate:            input.StartDate,
		EndDate:              input.EndDate,
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

// CreateEbarimt [Төлбөр төлөгдсөний дараа и-баримт 3.0 үүсгэх]
func (q *qpay) CreateEbarimt(input QPayEbarimtCreateInput) (QPayEbarimtResponse, error) {
	request := QPayEbarimtCreateRequest{
		PaymentID:           input.PaymentID,
		EbarimtReceiverType: input.EbarimtReceiverType,
		EbarimtReceiver:     input.EbarimtReceiver,
		DistrictCode:        input.DistrictCode,
		ClassificationCode:  input.ClassificationCode,
	}

	var response QPayEbarimtResponse
	err := q.httpRequestQPay(request, &response, QPayEbarimtCreate, "")
	if err != nil {
		return QPayEbarimtResponse{}, err
	}

	return response, nil
}

// CancelEbarimt [И-баримт 3.0 цуцлах]
func (q *qpay) CancelEbarimt(paymentId string) (QPayEbarimtResponse, error) {
	var response QPayEbarimtResponse
	err := q.httpRequestQPay(nil, &response, QPayEbarimtCancel, paymentId)
	if err != nil {
		return QPayEbarimtResponse{}, err
	}

	return response, nil
}
