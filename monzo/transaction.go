package monzo

// import "github.com/relvacode/monzo-transactions/monzo"

//go:generate easyjson $GOFILE

//easyjson:json
type Entity struct {
	ID      string `json:"id"`
	Created string `json:"created"`
}

//easyjson:json
type Address struct {
	Address   string  `json:"address"`
	City      string  `json:"city"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

//easyjson:json
type Merchant struct {
	Entity
	Name     string  `json:"name"`
	Category string  `json:"category"`
	Address  Address `json:"address"`
	GroupID  string  `json:"group_id"`
	Logo     string  `json:"logo"`
	Emoji    string  `json:"emoji"`
}

func init() {
	register("transaction.created", func() Event { return new(TransactionCreated) })
}

//easyjson:json
type TransactionCreated struct {
	Entity
	AccountID     string   `json:"account_id"`
	Currency      string   `json:"currency"`
	Amount        int      `json:"amount"`
	LocalCurrency string   `json:"local_currency"`
	LocalAmount   int      `json:"local_amount"`
	Description   string   `json:"description"`
	Category      string   `json:"category"`
	Settled       string  `json:"settled"`
	Merchant      Merchant `json:"merchant"`
	Notes         string   `json:"notes"`
}
