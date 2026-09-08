package qpay_quick

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"

	"resty.dev/v3"
)

type qpayquick struct {
	endpoint   string
	password   string
	username   string
	callback   string
	terminalID string

	// token is the credential installed by SetToken. The SDK reads it and
	// never populates it on its own; see Token.
	mu    sync.RWMutex
	token Token

	client *resty.Client
}

// QPayQuick [QPay Quick Pay SDK Interface / Интерфэйс]
//
// # Authentication
//
// This SDK does not manage tokens. Obtain one with [QPayQuick.Login] (or
// [QPayQuick.Refresh]), install it with [QPayQuick.SetToken], and every call
// below carries it. A call made with no token installed fails with
// [ErrNoToken]; a call whose token qPay rejects fails with [ErrUnauthorized],
// which is the signal to obtain a fresh token and retry.
type QPayQuick interface {
	// Login [Access Token авах] — one request, no caching.
	Login(ctx context.Context) (Token, error)

	// Refresh [Access Token шинэчлэх] — one request, no caching.
	Refresh(ctx context.Context, refreshToken string) (Token, error)

	// SetToken installs the token subsequent calls carry.
	SetToken(token Token)

	// Token returns the installed token.
	Token() Token

	// CreateCompany [Байгууллага бүртгэх]
	CreateCompany(input QpayCompanyCreateRequest) (QpayCompanyCreateResponse, error)

	// CreatePerson [Хувь хүн бүртгэх]
	CreatePerson(input QpayPersonCreateRequest) (QpayPersonCreateResponse, error)

	// UpdateCompany [Байгууллагаар бүртгэсэн мерчантын мэдээлэл шинэчлэх]
	UpdateCompany(merchantID string, input QpayCompanyCreateRequest) (QpayCompanyCreateResponse, error)

	// UpdatePerson [Хувь хүнээр бүртгэсэн мерчантын мэдээлэл шинэчлэх]
	UpdatePerson(merchantID string, input QpayPersonCreateRequest) (QpayPersonCreateResponse, error)

	// GetMerchant [Мерчантын мэдээлэл харах]
	GetMerchant(merchantID string) (QpayMerchantGetResponse, error)

	// DeleteMerchant [Бүртгэлтэй мерчантыг устгах]
	DeleteMerchant(merchantID string) (QpayGeneralResponse, error)

	// ListMerchant [Мерчантуудын жагсаалт авах]
	ListMerchant(page, limit int64) (QpayMerchantListResponse, error)

	// GetAimagHot [Аймаг/хотын кодны жагсаалт]
	GetAimagHot() ([]QpayLocationCode, error)

	// GetSumDuureg [Сум/дүүргийн кодны жагсаалт (аймаг/хотын кодоор)]
	GetSumDuureg(aimagHotCode string) ([]QpayLocationCode, error)

	// CreateInvoice [Төлбөрийн нэхэмжлэл үүсгэх]
	CreateInvoice(input QpayInvoiceRequest) (QpayInvoiceResponse, error)

	// GetInvoice [Үүсгэсэн нэхэмжлэлийн мэдээлэл харах]
	GetInvoice(invoiceId string) (QpayInvoiceGetResponse, error)

	// CancelInvoice [Үүсгэсэн нэхэмжлэлийг цуцлах]
	CancelInvoice(invoiceId string) (QpayGeneralResponse, error)

	// CheckPayment [Төлбөр төлөгдсөн эсэхийг шалгах]
	CheckPayment(invoiceID string) (QpayPaymentCheckResponse, error)
}

// Option defines an option for qpayquick initialization.
type Option func(*qpayquick)

// WithClient [Custom resty.Client ашиглах]
// This is useful for injecting a client with custom timeouts, certificates, etc.
func WithClient(client *resty.Client) Option {
	return func(q *qpayquick) {
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
	return func(q *qpayquick) {
		q.token = token
	}
}

// New [QPay Quick SDK-ийг шинээр үүсгэх]
// username: qPay-ээс өгсөн хэрэглэгчийн нэр
// password: qPay-ээс өгсөн нууц үг
// endpoint: Sandbox эсвэл Production хаяг
// callback: Төлбөр төлөгдсөний дараа дуудагдах URL
// terminalID: qPay-ээс өгсөн терминалын дугаар
//
// New performs no network I/O. The returned client has no token until one is
// installed with [QPayQuick.SetToken] or [WithToken]; see [QPayQuick] on
// authentication.
func New(username, password, endpoint, callback, terminalID string, options ...Option) QPayQuick {
	q := &qpayquick{
		endpoint:   endpoint,
		password:   password,
		username:   username,
		callback:   callback,
		terminalID: terminalID,
		client:     resty.New().SetTransport(newTransport()).SetTimeout(60 * time.Second),
	}

	for _, opt := range options {
		opt(q)
	}

	return q
}

// CreateCompany [Байгууллага бүртгэх]
func (q *qpayquick) CreateCompany(input QpayCompanyCreateRequest) (QpayCompanyCreateResponse, error) {
	var response QpayCompanyCreateResponse
	if err := q.httpRequestQPay(input, &response, QPayCreateCompany, ""); err != nil {
		return QpayCompanyCreateResponse{}, err
	}
	return response, nil
}

// CreatePerson [Хувь хүн бүртгэх]
func (q *qpayquick) CreatePerson(input QpayPersonCreateRequest) (QpayPersonCreateResponse, error) {
	var response QpayPersonCreateResponse
	if err := q.httpRequestQPay(input, &response, QPayCreatePerson, ""); err != nil {
		return QpayPersonCreateResponse{}, err
	}
	return response, nil
}

// GetMerchant [Мерчантын мэдээлэл харах]
func (q *qpayquick) GetMerchant(merchantID string) (QpayMerchantGetResponse, error) {
	var response QpayMerchantGetResponse
	if err := q.httpRequestQPay(nil, &response, QPayGetMerchant, merchantID); err != nil {
		return QpayMerchantGetResponse{}, err
	}
	return response, nil
}

// UpdateCompany [Байгууллагаар бүртгэсэн мерчантын мэдээлэл шинэчлэх]
func (q *qpayquick) UpdateCompany(merchantID string, input QpayCompanyCreateRequest) (QpayCompanyCreateResponse, error) {
	var response QpayCompanyCreateResponse
	if err := q.httpRequestQPay(input, &response, QPayUpdateCompany, merchantID); err != nil {
		return QpayCompanyCreateResponse{}, err
	}
	return response, nil
}

// UpdatePerson [Хувь хүнээр бүртгэсэн мерчантын мэдээлэл шинэчлэх]
func (q *qpayquick) UpdatePerson(merchantID string, input QpayPersonCreateRequest) (QpayPersonCreateResponse, error) {
	var response QpayPersonCreateResponse
	if err := q.httpRequestQPay(input, &response, QPayUpdatePerson, merchantID); err != nil {
		return QpayPersonCreateResponse{}, err
	}
	return response, nil
}

// DeleteMerchant [Бүртгэлтэй мерчантыг устгах]
func (q *qpayquick) DeleteMerchant(merchantID string) (QpayGeneralResponse, error) {
	var response QpayGeneralResponse
	if err := q.httpRequestQPay(nil, &response, QPayDeleteMerchant, merchantID); err != nil {
		return QpayGeneralResponse{}, err
	}
	return response, nil
}

// ListMerchant [Мерчантуудын жагсаалт авах]
func (q *qpayquick) ListMerchant(page, limit int64) (QpayMerchantListResponse, error) {
	request := QpayMerchantListRequest{Page: page, Limit: limit}
	var response QpayMerchantListResponse
	if err := q.httpRequestQPay(request, &response, QPayMerchantList, ""); err != nil {
		return QpayMerchantListResponse{}, err
	}
	return response, nil
}

// GetAimagHot [Аймаг/хотын кодны жагсаалт]
func (q *qpayquick) GetAimagHot() ([]QpayLocationCode, error) {
	var response []QpayLocationCode
	if err := q.httpRequestQPay(nil, &response, QPayGetAimagHot, ""); err != nil {
		return nil, err
	}
	return response, nil
}

// GetSumDuureg [Сум/дүүргийн кодны жагсаалт]
func (q *qpayquick) GetSumDuureg(aimagHotCode string) ([]QpayLocationCode, error) {
	var response []QpayLocationCode
	if err := q.httpRequestQPay(nil, &response, QPayGetSumDuureg, aimagHotCode); err != nil {
		return nil, err
	}
	return response, nil
}

// CancelInvoice [Үүсгэсэн нэхэмжлэлийг цуцлах]
func (q *qpayquick) CancelInvoice(invoiceId string) (QpayGeneralResponse, error) {
	var response QpayGeneralResponse
	if err := q.httpRequestQPay(nil, &response, QPayInvoiceCancel, invoiceId); err != nil {
		return QpayGeneralResponse{}, err
	}
	return response, nil
}

// CreateInvoice [Төлбөрийн нэхэмжлэл үүсгэх]
func (q *qpayquick) CreateInvoice(input QpayInvoiceRequest) (QpayInvoiceResponse, error) {
	if input.CallbackUrl == "" {
		input.CallbackUrl = q.callback
	}
	var response QpayInvoiceResponse
	if err := q.httpRequestQPay(input, &response, QPayInvoiceCreate, ""); err != nil {
		return QpayInvoiceResponse{}, err
	}
	return response, nil
}

// GetInvoice [Үүсгэсэн нэхэмжлэлийн мэдээлэл харах]
func (q *qpayquick) GetInvoice(invoiceId string) (QpayInvoiceGetResponse, error) {
	var response QpayInvoiceGetResponse
	if err := q.httpRequestQPay(nil, &response, QPayInvoiceGet, invoiceId); err != nil {
		return QpayInvoiceGetResponse{}, err
	}
	return response, nil
}

// CheckPayment [Төлбөр төлөгдсөн эсэхийг шалгах]
func (q *qpayquick) CheckPayment(invoiceID string) (QpayPaymentCheckResponse, error) {
	request := QpayPaymentCheckRequest{InvoiceID: invoiceID}
	var response QpayPaymentCheckResponse
	if err := q.httpRequestQPay(request, &response, QPayPaymentCheck, ""); err != nil {
		return QpayPaymentCheckResponse{}, err
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
