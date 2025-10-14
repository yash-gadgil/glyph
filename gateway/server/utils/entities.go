package utils

type Watchlist struct {
	UserID  int64    `json:"u_id"`
	Id      int64    `json:"id"`
	Name    string   `json:"name"`
	Symbols []string `json:"symbols"`
}

type Order struct {
	UserID int64  `json:"u_id"`
	Id     int64  `json:"id"`
	Qty    int64  `json:"qty"`
	Symbol string `json:"symbol"`
	Side   string `json:"side"`
	Type   string `json:"type"`
}

type OrderList struct {
	UserID int64   `json:"u_id"`
	Status string  `json:"status"`
	Orders []Order `json:"orders"`
}

type AuthReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResp struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}
