package structures

type ClientBalance struct {
	CurrencyCode  string
	ActualBalance float64
	FrozenBalance float64
}

// Invoice or Withdraw Request represents the data of the invoice creation request
type TransactionRequest struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
	Account  string  `json:"account"`
}
