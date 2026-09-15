package utils

import (
	"encoding/json"
	"testing"
)

// qPay sends money as a string, a number, or null — sometimes all three in one
// response. Each shape must decode, and the digits must survive exactly as
// written: these are tugriks, and normalising them through a float is how you
// lose one.
func TestAmount_UnmarshalJSON(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want Amount
	}{
		{"quoted decimal", `"10.00"`, "10.00"},
		{"bare integer", `10`, "10"},
		{"bare zero", `0`, "0"},
		{"bare decimal", `10.00`, "10.00"},
		{"null becomes empty", `null`, ""},
		{"empty string stays empty", `""`, ""},
		{"large value keeps every digit", `123456789012345`, "123456789012345"},
		{"quoted large value", `"123456789012345.99"`, "123456789012345.99"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got Amount
			if err := json.Unmarshal([]byte(tc.in), &got); err != nil {
				t.Fatalf("Unmarshal(%s): %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("Unmarshal(%s) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestAmount_UnmarshalJSON_rejectsNonScalar(t *testing.T) {
	for _, in := range []string{`{}`, `[]`, `true`} {
		var got Amount
		if err := json.Unmarshal([]byte(in), &got); err == nil {
			t.Errorf("Unmarshal(%s) should fail, got %q", in, got)
		}
	}
}

// Several of these fields sit in request bodies (QPayEbarimtInvoiceLine,
// QPayEbarimtTax). Marshalling must stay byte-for-byte what a plain string
// produced, or this change silently alters outgoing invoices.
func TestAmount_MarshalJSON_isByteIdenticalToString(t *testing.T) {
	for _, s := range []string{"", "0", "10.00", "123456789012345.99"} {
		gotAmount, err := json.Marshal(Amount(s))
		if err != nil {
			t.Fatalf("marshal Amount(%q): %v", s, err)
		}
		gotString, err := json.Marshal(s)
		if err != nil {
			t.Fatalf("marshal string(%q): %v", s, err)
		}
		if string(gotAmount) != string(gotString) {
			t.Errorf("Amount(%q) marshalled to %s, plain string to %s", s, gotAmount, gotString)
		}
	}
}
