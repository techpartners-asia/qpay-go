package qpay_v2

type (
	// qpayLoginResponse [Нэвтрэх хариу]
	// See: https://developer.qpay.mn/#auth-token
	qpayLoginResponse struct {
		TokenType        string `json:"token_type"`         // Токены төрөл (Bearer)
		RefreshToken     string `json:"refresh_token"`       // Шинэчлэх токен
		RefreshExpiresIn int    `json:"refresh_expires_in"` // Шинэчлэх токены хүчинтэй хугацаа (сек)
		AccessToken      string `json:"access_token"`       // Хандалтын токен
		ExpiresIn        int    `json:"expires_in"`         // Хандалтын токены хүчинтэй хугацаа (сек)
		Scope            string `json:"scope"`               // Хандах хүрээ
		NotBeforePolicy  string `json:"not-before-policy"`  // Бодлого
		SessionState     string `json:"session_state"`      // Сессийн төлөв
	}

	// QPayCreateInvoiceInput [Нэхэмжлэх үүсгэх оролтын өгөгдөл]
	// SDK-ийн CreateInvoice функцэд дамжуулах бүтэц.
	QPayCreateInvoiceInput struct {
		SenderCode       string               // qpay-ээс өгсөн нэхэмжлэхийн код (invoice_code)
		SenderBranchCode string               // Байгууллагын салбарын код
		ReceiverCode     string               // Байгууллагын терминал/хүлээн авагчийн код
		ReceiverData     *InvoiceReceiverData // Хүлээн авагчийн нэмэлт мэдээлэл (заавал биш)
		Description      string               // Нэхэмжлэлийн утга/тайлбар
		Amount           int64                // Мөнгөн дүн (бүхэл тоогоор)
		CallbackParam    map[string]string    // Төлбөр төлөгдсөний дараа дуудагдах URL-д нэмэгдэх параметрүүд
		Lines            []*QpayLine          // Нэхэмжлэлийн мөрүүд (заавал биш)
		Note             string               // Тэмдэглэл (заавал биш)
	}

	// InvoiceReceiverData [Нэхэмжлэл хүлээн авагчийн мэдээлэл]
	// See: https://developer.qpay.mn/#invoice-Create
	InvoiceReceiverData struct {
		Register string `json:"register"` // Регистрийн дугаар
		Name     string `json:"name"`     // Хүлээн авагчийн нэр
		Email    string `json:"email"`    // И-мэйл хаяг
		Phone    string `json:"phone"`    // Утасны дугаар
	}

	// QpayAdjustment [Хөнгөлөлт, нэмэгдэл, татвар]
	// See: https://developer.qpay.mn/#invoice-Create
	QpayAdjustment struct {
		Code        string `json:"adjustment_code"` // Код (VAT, NONE г.м)
		Description string `json:"description"`     // Тайлбар
		Amount      int64  `json:"amount"`          // Мөнгөн дүн
		Note        string `json:"note"`            // Тэмдэглэл
	}

	// QpayLine [Нэхэмжлэлийн мөр]
	// See: https://developer.qpay.mn/#invoice-Create
	QpayLine struct {
		TaxProductCode  string            `json:"tax_product_code"` // Татварын барааны код
		LineDescription string            `json:"line_description"` // Мөрийн тайлбар
		LineQuantity    string            `json:"line_quantity"`    // Тоо ширхэг
		LineUnitPrice   string            `json:"line_unit_price"`   // Нэгж үнэ
		Note            string            `json:"note"`             // Тэмдэглэл
		Discounts       []*QpayAdjustment `json:"discounts"`        // Хөнгөлөлтүүд
		Surcharges      []*QpayAdjustment `json:"surcharges"`       // Нэмэгдлүүд
		Taxes           []*QpayAdjustment `json:"taxes"`            // Татварууд
	}

	// QPaySimpleInvoiceRequest [Нэхэмжлэх үүсгэх хүсэлт]
	// See: https://developer.qpay.mn/#invoice-Create
	QPaySimpleInvoiceRequest struct {
		InvoiceCode         string               `json:"invoice_code"`           // qpay-ээс өгсөн нэхэмжлэхийн код
		SenderInvoiceCode   string               `json:"sender_invoice_no"`      // Байгууллагаас үүсгэх давтагдашгүй дугаар
		SenderBranchCode    string               `json:"sender_branch_code"`     // Байгууллагын салбарын код
		InvoiceReceiverCode string               `json:"invoice_receiver_code"`  // Хэрэглэгчийн ID/Код
		InvoiceReceiverData *InvoiceReceiverData `json:"invoice_receiver_data"` // Хэрэглэгчийн мэдээлэл
		InvoiceDescription  string               `json:"invoice_description"`    // Нэхэмжлэлийн утга
		Amount              int64                `json:"amount"`                 // Нийт дүн
		CallbackUrl         string               `json:"callback_url"`           // Төлбөрийн хариу авах URL
		InvoiceDueDate      string               `json:"invoice_due_date"`       // Хүчинтэй хугацаа (YYYY-MM-DD HH:mm:ss)
		AllowPartial        bool                 `json:"allow_partial"`          // Хувааж төлөхийг зөвшөөрөх эсэх
		MinimumAmount       int64                `json:"minimum_amount"`         // Хамгийн бага төлөх дүн
		AllowExceed         bool                 `json:"allow_exceed"`           // Илүү төлөлт зөвшөөрөх эсэх
		MaximumAmount       int64                `json:"maximum_amount"`         // Хамгийн их төлөх дүн
		Note                string               `json:"note"`                   // Тэмдэглэл
		Lines               []*QpayLine          `json:"lines"`                  // Нэхэмжлэлийн мөрүүд
	}

	// QPaySimpleInvoiceResponse [Нэхэмжлэх үүсгэх хариу]
	// See: https://developer.qpay.mn/#invoice-Create
	QPaySimpleInvoiceResponse struct {
		InvoiceID    string      `json:"invoice_id"`    // Нэхэмжлэлийн ID
		QpayShortUrl string      `json:"qPay_shortUrl"` // QR холбоос (Shortcut)
		QrText       string      `json:"qr_text"`       // QR текст утга
		QrImage      string      `json:"qr_image"`      // QR зураг (Base64)
		Urls         []*Deeplink `json:"urls"`           // Банкны аппликейшн линкүүд
	}

	// Deeplink [Банкны аппликейшн холбоос]
	Deeplink struct {
		Name        string `json:"name"`        // Банкны нэр
		Description string `json:"description"` // Тайлбар
		Logo        string `json:"logo"`        // Лого (Base64/URL)
		Link        string `json:"link"`        // Банкны апп руу үсрэх холбоос
	}

	// QpayInvoiceGetResponse [Нэхэмжлэлийн мэдээлэл харах хариу]
	// See: https://developer.qpay.mn/#invoice-Get
	QpayInvoiceGetResponse struct {
		InvoiceID          string             `json:"invoice_id"`          // Нэхэмжлэлийн ID
		InvoiceStatus      string             `json:"invoice_status"`      // Нэхэмжлэлийн төлөв (OPEN, CLOSED, CANCELLED)
		SenderInvoiceNo    string             `json:"sender_invoice_no"`   // Байгууллагын нэхэмжлэлийн дугаар
		InvoiceDescription string             `json:"invoice_description"` // Нэхэмжлэлийн утга
		GrossAmount        int64              `json:"gross_amount"`        // Үндсэн дүн
		DiscountAmount     int64              `json:"discount_amount"`     // Хөнгөлөлтийн дүн
		SurchargeAmount    int64              `json:"surcharge_amount"`    // Нэмэгдлийн дүн
		TaxAmount          int64              `json:"tax_amount"`          // Татварын дүн
		TotalAmount        int64              `json:"total_amount"`        // Эцсийн төлөх дүн
		InvoiceDueDate     string             `json:"invoice_due_date"`    // Дуусах хугацаа
		ExpiryDate         string             `json:"expiry_date"`         // Хүчингүй болох огноо
		CallbackUrl        string             `json:"callback_url"`        // Хариу авах URL
		Note               string             `json:"note"`                // Тэмдэглэл
		Lines              []*QpayLine        `json:"lines"`               // Мөрүүд
		Transactions       []*QpayTransaction `json:"transactions"`        // Гүйлгээнүүд
		Inputs             []*QpayInput       `json:"inputs"`              // Бусад оролтууд
	}

	// QpayInput [Оролтын талбар]
	QpayInput struct {
		ID    string `json:"id"`    // Оролтын ID
		Name  string `json:"name"`  // Оролтын нэр
		Type  string `json:"type"`  // Төрөл
		Label string `json:"label"` // Шошго (Харагдах нэр)
		Value string `json:"value"` // Утга
	}

	// QpayTransaction [Гүйлгээний бүртгэл]
	// See: https://developer.qpay.mn/#payment-get
	QpayTransaction struct {
		BankCode             string `json:"bank_code"`               // Банкны код
		TransactionID        string `json:"transaction_id"`          // QPay гүйлгээний дугаар
		TransactionNo        string `json:"transaction_no"`          // Банкны гүйлгээний дугаар
		TransactionDate      string `json:"transaction_date"`        // Гүйлгээ хийгдсэн огноо
		TransactionAmount    string `json:"transaction_amount"`      // Гүйлгээний дүн
		TransactionCurrency  string `json:"transaction_currency"`    // Валют
		AccountName          string `json:"account_name"`            // Дансны нэр
		AccountNumber        string `json:"account_number"`          // Дансны дугаар
		AccountBankCode      string `json:"account_bank_code"`       // Дансны банкны код
		Description          string `json:"description"`             // Гүйлгээний утга
		Status               string `json:"status"`                  // Төлөв
		PaymentID            string `json:"payment_id"`              // Төлбөрийн ID
		SettlementStatus     string `json:"settlement_status"`       // Тооцоо нийлсэн төлөв
	}

	// QpayPaymentCheckRequest [Төлбөр шалгах хүсэлт]
	// See: https://developer.qpay.mn/#payment-check
	QpayPaymentCheckRequest struct {
		ObjectType string     `json:"object_type"` // Төрөл (INVOICE, QR, ITEM)
		ObjectID   string     `json:"object_id"`   // Харгалзах ID
		Offset     QpayOffset `json:"offset"`      // Хуудаслалт
	}

	// QpayOffset [Хуудаслалт]
	QpayOffset struct {
		PageNumber int64 `json:"page_number"` // Хуудасны дугаар
		PageLimit  int64 `json:"page_limit"`  // Нэг хуудас дахь мөрийн тоо
	}

	// QpayPaymentCheckResponse [Төлбөр шалгах хариу]
	// See: https://developer.qpay.mn/#payment-check
	QpayPaymentCheckResponse struct {
		Count      int64      `json:"count"`       // Гүйлгээний тоо
		PaidAmount int64      `json:"paid_amount"` // Нийт төлөгдсөн дүн
		Rows       []*QpayRow `json:"rows"`        // Гүйлгээний жагсаалт
	}

	// QpayRow [Гүйлгээний мэдээлэл]
	// See: https://developer.qpay.mn/#payment-check
	QpayRow struct {
		PaymentID       string `json:"payment_id"`       // Төлбөрийн ID
		PaymentStatus   string `json:"payment_status"`   // Төлөв (NEW, PAID, FAILED, REFUNDED)
		PaymentDate     string `json:"payment_date"`     // Төлөгдсөн хугацаа
		PaymentFee      string `json:"payment_fee"`      // Шимтгэлийн дүн
		PaymentAmount   string `json:"payment_amount"`   // Төлөгдсөн дүн
		PaymentCurrency string `json:"payment_currency"` // Валют
		PaymentWallet   string `json:"payment_wallet"`   // Ашигласан воллет
		TransactionType string `json:"transaction_type"` // Төрөл (P2P, CARD)
	}

	// QpayPaymentCancelRequest [Төлбөр цуцлах хүсэлт]
	// See: https://developer.qpay.mn/#payment-cancel
	QpayPaymentCancelRequest struct {
		CallbackUrl string `json:"callback_url"` // Каллбак URL
		Note        string `json:"note"`         // Тэмдэглэл / Шалтгаан
	}

	// QpayPaymentListRequest [Төлбөрийн жагсаалт авах хүсэлт]
	// See: https://developer.qpay.mn/#payment-list
	QpayPaymentListRequest struct {
		MerchantID           string     `json:"merchant_id"`            // Мерчантын ID
		MerchantBranchCode   string     `json:"merchant_branch_code"`   // Салбарын код
		MerchantTerminalCode string     `json:"merchant_terminal_code"` // Терминалын код
		MerchantStaffCode    string     `json:"merchant_staff_code"`    // Ажилтны код
		Offset               QpayOffset `json:"offset"`                 // Хуудаслалт
	}
)
