package model

// User represents the minimal user record stored in DB/Redis.
type User struct {
	UserID       string `json:"user_id"`
	Deeplink     string `json:"deeplink"`
	PromoMessage string `json:"promo_message"`
}
