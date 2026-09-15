package qpay_v2

import (
	"encoding/json"
	"testing"
)

// A real, unedited GET /v2/invoice/{id} body captured from production qPay
// (merchant.qpay.mn) on 2026-09-15 for a 10 MNT probe invoice, since cancelled.
//
// It is the reason Amount exists: total_amount arrives quoted, gross_amount,
// tax_amount, surcharge_amount and discount_amount arrive as bare numbers, and
// minimum_amount/maximum_amount arrive as null — all in this one response. With
// those fields typed as string, decoding failed on the first number with
// "cannot unmarshal number into Go struct field
// QpayInvoiceGetResponse.gross_amount of type string", and GetInvoice returned
// an empty struct. An invented fixture would not have caught this; only a real
// response shows qPay mixing three JSON types for money.
const realInvoiceGetBody = `{"invoice_id":"ea98db5c-d1c4-4c3c-9d07-0d2c0c754759","invoice_status":"OPEN","sender_invoice_no":"rawprobe","sender_branch_code":null,"sender_branch_data":null,"sender_staff_code":null,"sender_staff_data":null,"sender_terminal_code":null,"sender_terminal_data":null,"invoice_description":"raw shape probe - do not pay","invoice_due_date":null,"enable_expiry":false,"expiry_date":"2026-09-15T08:16:01.273Z","allow_partial":false,"minimum_amount":null,"allow_exceed":false,"maximum_amount":null,"total_amount":"10.00","gross_amount":10,"tax_amount":0,"surcharge_amount":0,"discount_amount":0,"callback_url":"https://api-vms.mtm.mn/api/v1/webhook/qpay/","note":null,"lines":[{"tax_product_code":null,"line_description":"raw shape probe - do not pay","line_quantity":"1.00","line_unit_price":"10.00","note":"","discounts":[],"surcharges":[],"taxes":[]}],"transactions":[],"inputs":[]}`

func TestQpayInvoiceGetResponse_decodesRealProductionBody(t *testing.T) {
	var got QpayInvoiceGetResponse
	if err := json.Unmarshal([]byte(realInvoiceGetBody), &got); err != nil {
		t.Fatalf("a real qPay invoice-get body must decode: %v", err)
	}

	for _, tc := range []struct {
		field string
		got   Amount
		want  Amount
	}{
		{"total_amount (string in the wire body)", got.TotalAmount, "10.00"},
		{"gross_amount (number in the wire body)", got.GrossAmount, "10"},
		{"tax_amount (number)", got.TaxAmount, "0"},
		{"surcharge_amount (number)", got.SurchargeAmount, "0"},
		{"discount_amount (number)", got.DiscountAmount, "0"},
		{"minimum_amount (null)", got.MinimumAmount, ""},
		{"maximum_amount (null)", got.MaximumAmount, ""},
	} {
		if tc.got != tc.want {
			t.Errorf("%s = %q, want %q", tc.field, tc.got, tc.want)
		}
	}

	if got.InvoiceStatus != "OPEN" {
		t.Errorf("invoice_status = %q, want OPEN", got.InvoiceStatus)
	}
	if len(got.Lines) != 1 || got.Lines[0].LineUnitPrice != "10.00" {
		t.Errorf("nested line amounts did not decode: %+v", got.Lines)
	}
}
