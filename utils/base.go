package utils

import (
	"fmt"
	"strconv"
)

// StrToUint String to Uint parser
func StrToUint(value string) (uint, error) {
	u64, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, err
	}
	result := uint(u64)
	return result, nil
}

func AppendAsString(args ...interface{}) string {
	appendedStr := ""
	for _, arg := range args {
		appendedStr = appendedStr + fmt.Sprintf("%v", arg)
	}

	return appendedStr
}

// GetValidString renders any decoded JSON value as a string.
// Non-string values (JSON numbers decode to float64, booleans to bool)
// are formatted rather than type-asserted, so they cannot panic.
func GetValidString(source interface{}) string {
	switch v := source.(type) {
	case nil:
		return ""
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// GetValidFloat reads any decoded JSON value as a float64.
// qPay sends monetary amounts as both JSON strings ("100.00") and JSON
// numbers depending on the endpoint, so both are accepted. Anything that
// cannot be interpreted as a number yields 0.
func GetValidFloat(source interface{}) float64 {
	switch v := source.(type) {
	case nil:
		return 0
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		num, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0
		}
		return num
	default:
		return 0
	}
}
