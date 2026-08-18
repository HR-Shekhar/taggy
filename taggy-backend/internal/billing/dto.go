package billing

type createOrderResponse struct {
	KeyID       string `json:"key_id"`
	OrderID     string `json:"order_id"`
	Amount      int32  `json:"amount"`
	Currency    string `json:"currency"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PrefillName string `json:"prefill_name,omitempty"`
	PrefillEmail string `json:"prefill_email,omitempty"`
}

type verifyRequest struct {
	RazorpayOrderID   string `json:"razorpay_order_id" validate:"required"`
	RazorpayPaymentID string `json:"razorpay_payment_id" validate:"required"`
	RazorpaySignature string `json:"razorpay_signature" validate:"required"`
}

type verifyResponse struct {
	Subscription string `json:"subscription"`
	Message      string `json:"message"`
}

type statusResponse struct {
	Subscription        string `json:"subscription"`
	PremiumAmountPaise  int32  `json:"premium_amount_paise"`
	Currency            string `json:"currency"`
	BillingConfigured   bool   `json:"billing_configured"`
}
