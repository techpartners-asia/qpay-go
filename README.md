# qpay-go

Unofficial Go SDK for the [QPay](https://qpay.mn) payment gateway. Supports QPay V1, QPay V2, and QPay Quick APIs.

## Installation

```bash
go get github.com/techpartners-asia/qpay-go
```

**Requires Go 1.23+**

---

## Packages

| Package | Description |
|---|---|
| `qpay_v2` | QPay V2 — recommended for new integrations |
| `qpay_v1` | QPay V1 — legacy support |
| `qpay_quick` | QPay Quick — marketplace/sub-merchant API |

---

## QPay V2 (Recommended)

### Initialize

```go
import qpay "github.com/techpartners-asia/qpay-go/qpay_v2"

client, err := qpay.New(
    "YOUR_USERNAME",                // QPay username (client_id)
    "YOUR_PASSWORD",                // QPay password (client_secret)
    "https://merchant.qpay.mn/v2", // Production endpoint
    "https://yourapp.com/callback", // Callback base URL
    "YOUR_INVOICE_CODE",            // Invoice code assigned by QPay
    "YOUR_MERCHANT_ID",             // Merchant ID
)
if err != nil {
    log.Fatal(err)
}
```

**Sandbox endpoint:** `https://merchant-sandbox.qpay.mn/v2`

#### Custom HTTP client

Use `WithClient` to inject your own `resty.Client` with custom timeouts, TLS config, or middleware:

```go
import "resty.dev/v3"

httpClient := resty.New().SetTimeout(10 * time.Second)

client, err := qpay.New(
    "USERNAME", "PASSWORD", "ENDPOINT", "CALLBACK", "INVOICE_CODE", "MERCHANT_ID",
    qpay.WithClient(httpClient),
)
```

---

### Create Invoice

```go
invoice, _, err := client.CreateInvoice(qpay.QPayCreateInvoiceInput{
    SenderCode:       "SENDER-001",   // Your unique invoice/order number
    SenderBranchCode: "BRANCH-001",   // Your branch code (optional)
    ReceiverCode:     "RECEIVER-001", // Receiver identifier (optional)
    Description:      "Order #1234",
    Amount:           10000,          // Amount in MNT (integer)
    CallbackParam: map[string]string{
        "order_id": "1234",
    },
})
if err != nil {
    log.Fatal(err)
}

fmt.Println(invoice.InvoiceID)
fmt.Println(invoice.QrImage) // Base64 encoded QR image
fmt.Println(invoice.QrText)  // Raw QR text value
fmt.Println(invoice.Urls)    // Bank app deeplinks
```

### Get Invoice

```go
invoice, _, err := client.GetInvoice("INVOICE_ID")
if err != nil {
    log.Fatal(err)
}
fmt.Println(invoice.InvoiceStatus) // OPEN, CLOSED, CANCELLED
fmt.Println(invoice.TotalAmount)
```

### Check Payment

```go
result, _, err := client.CheckPayment("INVOICE_ID", 10, 1) // pageLimit, pageNumber
if err != nil {
    log.Fatal(err)
}

fmt.Println(result.Count)
fmt.Println(result.PaidAmount)

for _, row := range result.Rows {
    fmt.Println(row.PaymentID, row.PaymentStatus, row.PaymentAmount)
}
```

Payment statuses: `NEW`, `PAID`, `FAILED`, `REFUNDED`

### Cancel Invoice

```go
_, _, err := client.CancelInvoice("INVOICE_ID")
```

### Cancel Payment

```go
result, _, err := client.CancelPayment("INVOICE_ID", "PAYMENT_ID")
```

### Refund Payment

```go
_, _, err := client.RefundPayment("INVOICE_ID", "PAYMENT_ID")
```

### Get Payment

```go
payment, _, err := client.GetPayment("PAYMENT_ID")
```

---

### Token Management

The SDK handles authentication automatically. Tokens are cached and refreshed before expiry — you do not need to manage tokens manually. All methods are safe for concurrent use.

---

## QPay V1

```go
import qpay "github.com/techpartners-asia/qpay-go/qpay_v1"

client := qpay.New(
    "CLIENT_ID",
    "CLIENT_SECRET",
    "https://sandbox.qpay.mn/v1", // endpoint
    "https://yourapp.com/callback",
    "MERCHANT_ID",
    "TEMPLATE_ID",
    "BRANCH_ID",
    "POS_ID",
)

// Create invoice
invoice, err := client.CreateInvoice(qpay.QPayInvoiceCreateRequest{
    BillNo:      "ORDER-001",
    Date:        "2024-01-01",
    Description: "Order payment",
    Amount:      10000,
})

// Get invoice
invoice, err := client.GetInvoice("INVOICE_ID")

// Check payment
payment, err := client.CheckPayment("PAYMENT_ID")
```

---

## QPay Quick

QPay Quick is a marketplace API for platforms that onboard sub-merchants.

```go
import qpay "github.com/techpartners-asia/qpay-go/qpay_quick"

client := qpay.New(
    "USERNAME",
    "PASSWORD",
    "https://quickpay.qpay.mn", // endpoint
    "https://yourapp.com/callback",
    "INVOICE_CODE",
    "TERMINAL_ID",
)

// Register a company merchant
company, err := client.CreateCompany(qpay.QpayCompanyCreateRequest{
    Name:       "Example LLC",
    RegisterNo: "1234567",
    Phone:      "99001122",
    Email:      "info@example.mn",
    City:       "Ulaanbaatar",
    District:   "Bayanzurkh",
    Address:    "1st khoroo",
    MCCcode:    "5999",
})

// Register a person merchant
person, err := client.CreatePerson(qpay.QpayPersonCreateRequest{
    RegisterNo: "УУ12345678",
    FirstName:  "Bat",
    LastName:   "Bold",
    Phone:      "99001122",
    Email:      "bat@example.mn",
    City:       "Ulaanbaatar",
})

// Get a merchant
merchant, err := client.GetMerchant("MERCHANT_ID")

// List merchants
merchants, err := client.ListMerchant(qpay.QpayOffset{
    PageNumber: 1,
    PageLimit:  20,
})

// Create invoice for a sub-merchant
invoice, err := client.CreateInvoice(qpay.QpayInvoiceRequest{
    MerchantID:  "MERCHANT_ID",
    Amount:      10000,
    Currency:    "MNT",
    Description: "Order payment",
    CallbackUrl: "https://yourapp.com/callback",
})

// Get invoice
invoice, err := client.GetInvoice("INVOICE_ID")

// Check payment
payment, err := client.CheckPayment("INVOICE_ID")
fmt.Println(payment.InvoiceStatus) // OPEN, PAID, CLOSED
```

---

## Error Handling

All methods return a standard Go `error`. On HTTP errors, the error contains the raw response body from QPay.

```go
invoice, _, err := client.CreateInvoice(input)
if err != nil {
    log.Printf("QPay error: %v", err)
    return
}
```

---

## QPay API Reference

- [QPay V2 Developer Docs](https://developer.qpay.mn)
- [QPay Quick Developer Docs](https://developer.qpay.mn/quick)

---

## License

MIT
