package qpay_v2

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newEbarimtMockServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/token" {
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(qpayLoginResponse{
				TokenType:        "bearer",
				AccessToken:      "test-access-token",
				RefreshToken:     "test-refresh-token",
				ExpiresIn:        time.Now().Add(24 * time.Hour).Unix(),
				RefreshExpiresIn: time.Now().Add(48 * time.Hour).Unix(),
				Scope:            "get_token",
				SessionState:     "test",
			})
			if err != nil {
				t.Errorf("failed to encode auth response: %v", err)
			}
			return
		}

		handler(w, r)
	}))
}

func TestCreateInvoiceUsesExcelSimpleRequestExample(t *testing.T) {
	var invoiceCalled bool
	srv := newEbarimtMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/invoice" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		invoiceCalled = true

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
		}
		payload := string(body)
		for _, want := range []string{
			`"invoice_code":"TEST_INVOICE"`,
			`"sender_invoice_no":"1234567"`,
			`"invoice_receiver_code":"terminal"`,
			`"invoice_description":"test"`,
			`"sender_branch_code":"SALBAR1"`,
			`"amount":100`,
			`"callback_url":"https://bd5492c3ee85.ngrok.io/payments?payment_id=1234567"`,
		} {
			if !strings.Contains(payload, want) {
				t.Errorf("payload missing %s in %s", want, payload)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"invoice_id":"d50f49f2-9032-4a74-8929-530531f28f63",
			"qr_text":"0002010102121531279404962794049600022310027138152045734530349654031005802MN5904TEST6011Ulaanbaatar6244010712345670504test0721qWlrS8_zUpplFJmmfBGXc6304C66D",
			"qr_image":"iVBORw0KGgoAAAANSUhEUgAAASwAAAEsCAY",
			"qPay_shortUrl":"https://s.qpay.mn/z1lKnIO5T",
			"urls":[
				{
					"name":"Khan bank",
					"description":"Хаан банк",
					"logo":"https://qpay.mn/q/logo/khanbank.png",
					"link":"khanbank://q?qPay_QRcode=000201"
				}
			]
		}`))
	})
	defer srv.Close()

	q := newTestQPay(srv.URL)
	q.invoiceCode = "TEST_INVOICE"
	q.callback = "https://bd5492c3ee85.ngrok.io/payments?payment_id=1234567"

	res, err := q.CreateInvoice(QPayCreateInvoiceInput{
		SenderInvoiceNo:     "1234567",
		InvoiceReceiverCode: "terminal",
		InvoiceDescription:  "test",
		SenderBranchCode:    "SALBAR1",
		Amount:              100,
	})
	if err != nil {
		t.Fatalf("CreateInvoice failed: %v", err)
	}
	if !invoiceCalled {
		t.Fatal("expected /invoice to be called")
	}
	if res.InvoiceID != "d50f49f2-9032-4a74-8929-530531f28f63" {
		t.Fatalf("unexpected invoice id: %s", res.InvoiceID)
	}
	if res.QpayShortUrl != "https://s.qpay.mn/z1lKnIO5T" {
		t.Fatalf("unexpected short url: %s", res.QpayShortUrl)
	}
	if len(res.Urls) != 1 {
		t.Fatalf("expected urls to populate Urls, got %d", len(res.Urls))
	}
}

func TestCreateEbarimtInvoiceSendsV3InvoicePayload(t *testing.T) {
	var invoiceCalled bool
	srv := newEbarimtMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/invoice" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		invoiceCalled = true

		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-access-token" {
			t.Errorf("unexpected authorization header: %s", got)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
		}
		payload := string(body)
		for _, want := range []string{
			`"invoice_code":"TEST_EB_INVOICE"`,
			`"sender_invoice_no":"TEST_INVOICE_23"`,
			`"invoice_receiver_code":"23"`,
			`"sender_branch_code":"TEST_BRANCH"`,
			`"invoice_description":"Test invoice"`,
			`"callback_url":"https://example.com/callback"`,
			`"tax_type":"1"`,
			`"district_code":"0101"`,
			`"line_description":"Улаан буудайн үр"`,
			`"classification_code":"0111100"`,
			`"line_quantity":"1.00"`,
			`"line_unit_price":"1000.00"`,
			`"tax_code":"VAT"`,
			`"amount":"89.2857"`,
			`"tax_code":"CITY_TAX"`,
			`"amount":"17.8571"`,
			`"line_description":"Бусад төрлийн сорго будаа"`,
			`"classification_code":"0114200"`,
			`"amount":"90.91"`,
		} {
			if !strings.Contains(payload, want) {
				t.Errorf("payload missing %s in %s", want, payload)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"invoice_id":"d50f49f2-9032-4a74-8929-530531f28f63",
			"qr_text":"000201",
			"qPay_shortUrl":"https://s.qpay.mn/z1lKnIO5T",
			"urls":[
				{
					"name":"Khan bank",
					"description":"Хаан банк",
					"logo":"https://qpay.mn/q/logo/khanbank.png",
					"link":"khanbank://q?qPay_QRcode=000201"
				}
			]
		}`))
	})
	defer srv.Close()

	q := newTestQPay(srv.URL)
	q.callback = "https://example.com/callback"

	res, err := q.CreateEbarimtInvoice(QPayCreateEbarimtInvoiceInput{
		InvoiceCode:         "TEST_EB_INVOICE",
		SenderInvoiceNo:     "TEST_INVOICE_23",
		InvoiceReceiverCode: "23",
		SenderBranchCode:    "TEST_BRANCH",
		InvoiceDescription:  "Test invoice",
		TaxType:             QPayTaxTypeVAT,
		DistrictCode:        "0101",
		Lines: []*QPayEbarimtInvoiceLine{
			{
				TaxProductCode:     "",
				LineDescription:    "Улаан буудайн үр",
				LineQuantity:       "1.00",
				LineUnitPrice:      "1000.00",
				Note:               "TEST",
				ClassificationCode: "0111100",
				Taxes: []*QPayEbarimtTax{
					{
						TaxCode:     QPayTaxCodeVAT,
						Description: "НӨАТ",
						Amount:      "89.2857",
						Note:        "НӨАТ",
					},
					{
						TaxCode:     QPayTaxCodeCity,
						Description: "City tax",
						Amount:      "17.8571",
						Note:        "TEST",
					},
				},
			},
			{
				TaxProductCode:     "",
				LineDescription:    "Бусад төрлийн сорго будаа",
				LineQuantity:       "1.00",
				LineUnitPrice:      "1000.00",
				Note:               "TEST",
				ClassificationCode: "0114200",
				Taxes: []*QPayEbarimtTax{
					{
						TaxCode:     QPayTaxCodeVAT,
						Description: "НӨАТ",
						Amount:      "90.91",
						Note:        "НӨАТ",
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateEbarimtInvoice failed: %v", err)
	}
	if !invoiceCalled {
		t.Fatal("expected /invoice to be called")
	}
	if res.InvoiceID != "d50f49f2-9032-4a74-8929-530531f28f63" {
		t.Fatalf("unexpected invoice id: %s", res.InvoiceID)
	}
	if len(res.Urls) != 1 {
		t.Fatalf("expected urls to populate Urls, got %d", len(res.Urls))
	}
}

func TestCreateEbarimtInvoiceDisablesVatCalculationForFreeAndExemptTaxTypes(t *testing.T) {
	q := newTestQPay("https://qpay.test")

	for _, taxType := range []QPayTaxType{QPayTaxTypeNoVAT, QPayTaxTypeVATExempt} {
		req := q.newEbarimtInvoiceRequest(QPayCreateEbarimtInvoiceInput{TaxType: taxType})
		if req.CalculateVat == nil {
			t.Fatalf("expected calculate_vat=false for tax type %s", taxType)
		}
		if *req.CalculateVat {
			t.Fatalf("expected calculate_vat=false for tax type %s", taxType)
		}
	}

	req := q.newEbarimtInvoiceRequest(QPayCreateEbarimtInvoiceInput{TaxType: QPayTaxTypeVAT})
	if req.CalculateVat != nil {
		t.Fatal("expected calculate_vat to be omitted for VAT taxable products by default")
	}

	value := true
	req = q.newEbarimtInvoiceRequest(QPayCreateEbarimtInvoiceInput{
		TaxType:      QPayTaxTypeNoVAT,
		CalculateVat: &value,
	})
	if req.CalculateVat == nil || !*req.CalculateVat {
		t.Fatal("expected explicit calculate_vat override to be preserved")
	}
}

func TestCreateEbarimtSendsV3Payload(t *testing.T) {
	srv := newEbarimtMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ebarimt_v3/create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var payload QPayEbarimtCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if payload.PaymentID != "019276866891878" {
			t.Errorf("unexpected payment_id: %s", payload.PaymentID)
		}
		if payload.EbarimtReceiverType != QPayEbarimtReceiverCitizen {
			t.Errorf("unexpected receiver type: %s", payload.EbarimtReceiverType)
		}
		if payload.EbarimtReceiver != "88614450" {
			t.Errorf("unexpected receiver: %s", payload.EbarimtReceiver)
		}
		if payload.DistrictCode != "3505" {
			t.Errorf("unexpected district code: %s", payload.DistrictCode)
		}
		if payload.ClassificationCode != "0000010" {
			t.Errorf("unexpected classification code: %s", payload.ClassificationCode)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id":"ca48461c-0b85-438d-b8f4-8b46582a668c",
			"ebarimt_by":"QPAY",
			"ebarimt_receiver_type":"CITIZEN",
			"ebarimt_receiver":"88614450",
			"ebarimt_district_code":"3505",
			"merchant_branch_code":"BRANCH1",
			"merchant_staff_code":"online",
			"merchant_register_no":"5395305",
			"g_payment_id":"019276866891878",
			"paid_by":"P2P",
			"object_type":"INVOICE",
			"object_id":"18f4d9be-9ad7-42d4-95b9-c0d2f9e75900",
			"amount":"200.00",
			"vat_amount":"0.00",
			"city_tax_amount":"0.00",
			"ebarimt_qr_data":"13843170943750114352",
			"ebarimt_lottery":"HV 83198235",
			"barimt_status":"REGISTERED",
			"barimt_status_date":"2024-11-04T05:45:42.945Z",
			"ebarimt_receiver_phone":"88*144*0",
			"tax_type":"2",
			"merchant_tin":"30101065006",
			"ebarimt_receipt_id":"030101065006000090690000210005595",
			"status":true
		}`))
	})
	defer srv.Close()

	q := newTestQPay(srv.URL)

	res, err := q.CreateEbarimt(QPayEbarimtCreateInput{
		PaymentID:           "019276866891878",
		EbarimtReceiverType: QPayEbarimtReceiverCitizen,
		EbarimtReceiver:     "88614450",
		DistrictCode:        "3505",
		ClassificationCode:  "0000010",
	})
	if err != nil {
		t.Fatalf("CreateEbarimt failed: %v", err)
	}
	if res.BarimtStatus != "REGISTERED" {
		t.Fatalf("unexpected barimt status: %s", res.BarimtStatus)
	}
	if res.EbarimtQRData == "" {
		t.Fatal("expected ebarimt QR data")
	}
	if res.EbarimtReceiptID != "030101065006000090690000210005595" {
		t.Fatalf("unexpected receipt id: %s", res.EbarimtReceiptID)
	}
}

func TestCancelEbarimtUsesPaymentID(t *testing.T) {
	srv := newEbarimtMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ebarimt_v3/019276866891878" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id":"f699a4be-4cc7-4153-8dbc-942c80b61c6d",
			"ebarimt_by":"QPAY",
			"ebarimt_receiver_type":"CITIZEN",
			"ebarimt_receiver":"88614450",
			"ebarimt_district_code":"27",
			"merchant_branch_code":"Online",
			"merchant_register_no":"5395305",
			"g_payment_id":"019276866891878",
			"paid_by":"P2P",
			"object_type":"INVOICE",
			"object_id":"d50f49f2-9032-4a74-8929-530531f28f63",
			"amount":"100.00",
			"vat_amount":"9.09",
			"city_tax_amount":"0.00",
			"ebarimt_qr_data":"1048679142796755211072066810095357245730236961938408062673800276589422403457091041939370847562231566282076211053347122650907115463645280560716980518356003301201495188300421461150850576589935238730782470750003230891402302730235520970468521966821365715411520464352448558381298487948421264829397194559084948204313059935",
			"ebarimt_lottery":"CK 59114203",
			"barimt_status":"CANCELED",
			"barimt_status_date":"2024-05-14T01:56:49.724Z",
			"ebarimt_receiver_phone":"88*14*50",
			"tax_type":"1",
			"status":true
		}`))
	})
	defer srv.Close()

	q := newTestQPay(srv.URL)

	res, err := q.CancelEbarimt("019276866891878")
	if err != nil {
		t.Fatalf("CancelEbarimt failed: %v", err)
	}
	if res.BarimtStatus != "CANCELED" {
		t.Fatalf("unexpected barimt status: %s", res.BarimtStatus)
	}
	if res.GPaymentID != "019276866891878" {
		t.Fatalf("unexpected payment id: %s", res.GPaymentID)
	}
}
