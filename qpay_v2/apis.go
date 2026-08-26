package qpay_v2

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/techpartners-asia/qpay-go/utils"
	"resty.dev/v3"
)

var (
	// QPayAuthToken [Access Token авах]
	QPayAuthToken = utils.API{
		Url:    "/auth/token",
		Method: http.MethodPost,
	}
	// QPayAuthRefresh [Access Token шинэчлэх]
	QPayAuthRefresh = utils.API{
		Url:    "/auth/refresh",
		Method: http.MethodPost,
	}
	// QPayInvoiceCreate [Нэхэмжлэх үүсгэх]
	QPayInvoiceCreate = utils.API{
		Url:    "/invoice",
		Method: http.MethodPost,
	}
	// QPayInvoiceGet [Нэхэмжлэх харах]
	QPayInvoiceGet = utils.API{
		Url:    "/invoice/",
		Method: http.MethodGet,
	}
	// QPayInvoiceCancel [Нэхэмжлэх цуцлах]
	QPayInvoiceCancel = utils.API{
		Url:    "/invoice/",
		Method: http.MethodDelete,
	}
	// QPayPaymentGet [Төлбөр харах]
	QPayPaymentGet = utils.API{
		Url:    "/payment/",
		Method: http.MethodGet,
	}
	// QPayPaymentCheck [Төлбөр шалгах]
	QPayPaymentCheck = utils.API{
		Url:    "/payment/check",
		Method: http.MethodPost,
	}
	// QPayPaymentCancel [Төлбөр цуцлах]
	QPayPaymentCancel = utils.API{
		Url:    "/payment/cancel/",
		Method: http.MethodDelete,
	}
	// QPayPaymentRefund [Төлбөр буцаах]
	QPayPaymentRefund = utils.API{
		Url:    "/payment/refund/",
		Method: http.MethodDelete,
	}
	// QPayPaymentList [Төлбөрийн жагсаалт]
	QPayPaymentList = utils.API{
		Url:    "/payment/list",
		Method: http.MethodPost,
	}
	// QPayEbarimtCreate [И-баримт үүсгэх]
	QPayEbarimtCreate = utils.API{
		Url:    "/ebarimt_v3/create",
		Method: http.MethodPost,
	}
	// QPayEbarimtCancel [И-баримт цуцлах]
	QPayEbarimtCancel = utils.API{
		Url:    "/ebarimt_v3/",
		Method: http.MethodDelete,
	}
)

// httpRequestQPay [Internal: QPay API-руу HTTP хүсэлт илгээх туслах функц]
// body: Хүсэлтийн бие (POST/PUT үед)
// result: Хариуг задлах бүтэц (struct pointer)
// api: utils.API төрлийн эндпоинт тохиргоо
// urlExt: URL-д залгагдах нэмэлт ID (invoice_id, payment_id г.м)
func (q *qpay) httpRequestQPay(body interface{}, result interface{}, api utils.API, urlExt string) error {
	token := q.Token()
	if token.IsZero() {
		// The SDK no longer authenticates behind the caller's back, so an
		// absent token is reported rather than quietly fetched: sending the
		// request with an empty bearer would surface as a confusing 401.
		return ErrNoToken
	}

	url := q.endpoint + api.Url + utils.EscapePathSegment(urlExt)
	req := q.client.R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(token.AccessToken).
		SetResult(result)

	// Standard guard: avoid sending identity bodies on non-mutation requests
	if body != nil {
		req.SetBody(body)
	}

	res, err := req.Execute(api.Method, url)
	if err != nil {
		return err
	}
	defer closeBody(res)

	// A rejected token is its own error so the caller can tell "replace the
	// token and retry" apart from "qPay refused this request".
	if res.StatusCode() == http.StatusUnauthorized || res.StatusCode() == http.StatusForbidden {
		return fmt.Errorf("%w (Status: %d): %s", ErrUnauthorized,
			res.StatusCode(), utils.TruncateForError(res.String()))
	}

	// Anything outside 2xx is an error. Checking only for >= 400 let 3xx
	// responses through with `result` left at its zero value, which reads
	// downstream as a successful-but-empty payment.
	if !res.IsStatusSuccess() {
		return fmt.Errorf("%s-QPay response error: %s (Status: %d)",
			time.Now().Format("2006-01-02 15:04:05"),
			utils.TruncateForError(res.String()),
			res.StatusCode())
	}

	return nil
}

// closeBody releases a response body. Resty drains and closes it while
// decoding 2xx and >=400 responses, but not for 204 or 3xx, and the body's
// Close is what cancels the per-request timeout context.
func closeBody(res *resty.Response) {
	if res != nil && res.Body != nil {
		_ = res.Body.Close()
	}
}

// Login [qPay-ээс Access Token авах]
//
// Login performs exactly one request and caches nothing: the returned token is
// the caller's to hold, store and install with [QPay.SetToken]. Concurrent
// callers each issue their own request, so deduplicating them is the caller's
// job too — the SDK has no way to know whether a token is shared across
// processes.
//
// See: https://developer.qpay.mn/#auth-token
func (q *qpay) Login(ctx context.Context) (Token, error) {
	var authRes qpayLoginResponse
	res, err := q.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBasicAuth(q.username, q.password).
		SetResult(&authRes).
		Post(q.endpoint + QPayAuthToken.Url)
	if err != nil {
		return Token{}, err
	}
	defer closeBody(res)
	if !res.IsStatusSuccess() {
		return Token{}, fmt.Errorf("%s-QPay auth failed: %s (Status: %d)",
			time.Now().Format("2006-01-02 15:04:05"), utils.TruncateForError(res.String()), res.StatusCode())
	}
	if authRes.AccessToken == "" {
		// Returning a tokenless response as success would send every later
		// request out with an empty bearer.
		return Token{}, errors.New("qpay: auth response contained no access token")
	}
	return tokenFrom(authRes), nil
}

// Refresh [Refresh token ашиглан access token шинэчлэх]
//
// Refresh is a single request, like [QPay.Login]. It does not fall back to a
// full login when the refresh token is rejected: the caller sees the error and
// decides, because only the caller knows whether a shared token was already
// replaced by someone else.
func (q *qpay) Refresh(ctx context.Context, refreshToken string) (Token, error) {
	if refreshToken == "" {
		return Token{}, errors.New("qpay: refresh token is required")
	}

	var authRes qpayLoginResponse
	res, err := q.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetAuthToken(refreshToken).
		SetResult(&authRes).
		Post(q.endpoint + QPayAuthRefresh.Url)
	if err != nil {
		return Token{}, err
	}
	defer closeBody(res)
	if res.StatusCode() == http.StatusUnauthorized || res.StatusCode() == http.StatusForbidden {
		return Token{}, fmt.Errorf("%w (Status: %d): %s", ErrUnauthorized,
			res.StatusCode(), utils.TruncateForError(res.String()))
	}
	if !res.IsStatusSuccess() {
		return Token{}, fmt.Errorf("%s-QPay refresh failed: %s (Status: %d)",
			time.Now().Format("2006-01-02 15:04:05"), utils.TruncateForError(res.String()), res.StatusCode())
	}
	if authRes.AccessToken == "" {
		return Token{}, errors.New("qpay: refresh response contained no access token")
	}
	return tokenFrom(authRes), nil
}

// SetToken installs the token subsequent calls will carry. Passing the zero
// Token clears it, which makes the next call fail with [ErrNoToken] rather
// than reach qPay unauthenticated.
func (q *qpay) SetToken(token Token) {
	q.mu.Lock()
	q.token = token
	q.mu.Unlock()
}

// Token returns the installed token.
func (q *qpay) Token() Token {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.token
}
