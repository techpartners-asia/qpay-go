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
		SenderCode           string                    // qpay-ээс өгсөн нэхэмжлэхийн код (invoice_code)
		SenderBranchCode     string                    // Байгууллагын салбарын код
		SenderBranchData     *SenderBranchData         // Салбарын мэдээлэл (заавал биш)
		SenderTerminalCode   string                    // Терминалын код
		SenderTerminalData   *SenderTerminalData       // Терминалын мэдээлэл (заавал биш)
		SenderStaffCode      string                    // Ажилтны код
		SenderStaffData      interface{}               // Ажилтны мэдээлэл (заавал биш)
		ReceiverCode         string                    // Хэрэглэгчийн ID/Код
		ReceiverData         *InvoiceReceiverData      // Хэрэглэгчийн мэдээлэл (заавал биш)
		Description          string                    // Нэхэмжлэлийн утга/тайлбар
		Amount               int64                     // Мөнгөн дүн (бүхэл тоогоор)
		CallbackParam        map[string]string         // URL-д нэмэгдэх параметрүүд
		Note                 string                    // Тэмдэглэл (заавал биш)
		Lines                []*QpayLineRequest        // Нэхэмжлэлийн мөрүүд (заавал биш)
		InvoiceDueDate       string                    // Төлөх эцсийн хугацаа (YYYY-MM-DD HH:mm:ss)
		ExpiryDate           string                    // Дуусах хугацаа (YYYY-MM-DD HH:mm:ss)
		EnableExpiry         bool                      // Дуусах хугацаа ашиглах эсэх
		AllowPartial         bool                      // Хувааж төлөх зөвшөөрөх
		MinimumAmount        int64                     // Хамгийн бага төлөх дүн
		AllowExceed          bool                      // Илүү төлөлт зөвшөөрөх
		MaximumAmount        int64                     // Хамгийн их төлөх дүн
		CalculateVat         bool                      // НӨАТ тооцох эсэх
		TaxCustomerCode      string                    // ИБаримт үүсгүүлэх байгууллага/хэрэглэгчийн регистр
		LineTaxCode          string                    // БТҮК код (Мөр хоосон үед ашиглана)
		Transactions         []*QpayTransactionRequest // Гүйлгээний мэдээлэл (Данс тохируулах)
	}

	// QpayAddress [Хаягийн мэдээлэл]
	QpayAddress struct {
		City      string `json:"city,omitempty"`      // Хот
		District  string `json:"district,omitempty"`  // Дүүрэг
		Street    string `json:"street,omitempty"`    // Гудамж
		Building  string `json:"building,omitempty"`  // Барилга
		Address   string `json:"address,omitempty"`   // Хаяг
		Zipcode   string `json:"zipcode,omitempty"`   // Зип код
		Longitude string `json:"longitude,omitempty"` // Уртраг
		Latitude  string `json:"latitude,omitempty"`  // Өргөрөг
	}

	// SenderBranchData [Салбарын мэдээлэл]
	SenderBranchData struct {
		Register string       `json:"register,omitempty"` // Салбарын регистр
		Name     string       `json:"name,omitempty"`     // Салбарын нэр
		Email    string       `json:"email,omitempty"`    // И-мэйл хаяг
		Phone    string       `json:"phone,omitempty"`    // Утасны дугаар
		Address  *QpayAddress `json:"address,omitempty"`  // Хаяг
	}

	// SenderTerminalData [Терминалын мэдээлэл]
	SenderTerminalData struct {
		Name string `json:"name"` // Терминалын нэр
	}

	// InvoiceReceiverData [Нэхэмжлэл хүлээн авагчийн мэдээлэл]
	// See: https://developer.qpay.mn/#invoice-Create
	InvoiceReceiverData struct {
		Register string       `json:"register,omitempty"` // Хэрэглэгчийн регистр
		Name     string       `json:"name,omitempty"`     // Нэр
		Email    string       `json:"email,omitempty"`    // И-мэйл хаяг
		Phone    string       `json:"phone,omitempty"`    // Утасны дугаар
		Note     string       `json:"note,omitempty"`     // Тэмдэглэл
		Address  *QpayAddress `json:"address,omitempty"`  // Хаяг
	}

	// QpayAdjustmentDiscount [Хөнгөлөлт]
	QpayAdjustmentDiscount struct {
		Code        string `json:"discount_code"` // Код
		Description string `json:"description"`   // Тайлбар
		Amount      int64  `json:"amount"`        // Мөнгөн дүн
		Note        string `json:"note,omitempty"` // Тэмдэглэл
	}

	// QpayAdjustmentSurcharge [Нэмэгдэл]
	QpayAdjustmentSurcharge struct {
		Code        string `json:"surcharge_code"` // Код
		Description string `json:"description"`     // Тайлбар
		Amount      int64  `json:"amount"`          // Мөнгөн дүн
		Note        string `json:"note,omitempty"`   // Тэмдэглэл
	}

	// QpayAdjustmentTax [Татвар]
	QpayAdjustmentTax struct {
		Code        string `json:"tax_code"`           // Код (CITY_TAX, VAT)
		Description string `json:"description"`         // Тайлбар
		Amount      int64  `json:"amount"`              // Мөнгөн дүн
		CityTax     int64  `json:"city_tax,omitempty"`  // Хотын татвар
		Note        string `json:"note,omitempty"`       // Тэмдэглэл
	}

	// QpayLineRequest [Нэхэмжлэлийн мөр - Хүсэлт илгээхэд ашиглана]
	// See: https://developer.qpay.mn/#invoice-Create
	QpayLineRequest struct {
		SenderProductCode string                     `json:"sender_product_code,omitempty"` // Байгууллагын барааны код
		TaxProductCode    string                     `json:"tax_product_code,omitempty"`    // БТҮК код
		LineDescription   string                     `json:"line_description"`              // Мөрийн тайлбар
		LineQuantity      int64                      `json:"line_quantity"`                 // Тоо ширхэг (Тоо)
		LineUnitPrice     int64                      `json:"line_unit_price"`               // Нэгж үнэ (Тоо)
		Note              string                     `json:"note,omitempty"`                // Тэмдэглэл
		Discounts         []*QpayAdjustmentDiscount  `json:"discounts,omitempty"`           // Хөнгөлөлтүүд
		Surcharges        []*QpayAdjustmentSurcharge `json:"surcharges,omitempty"`          // Нэмэгдлүүд
		Taxes             []*QpayAdjustmentTax       `json:"taxes,omitempty"`               // Татварууд
	}

	// QpayAccountRequest [Дансны мэдээлэл]
	QpayAccountRequest struct {
		AccountBankCode string `json:"account_bank_code"` // Банкны код
		AccountNumber   string `json:"account_number"`    // Дансны дугаар
		AccountName     string `json:"account_name"`      // Дансны нэр
		AccountCurrency string `json:"account_currency"`  // Валют (MNT)
	}

	// QpayTransactionRequest [Гүйлгээний мэдээлэл (Хүсэлт)]
	QpayTransactionRequest struct {
		Description string                `json:"description"`       // Гүйлгээний утга
		Amount      int64                 `json:"amount"`            // Мөнгөн дүн
		Accounts    []*QpayAccountRequest `json:"accounts,omitempty"` // Банкны данснууд
	}

	// QpayLineResponse [Нэхэмжлэлийн мөр - Хариу авахад ашиглана]
	// QPay API нь хариунд тоон утгуудыг текст ("string") байдлаар буцаадаг.
	QpayLineResponse struct {
		TaxProductCode  string                     `json:"tax_product_code"` // Татварын барааны код
		LineDescription string                     `json:"line_description"` // Мөрийн тайлбар
		LineQuantity    string                     `json:"line_quantity"`    // Тоо ширхэг (Текст)
		LineUnitPrice   string                     `json:"line_unit_price"`   // Нэгж үнэ (Текст)
		Note            string                     `json:"note"`             // Тэмдэглэл
		Discounts       []*QpayAdjustmentDiscount  `json:"discounts"`        // Хөнгөлөлтүүд
		Surcharges      []*QpayAdjustmentSurcharge `json:"surcharges"`       // Нэмэгдлүүд
		Taxes           []*QpayAdjustmentTax       `json:"taxes"`            // Татварууд
	}

	// QPaySimpleInvoiceRequest [Нэхэмжлэх үүсгэх хүсэлт]
	// See: https://developer.qpay.mn/#invoice-Create
	QPaySimpleInvoiceRequest struct {
		InvoiceCode         string                    `json:"invoice_code"`                    // qpay-ээс өгсөн нэхэмжлэхийн код
		SenderInvoiceNo     string                    `json:"sender_invoice_no"`               // Байгууллагаас үүсгэх дугаар
		SenderBranchCode    string                    `json:"sender_branch_code,omitempty"`    // Салбарын код
		SenderBranchData    *SenderBranchData         `json:"sender_branch_data,omitempty"`    // Салбарын мэдээлэл
		SenderTerminalCode  string                    `json:"sender_terminal_code,omitempty"`  // Терминалын код
		SenderTerminalData  *SenderTerminalData       `json:"sender_terminal_data,omitempty"`  // Терминалын мэдээлэл
		SenderStaffCode     string                    `json:"sender_staff_code,omitempty"`     // Ажилтны код
		SenderStaffData     interface{}               `json:"sender_staff_data,omitempty"`     // Ажилтны мэдээлэл
		InvoiceReceiverCode string                    `json:"invoice_receiver_code"`           // Хэрэглэгчийн ID/Код
		InvoiceReceiverData *InvoiceReceiverData      `json:"invoice_receiver_data,omitempty"` // Хэрэглэгчийн мэдээлэл
		InvoiceDescription  string                    `json:"invoice_description"`             // Нэхэмжлэлийн утга
		Amount              int64                     `json:"amount"`                          // Нийт дүн
		CallbackUrl         string                    `json:"callback_url"`                    // Хариу авах URL
		InvoiceDueDate      string                    `json:"invoice_due_date,omitempty"`       // Хүчинтэй хугацаа
		ExpiryDate          string                    `json:"expiry_date,omitempty"`           // Дуусах хугацаа
		EnableExpiry        bool                      `json:"enable_expiry"`                   // Дуусах хугацаа ашиглах
		AllowPartial        bool                      `json:"allow_partial"`                   // Хувааж төлөх
		MinimumAmount       interface{}               `json:"minimum_amount"`                  // Хамгийн бага төлөх дүн (null allowed)
		AllowExceed          bool                      `json:"allow_exceed"`                    // Илүү төлөлт
		MaximumAmount       interface{}               `json:"maximum_amount"`                  // Хамгийн их төлөх дүн (null allowed)
		CalculateVat        bool                      `json:"calculate_vat"`                   // НӨАТ тооцох
		Note                string                    `json:"note,omitempty"`                  // Тэмдэглэл
		Lines               []*QpayLineRequest        `json:"lines,omitempty"`                 // Нэхэмжлэлийн мөрүүд
		TaxCustomerCode     string                    `json:"tax_customer_code,omitempty"`     // ИБаримт регистр
		LineTaxCode         string                    `json:"line_tax_code,omitempty"`         // БТҮК код
		Transactions        []*QpayTransactionRequest `json:"transactions,omitempty"`          // Дансны тохиргоо
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
		EnableExpiry       bool               `json:"enable_expiry"`       // Дуусах хугацаа ашиглах
		AllowPartial       bool               `json:"allow_partial"`       // Хувааж төлөх
		AllowExceed        bool               `json:"allow_exceed"`        // Илүү төлөлт
		MinimumAmount      int64              `json:"minimum_amount"`      // Хамгийн бага дүн
		MaximumAmount      int64              `json:"maximum_amount"`      // Хамгийн их дүн
		SenderBranchCode   string             `json:"sender_branch_code"`  // Салбарын код
		CallbackUrl        string             `json:"callback_url"`        // Хариу авах URL
		Note               string             `json:"note"`                // Тэмдэглэл
		Lines              []*QpayLineResponse `json:"lines"`               // Мөрүүд (Хариунд текст байдлаар ирдэг)
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
		Count      int64      `json:"count"`       // Нийт гүйлгээний мөрийн тоо
		PaidAmount int64      `json:"paid_amount"` // Гүйлгээний дүн
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

	// QPayPaymentListInput [Төлбөрийн жагсаалт авах оролтын өгөгдөл]
	// SDK-ийн GetPaymentList функцэд дамжуулах бүтэц.
	QPayPaymentListInput struct {
		ObjectType   string // Обьектын төрөл (MERCHANT, INVOICE, QR)
		ObjectID     string // Обьектын ID
		BranchCode   string // Салбарын код
		TerminalCode string // Терминалын код
		StaffCode    string // Ажилтны код
		PageLimit    int64  // Нэг хуудас дахь мөрийн тоо
		PageNumber   int64  // Хуудасны дугаар
	}

	// QpayPaymentListRequest [Төлбөрийн жагсаалт авах хүсэлт]
	// See: https://developer.qpay.mn/#payment-list
	QpayPaymentListRequest struct {
		ObjectType           string     `json:"object_type"`            // MERCHANT, INVOICE, QR
		ObjectID             string     `json:"object_id"`              // Merchant ID, Invoice ID эсвэл QR код
		MerchantBranchCode   string     `json:"merchant_branch_code,omitempty"` // Салбарын код
		MerchantTerminalCode string     `json:"merchant_terminal_code,omitempty"` // Терминалын код
		MerchantStaffCode    string     `json:"merchant_staff_code,omitempty"`    // Ажилтны код
		Offset               QpayOffset `json:"offset"`                 // Хуудаслалт
	}

	// QpayPaymentRow [Төлбөрийн жагсаалтын мөр]
	// See: https://developer.qpay.mn/#payment-list
	QpayPaymentRow struct {
		PaymentID          string `json:"payment_id"`          // QPay-ээс үүссэн гүйлгээний дугаар
		PaymentDate        string `json:"payment_date"`        // Гүйлгээний огноо
		PaymentStatus      string `json:"payment_status"`      // NEW, FAILED, PAID, REFUNDED
		PaymentFee         string `json:"payment_fee"`         // Шимтгэлийн дүн
		PaymentAmount      string `json:"payment_amount"`      // Гүйлгээний үнийн дүн
		PaymentCurrency    string `json:"payment_currency"`    // Валют (MNT)
		PaymentWallet      string `json:"payment_wallet"`      // Воллетийн дугаар
		PaymentName        string `json:"payment_name"`        // Төлбөрийн нэр (Юнивишн г.м)
		PaymentDescription string `json:"payment_description"` // Гүйлгээний утга
		QrCode             string `json:"qr_code"`             // Ашиглагдсан QR код
		PaidBy             string `json:"paid_by"`             // Төрөл (P2P, CARD)
		ObjectType         string `json:"object_type"`         // MERCHANT, INVOICE, QR
		ObjectID           string `json:"object_id"`           // Харгалзах ID
	}

	// QpayPaymentListResponse [Төлбөрийн жагсаалт авах хариу]
	// See: https://developer.qpay.mn/#payment-list
	QpayPaymentListResponse struct {
		Count      int64             `json:"count"`       // Нийт мөрийн тоо
		PaidAmount int64             `json:"paid_amount"` // Нийт дүн
		Rows       []*QpayPaymentRow `json:"rows"`        // Төлбөрийн жагсаалт
	}

	// QpayGeneralResponse [Ерөнхий хариу]
	// Амжилттай болсон эсвэл алдааны мэдээллийг агуулсан ерөнхий бүтэц.
	QpayGeneralResponse struct {
		Error        string `json:"error,omitempty"`          // Алдааны код
		Message      string `json:"message,omitempty"`        // Алдааны мэдээлэл
		Status       string `json:"status,omitempty"`         // Төлөв (SUCCESS, FAILED г.м)
		InvoiceID    string `json:"invoice_id,omitempty"`     // Нэхэмжлэлийн ID
		PaymentID    string `json:"payment_id,omitempty"`     // Төлбөрийн ID
		QpayShortUrl string `json:"qPay_shortUrl,omitempty"`  // Shortcut URL
	}
)
