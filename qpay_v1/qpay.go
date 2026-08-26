package qpay_v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/techpartners-asia/qpay-go/utils"
)

type qpay struct {
	endpoint      string
	client_id     string
	client_secret string
	grant_type    string
	callback      string
	merchantId    string
	templateId    string
	branchId      string
	posId         string

	// token is the credential installed by SetToken. The SDK reads it and
	// never populates it on its own; see Token.
	mu    sync.RWMutex
	token Token

	client *http.Client
}

// QPay [QPay V1 SDK Interface]
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

	CreateInvoice(input QPayInvoiceCreateRequest) (QPaySimpleInvoiceResponse, error)
	GetInvoice(invoiceId string) (QpayInvoiceGetResponse, error)
	CheckPayment(paymentID string) (QpayPaymentCheckResponse, error)
}

// New performs no network I/O. The returned client has no token until one is
// installed with [QPay.SetToken]; see [QPay] on authentication.
func New(client_id, client_secret, endpoint, callback, merchantId, templateId, branchId, posId string) QPay {
	return &qpay{
		endpoint:      endpoint,
		client_id:     client_id,
		client_secret: client_secret,
		grant_type:    "client",
		callback:      callback,
		merchantId:    merchantId,
		templateId:    templateId,
		branchId:      branchId,
		posId:         posId,
		client:        utils.NewHTTPClient(),
	}
}

func (q *qpay) CreateInvoice(input QPayInvoiceCreateRequest) (QPaySimpleInvoiceResponse, error) {
	input.BranchID = q.branchId
	input.PosID = q.posId
	input.MerchantID = q.merchantId
	input.TemplateID = q.templateId

	res, err := q.httpRequestQPay(input, QPayInvoiceCreate, "")
	if err != nil {
		return QPaySimpleInvoiceResponse{}, err
	}

	var response QPaySimpleInvoiceResponse
	if err := json.Unmarshal(res, &response); err != nil {
		return QPaySimpleInvoiceResponse{}, fmt.Errorf("qpay: decode invoice create response: %w", err)
	}

	return response, nil
}

func (q *qpay) GetInvoice(invoiceId string) (QpayInvoiceGetResponse, error) {
	res, err := q.httpRequestQPay(nil, QPayInvoiceGet, invoiceId)
	if err != nil {
		return QpayInvoiceGetResponse{}, err
	}

	var response QpayInvoiceGetResponse
	if err := json.Unmarshal(res, &response); err != nil {
		return QpayInvoiceGetResponse{}, fmt.Errorf("qpay: decode invoice get response: %w", err)
	}

	return response, nil
}

func (q *qpay) CheckPayment(paymentID string) (QpayPaymentCheckResponse, error) {
	var response QpayPaymentCheckResponse

	res, err := q.httpRequestQPay(nil, QPayPaymentCheck, paymentID)
	if err != nil {
		return response, err
	}

	if err := json.Unmarshal(res, &response); err != nil {
		return QpayPaymentCheckResponse{}, fmt.Errorf("qpay: decode payment check response: %w", err)
	}

	return response, nil
}
