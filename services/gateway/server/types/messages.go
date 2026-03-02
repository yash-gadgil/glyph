package types

type Watchlist struct {
	UserID  string   `json:"userId"`
	Id      int64    `json:"id"`
	Name    string   `json:"name"`
	Symbols []string `json:"symbols"`
}

type Order struct {
	ID          string `json:"id"`
	UserID      string `json:"userId"`
	Symbol      string `json:"symbol"`
	Side        string `json:"side"`
	OrderType   string `json:"orderType"`
	TimeInForce string `json:"timeInForce"`
	Qty         int64  `json:"qty"`
	FilledQty   int64  `json:"filledQty"`
	Price       int64  `json:"price,omitempty"`
	StopPrice   int64  `json:"stopPrice,omitempty"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type CreateOrderRequest struct {
	Symbol      string `json:"symbol"`
	Side        string `json:"side"`
	OrderType   string `json:"orderType"`
	TimeInForce string `json:"timeInForce"`
	Quantity    int64  `json:"quantity"`
	Price       int64  `json:"price,omitempty"`
	StopPrice   int64  `json:"stopPrice,omitempty"`
}

type SignupReq struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SigninReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RecoveryReq struct {
	Email string `json:"email"`
}

type ResetPasswordReq struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type Bar struct {
	Symbol string  `json:"symbol"`
	Open   float32 `json:"open"`
	High   float32 `json:"high"`
	Low    float32 `json:"low"`
	Close  float32 `json:"close"`
}

type SymbolsArr struct {
	Symbols []string `json:"symbols"`
}
