package utils

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type API struct {
	Url    string `json:"url"`
	Method string `json:"method"`
}

const (
	TimeFormatYYYYMMDDHHMMSS = "20060102150405"
	TimeFormatYYYYMMDD       = "20060102"
	HttpContent              = "application/json"
	XForm                    = "application/x-www-form-urlencoded"
	XmlContent               = "application/xml"
)

// MaxResponseBytes caps how much of a response body is read into memory.
// qPay responses are small JSON documents; without a cap a hostile or
// misbehaving endpoint can exhaust the caller's memory.
const MaxResponseBytes = 8 << 20 // 8 MiB

// EscapePathSegment escapes an id before it is appended to an API path.
// Invoice and payment ids frequently originate from end users, and raw
// concatenation lets "../", "?" or "#" retarget the request at a different
// qPay endpoint.
func EscapePathSegment(segment string) string {
	return url.PathEscape(segment)
}

// BuildCallbackURL appends params to a callback URL, preserving any query
// string the URL already carries. fmt.Sprintf("%s?%s", ...) produces a
// second "?" when one is already present, which silently drops parameters.
func BuildCallbackURL(base string, params map[string]string) string {
	if len(params) == 0 {
		return base
	}

	u, err := url.Parse(base)
	if err != nil {
		// Not parseable as a URL — fall back to naive joining so the caller
		// still gets its parameters through rather than losing them.
		vals := url.Values{}
		for k, v := range params {
			vals.Add(k, v)
		}
		sep := "?"
		if strings.Contains(base, "?") {
			sep = "&"
		}
		return base + sep + vals.Encode()
	}

	vals := u.Query()
	for k, v := range params {
		vals.Set(k, v)
	}
	u.RawQuery = vals.Encode()
	return u.String()
}

// TruncateForError shortens a response body before embedding it in an error.
// Bodies can be large and may echo request data, so error strings that reach
// logs are kept bounded.
func TruncateForError(s string) string {
	const limit = 512
	if len(s) <= limit {
		return s
	}
	return s[:limit] + "...(truncated)"
}

// NewHTTPClient returns an http.Client with timeouts applied.
// http.DefaultClient has no timeout, so a stalled qPay endpoint blocks the
// calling goroutine indefinitely.
func NewHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
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
		},
	}
}

// func HttpRequestMongolchat(body interface{}, api helper.API) (res *http.Response, err error) {
// 	var requestByte []byte
// 	var requestBody *bytes.Reader
// 	if body == nil {
// 		requestBody = bytes.NewReader(nil)
// 	} else {
// 		requestByte, _ = json.Marshal(body)
// 		requestBody = bytes.NewReader(requestByte)
// 	}

// 	req, _ := http.NewRequest(api.Method, viper.GetString("mongolchat.endpoint")+api.Url, requestBody)

// 	req.Header.Add("Content-Type", helper.HttpContent)
// 	req.Header.Add("api-key", viper.GetString("mongolchat.apikey"))
// 	req.Header.Add("Authorization", "WorkerKey "+viper.GetString("mongolchat.workerkey"))

// 	res, err = http.DefaultClient.Do(req)
// 	return
// }

// func HttpRequestSocialpay(body interface{}, api helper.API) (res *http.Response, err error) {
// 	var requestByte []byte
// 	var requestBody *bytes.Reader
// 	if body == nil {
// 		requestBody = bytes.NewReader(nil)
// 	} else {
// 		requestByte, _ = json.Marshal(body)
// 		requestBody = bytes.NewReader(requestByte)
// 	}

// 	req, _ := http.NewRequest(api.Method, viper.GetString("socialpay.endpoint")+api.Url, requestBody)
// 	req.Header.Add("Content-Type", helper.HttpContent)
// 	res, err = http.DefaultClient.Do(req)
// 	return
// }

// func HttpRequestQpay(body interface{}, api helper.API, urlExt string) (res *http.Response, err error) {
// 	authObj, authErr := shared.AuthQPayV2()
// 	if authErr != nil {
// 		fmt.Println(authErr.Error())
// 		err = authErr
// 		return
// 	}
// 	var requestByte []byte
// 	var requestBody *bytes.Reader
// 	if body == nil {
// 		requestBody = bytes.NewReader(nil)
// 	} else {
// 		requestByte, _ = json.Marshal(body)
// 		requestBody = bytes.NewReader(requestByte)
// 	}

// 	req, _ := http.NewRequest(api.Method, viper.GetString("qpay.endpoint")+api.Url+urlExt, requestBody)

// 	req.Header.Add("Content-Type", helper.HttpContent)
// 	req.Header.Add("Authorization", "Bearer "+authObj.AccessToken)

// 	res, err = http.DefaultClient.Do(req)
// 	return
// }

// func HttpRequestUPoint(body interface{}, api helper.API) (res *http.Response, err error) {
// 	var requestByte []byte
// 	var requestBody *bytes.Reader
// 	if body == nil {
// 		requestBody = bytes.NewReader(nil)
// 	} else {
// 		requestByte, _ = json.Marshal(body)
// 		requestBody = bytes.NewReader(requestByte)
// 	}

// 	req, _ := http.NewRequest(api.Method, viper.GetString("upoint.endpoint")+api.Url, requestBody)

// 	req.Header.Add("Content-Type", helper.HttpContent)
// 	req.Header.Add("Authorization", "Token "+viper.GetString("upoint.token"))

// 	res, err = http.DefaultClient.Do(req)
// 	return
// }

// func HttpRequestEbarimt(body interface{}, api helper.API, ext string) (res *http.Response, err error) {
// 	var requestByte []byte
// 	var requestBody *bytes.Reader
// 	if body == nil {
// 		requestBody = bytes.NewReader(nil)
// 	} else {
// 		requestByte, _ = json.Marshal(body)
// 		requestBody = bytes.NewReader(requestByte)
// 	}
// 	req, err := http.NewRequest(api.Method, viper.GetString("ebarimt.endpoint")+api.Url+ext, requestBody)
// 	if err != nil {
// 		err = errors.New("НӨАТ хүсэтийг боловсруулж чадсангүй")
// 		return
// 	}
// 	req.Header.Add("Content-Type", helper.HttpContent)
// 	res, err = http.DefaultClient.Do(req)
// 	return
// }

// StatusSuccess reports whether an HTTP status is 2xx.
//
// Spelled out against the raw status code rather than resty's own helper on
// purpose: that helper is IsSuccess in resty v3 beta and IsStatusSuccess in
// v3 rc, so calling it pins this module to one resty release and drags every
// consumer onto the same one. Consumers embed qPay alongside other SDKs still
// built against the beta, and a rename inside a dependency is not a reason to
// force a migration on them. StatusCode() exists in both.
func StatusSuccess(code int) bool {
	return code >= 200 && code < 300
}
