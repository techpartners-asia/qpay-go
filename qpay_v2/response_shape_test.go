package qpay_v2

import (
	"encoding/json"
	"testing"
)

func TestInvoiceGetResponseAcceptsDecimalStringAmounts(t *testing.T) {
	var res QpayInvoiceGetResponse
	err := json.Unmarshal([]byte(`{
		"invoice_id": "c565079c-11e6-45c2-a802-fc0f2f784993",
		"invoice_status": "OPEN",
		"sender_invoice_no": "SMOKE_SIMPLE_20260505141923",
		"gross_amount": "100.00",
		"discount_amount": "0.00",
		"surcharge_amount": "0.00",
		"tax_amount": "0.00",
		"total_amount": "100.00",
		"minimum_amount": null,
		"maximum_amount": null
	}`), &res)
	if err != nil {
		t.Fatalf("failed to unmarshal invoice get response: %v", err)
	}
	if res.TotalAmount != "100.00" {
		t.Fatalf("unexpected total amount: %s", res.TotalAmount)
	}
	if res.MinimumAmount != "" {
		t.Fatalf("expected null minimum amount to decode as empty string, got %s", res.MinimumAmount)
	}
}

func TestPaymentGetResponseUsesExcelP2PExampleShape(t *testing.T) {
	var res QpayPaymentGetResponse
	err := json.Unmarshal([]byte(`{
		"payment_id": "493622150113497",
		"payment_status": "PAID",
		"payment_fee": "1.00",
		"payment_amount": "100.00",
		"payment_currency": "MNT",
		"payment_date": "2022-03-11T05:57:47.336Z",
		"payment_wallet": "0fc9b71c-cd87-4ffd-9cac-2279ebd9deb0",
		"object_type": "INVOICE",
		"object_id": "d50f49f2-9032-4a74-8929-530531f28f63",
		"next_payment_date": null,
		"next_payment_datetime": null,
		"transaction_type": "P2P",
		"card_transactions": [],
		"p2p_transactions": [
			{
				"transaction_bank_code": "050000",
				"account_bank_code": "050000",
				"account_bank_name": "Хаан банк",
				"account_number": "50*******",
				"status": "SUCCESS",
				"amount": "99.00",
				"currency": "MNT",
				"settlement_status": "SETTLED"
			}
		]
	}`), &res)
	if err != nil {
		t.Fatalf("failed to unmarshal payment get response: %v", err)
	}
	if res.PaymentAmount != "100.00" {
		t.Fatalf("unexpected payment amount: %s", res.PaymentAmount)
	}
	if res.TransactionType != "P2P" {
		t.Fatalf("unexpected transaction type: %s", res.TransactionType)
	}
	if len(res.P2PTransactions) != 1 {
		t.Fatalf("expected one p2p transaction, got %d", len(res.P2PTransactions))
	}
	if res.P2PTransactions[0].Amount != "99.00" {
		t.Fatalf("unexpected p2p amount: %s", res.P2PTransactions[0].Amount)
	}
}

func TestPaymentCheckResponseUsesExcelRowShape(t *testing.T) {
	var res QpayPaymentCheckResponse
	err := json.Unmarshal([]byte(`{
		"count": 1,
		"paid_amount": 100,
		"rows": [
			{
				"payment_id": "8e25b4d5-fe5a-4d0f-b050-def68a82aaad",
				"payment_status": "PAID",
				"payment_amount": "100.00",
				"trx_fee": "1.00",
				"payment_currency": "MNT",
				"payment_wallet": "qPay wallet",
				"payment_type": "CARD",
				"next_payment_date": null,
				"next_payment_datetime": null,
				"card_transactions": [
					{
						"card_type": "UNIONPAY",
						"is_cross_border": false,
						"amount": "100.00",
						"currency": "MNT",
						"date": "2022-03-11T06:23:48.586Z",
						"status": "SUCCESS",
						"settlement_status": "PENDING",
						"settlement_status_date": "2022-03-11T06:23:48.587Z"
					}
				],
				"p2p_transactions": []
			}
		]
	}`), &res)
	if err != nil {
		t.Fatalf("failed to unmarshal payment check response: %v", err)
	}
	if len(res.Rows) != 1 {
		t.Fatalf("expected one row, got %d", len(res.Rows))
	}
	row := res.Rows[0]
	if row.TrxFee != "1.00" {
		t.Fatalf("unexpected trx fee: %s", row.TrxFee)
	}
	if row.PaymentType != "CARD" {
		t.Fatalf("unexpected payment type: %s", row.PaymentType)
	}
	if len(row.CardTransactions) != 1 {
		t.Fatalf("expected one card transaction, got %d", len(row.CardTransactions))
	}
	if row.CardTransactions[0].Amount != "100.00" {
		t.Fatalf("unexpected card amount: %s", row.CardTransactions[0].Amount)
	}
}
