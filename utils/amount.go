package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Amount is a qPay money field.
//
// qPay does not commit to one JSON type for these. A single real
// GET /v2/invoice/{id} response carries all three shapes at once:
//
//	"total_amount":"10.00",   // string
//	"gross_amount":10,        // number
//	"tax_amount":0,           // number
//	"minimum_amount":null,    // null
//
// Typing them all as string made every such response fail to decode with
// "cannot unmarshal number into Go struct field ... of type string", and
// because the SDK returns the zero value on a decode error the caller got an
// empty struct back rather than the amounts. Typing them all as a number would
// fail on the string half the same way.
//
// So Amount accepts whichever shape arrives and keeps the digits as written —
// no float conversion, because these are money and a round trip through
// float64 is how you lose a tugrik. null becomes the empty string, matching
// what the old string fields did with it.
type Amount string

// String returns the amount exactly as qPay sent it.
func (a Amount) String() string { return string(a) }

func (a *Amount) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)

	if bytes.Equal(b, []byte("null")) {
		*a = ""
		return nil
	}

	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return fmt.Errorf("qpay: amount as string: %w", err)
		}
		*a = Amount(s)
		return nil
	}

	// json.Number keeps the literal text, so 10, 10.00 and 1e2 survive as
	// written instead of being normalised through a float.
	var n json.Number
	if err := json.Unmarshal(b, &n); err != nil {
		return fmt.Errorf("qpay: amount %q is neither a string nor a number: %w", b, err)
	}
	*a = Amount(n.String())
	return nil
}

// MarshalJSON always emits a JSON string, which is byte-for-byte what these
// fields produced when they were plain strings — several of them sit in
// request bodies, and this change must not alter a single outgoing byte.
func (a Amount) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(a))
}
