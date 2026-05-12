package qpay_quick

type (
	qpayLoginResponse struct {
		TokenType        string `json:"token_type"`         // Токены төрөл (Bearer)
		RefreshToken     string `json:"refresh_token"`      // Шинэчлэх токен
		RefreshExpiresIn int64  `json:"refresh_expires_in"` // Шинэчлэх токены хүчинтэй хугацаа (Unix timestamp)
		AccessToken      string `json:"access_token"`       // Хандалтын токен
		ExpiresIn        int64  `json:"expires_in"`         // Хандалтын токены хүчинтэй хугацаа (Unix timestamp)
		Scope            string `json:"scope"`              // Хандах хүрээ
		NotBeforePolicy  string `json:"not-before-policy"`  // Бодлого
		SessionState     string `json:"session_state"`      // Сессийн төлөв
	}

	// QpayCompanyCreateRequest [Байгууллагаар мерчант үүсгэх/шинэчлэх хүсэлт]
	QpayCompanyCreateRequest struct {
		OwnerRegNo     string `json:"owner_register_no,omitempty"`  // Эзэмшигчийн регистр
		OwnerFirstName string `json:"owner_first_name"`             // Эзэмшигчийн овог
		OwnerLastName  string `json:"owner_last_name"`              // Эзэмшигчийн нэр
		LocationLat    string `json:"location_lat,omitempty"`       // Өргөрөг
		LocationLng    string `json:"location_lng,omitempty"`       // Уртраг
		RegisterNo     string `json:"register_number"`              // Байгууллагын регистр
		CompanyName    string `json:"company_name"`                 // Байгууллагын нэр
		Name           string `json:"name"`                         // Бизнесийн нэр
		NameEng        string `json:"name_eng,omitempty"`           // Бизнесийн англи нэр
		MCCcode        string `json:"mcc_code"`                     // МCC код
		City           string `json:"city"`                         // Хот, аймгийн код
		District       string `json:"district"`                     // Сум, дүүргийн код
		Address        string `json:"address"`                      // Хаяг
		Phone          string `json:"phone"`                        // Утас
		Email          string `json:"email"`                        // И-мэйл
	}

	// QpayCompanyCreateResponse [Байгууллагаар мерчант үүсгэсэн хариу]
	QpayCompanyCreateResponse struct {
		ID                    string `json:"id"`                      // Мерчантын ID
		VendorID              string `json:"vendor_id"`               // Вендорын ID
		Type                  string `json:"type"`                    // COMPANY
		RegisterNo            string `json:"register_number"`         // Регистр
		Name                  string `json:"name"`                    // Бизнесийн нэр
		NameEng               string `json:"name_eng"`                // Англи нэр
		OwnerRegNo            string `json:"owner_register_no"`       // Эзэмшигчийн регистр
		OwnerFirstName        string `json:"owner_first_name"`        // Эзэмшигчийн овог
		OwnerLastName         string `json:"owner_last_name"`         // Эзэмшигчийн нэр
		CompanyName           string `json:"company_name"`            // Байгууллагын нэр
		GBusinessDirectionID  string `json:"g_business_direction_id"` // Бизнесийн чиглэлийн ID
		MCCcode               string `json:"mcc_code"`                // МCC код
		City                  string `json:"city"`                    // Хот
		District              string `json:"district"`                // Дүүрэг
		Address               string `json:"address"`                 // Хаяг
		Phone                 string `json:"phone"`                   // Утас
		Email                 string `json:"email"`                   // И-мэйл
		LocationLat           string `json:"location_lat"`            // Өргөрөг
		LocationLng           string `json:"location_lng"`            // Уртраг
	}

	// QpayPersonCreateRequest [Хувь хүнээр мерчант үүсгэх/шинэчлэх хүсэлт]
	QpayPersonCreateRequest struct {
		RegisterNo      string `json:"register_number"`           // Хувь хүний регистр
		FirstName       string `json:"first_name"`                // Овог
		LastName        string `json:"last_name"`                 // Нэр
		BusinessName    string `json:"business_name"`             // Бизнесийн нэр
		BusinessNameEng string `json:"business_name_eng,omitempty"` // Англи нэр
		MCCcode         string `json:"mcc_code"`                  // МCC код
		City            string `json:"city"`                      // Хот
		District        string `json:"district"`                  // Дүүрэг
		Address         string `json:"address"`                   // Хаяг
		Phone           string `json:"phone"`                     // Утас
		Email           string `json:"email"`                     // И-мэйл
	}

	// QpayPersonCreateResponse [Хувь хүнээр мерчант үүсгэсэн хариу]
	QpayPersonCreateResponse struct {
		ID                   string `json:"id"`
		VendorID             string `json:"vendor_id"`
		Type                 string `json:"type"`                    // PERSON
		RegisterNo           string `json:"register_number"`
		FirstName            string `json:"first_name"`
		LastName             string `json:"last_name"`
		BusinessName         string `json:"business_name"`
		BusinessNameEng      string `json:"business_name_eng"`
		GBusinessDirectionID string `json:"g_business_direction_id"`
		MCCcode              string `json:"mcc_code"`
		City                 string `json:"city"`
		District             string `json:"district"`
		Address              string `json:"address"`
		Phone                string `json:"phone"`
		Email                string `json:"email"`
	}

	// QpayMerchantListRequest [Мерчантын жагсаалт хүсэлт]
	// Doc body: {"page": 1, "limit": 10}
	QpayMerchantListRequest struct {
		Page  int64 `json:"page"`  // Хуудасны дугаар (min 1)
		Limit int64 `json:"limit"` // Нэг хуудас дахь мөрийн тоо (min 1, max 1000)
	}

	// QpayMerchantListResponse [Мерчантын жагсаалт хариу]
	QpayMerchantListResponse struct {
		Count int                       `json:"count"`
		Items []QpayMerchantGetResponse `json:"rows"`
	}

	// QpayMerchantGetResponse [Мерчантын мэдээлэл]
	QpayMerchantGetResponse struct {
		CreateDate           string `json:"created_date"`
		ID                   string `json:"id"`
		Type                 string `json:"type"` // COMPANY | PERSON
		RegisterNo           string `json:"register_number"`
		Name                 string `json:"name"`
		NameEng              string `json:"name_eng"`
		FirstName            string `json:"first_name"`
		LastName             string `json:"last_name"`
		BusinessName         string `json:"business_name"`
		BusinessNameEng      string `json:"business_name_eng"`
		CompanyName          string `json:"company_name"`
		OwnerRegNo           string `json:"owner_register_no"`
		OwnerFirstName       string `json:"owner_first_name"`
		OwnerLastName        string `json:"owner_last_name"`
		GBusinessDirectionID string `json:"g_business_direction_id"`
		MCCcode              string `json:"mcc_code"`
		City                 string `json:"city"`
		District             string `json:"district"`
		Address              string `json:"address"`
		Phone                string `json:"phone"`
		Email                string `json:"email"`
		LocationLat          string `json:"location_lat"`
		LocationLng          string `json:"location_lng"`
		WechatRegistered     bool   `json:"wechat_registered"`
		WechatTerminalID     string `json:"wechat_terminal_id"`
	}

	// QpayLocationCode [Хот/аймаг эсвэл сум/дүүргийн код]
	QpayLocationCode struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}

	// QpayInvoiceRequest [Нэхэмжлэх үүсгэх хүсэлт]
	QpayInvoiceRequest struct {
		MerchantID   string                   `json:"merchant_id"`
		BranchCode   string                   `json:"branch_code,omitempty"`
		Amount       float64                  `json:"amount"`
		Currency     string                   `json:"currency"`
		CustomerName string                   `json:"customer_name"`
		CustomerLogo string                   `json:"customer_logo"`
		CallbackUrl  string                   `json:"callback_url"`
		Description  string                   `json:"description"`
		MCCcode      string                   `json:"mcc_code,omitempty"`
		BankAccounts []QpayBankAccountRequest `json:"bank_accounts"`
	}

	// QpayInvoiceResponse [Нэхэмжлэх үүсгэсэн хариу]
	QpayInvoiceResponse struct {
		ID                  string                    `json:"id"`
		TerminalID          string                    `json:"terminal_id"`
		Amount              string                    `json:"amount"`
		QrCode              string                    `json:"qr_code"`
		Description         string                    `json:"description"`
		InvoiceStatus       string                    `json:"invoice_status"`
		InvoiceStatusDate   string                    `json:"invoice_status_date"`
		CallbackUrl         string                    `json:"callback_url"`
		CustomerName        string                    `json:"customer_name"`
		CustomerLogo        string                    `json:"customer_logo"`
		Currency            string                    `json:"currency"`
		MCCcode             string                    `json:"mcc_code"`
		LegacyID            string                    `json:"legacy_id"`
		VendorID            string                    `json:"vendor_id"`
		ProcessCodeID       string                    `json:"process_code_id"`
		QrImage             string                    `json:"qr_image"`
		InvoiceBankAccounts []QpayBankAccountResponse `json:"invoice_bank_accounts"`
		Urls                []QpayUrls                `json:"urls"`
	}

	// QpayUrls [Банкны апп-ын Deeplink холбоос]
	QpayUrls struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Logo        string `json:"logo"`
		Link        string `json:"link"`
	}

	// QpayBankAccountRequest [Төлбөр хүлээн авах данс]
	QpayBankAccountRequest struct {
		AccountBankCode string `json:"account_bank_code"`
		AccountNumber   string `json:"account_number"`
		AccountName     string `json:"account_name"`
		IsDefault       bool   `json:"is_default"`
	}

	// QpayInvoiceGetResponse [Нэхэмжлэлийн мэдээлэл]
	QpayInvoiceGetResponse struct {
		ID                  string                    `json:"id"`
		TerminalID          string                    `json:"terminal_id"`
		Amount              string                    `json:"amount"`
		QrCode              string                    `json:"qr_code"`
		Description         string                    `json:"description"`
		InvoiceStatus       string                    `json:"invoice_status"`
		InvoiceStatusDate   string                    `json:"invoice_status_date"`
		CallbackUrl         string                    `json:"callback_url"`
		CustomerName        string                    `json:"customer_name"`
		CustomerLogo        string                    `json:"customer_logo"`
		Currency            string                    `json:"currency"`
		MCCcode             string                    `json:"mcc_code"`
		LegacyID            string                    `json:"legacy_id"`
		VendorID            string                    `json:"vendor_id"`
		ProcessCodeID       string                    `json:"process_code_id"`
		QrImage             string                    `json:"qr_image"`
		InvoiceBankAccounts []QpayBankAccountResponse `json:"invoice_bank_accounts"`
		Urls                []QpayUrls                `json:"urls"`
	}

	// QpayBankAccountResponse [Дансны мэдээлэл (хариу)]
	QpayBankAccountResponse struct {
		ID              string `json:"id"`
		AccountBankCode string `json:"account_bank_code"`
		AccountNumber   string `json:"account_number"`
		AccountName     string `json:"account_name"`
		IsDefault       bool   `json:"is_default"`
		InvoiceID       string `json:"invoice_id"`
	}

	// QpayPaymentCheckRequest [Төлбөр шалгах хүсэлт]
	QpayPaymentCheckRequest struct {
		InvoiceID string `json:"invoice_id"`
	}

	// QpayPaymentCheckResponse [Төлбөр шалгах хариу]
	QpayPaymentCheckResponse struct {
		ID                string        `json:"id"`
		InvoiceStatus     string        `json:"invoice_status"`
		InvoiceStatusDate string        `json:"invoice_status_date"`
		Payments          []QpayPayment `json:"payments"`
	}

	// QpayPayment [Төлбөрийн мэдээлэл]
	QpayPayment struct {
		ID                 string             `json:"id"`
		TerminalID         string             `json:"terminal_id"`
		WalletCustomerID   string             `json:"wallet_customer_id"`
		Amount             string             `json:"amount"`
		Currency           string             `json:"currency"`
		PaymentName        string             `json:"payment_name"`
		PaymentDescription string             `json:"payment_description"`
		PaidBy             string             `json:"paid_by"`
		Note               string             `json:"note"`
		PaymentStatus      string             `json:"payment_status"`
		PaymentStatusDate  string             `json:"payment_status_date"`
		Transactions       []QpayTransactions `json:"transactions"`
	}

	// QpayTransactions [Гүйлгээний мэдээлэл]
	QpayTransactions struct {
		ID                  string `json:"id"`
		Description         string `json:"description"`
		TransactionBankCode string `json:"transaction_bank_code"`
		AccountBankCode     string `json:"account_bank_code"`
		AccountBankName     string `json:"account_bank_name"`
		AccountNumber       string `json:"account_number"`
		Status              string `json:"status"`
		Amount              string `json:"amount"`
		Currency            string `json:"currency"`
	}

	// QpayGeneralResponse [Ерөнхий хариу — амжилттай эсвэл алдааны мэдээлэл]
	QpayGeneralResponse struct {
		Error   string `json:"error,omitempty"`
		Message string `json:"message,omitempty"`
		Status  string `json:"status,omitempty"`
	}
)
