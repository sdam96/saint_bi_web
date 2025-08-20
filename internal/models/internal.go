package models

type User struct {
	ID         int
	Username   string
	Password   string
	FirstLogin bool
}

type Connection struct {
	ID             int    `json:"ID"`
	Alias          string `json:"Alias"`
	ApiURL         string `json:"ApiURL"`
	ApiUser        string `json:"ApiUser"`
	ApiPassword    string `json:"ApiPassword"`
	RefreshSeconds int    `json:"RefreshSeconds"`
	ConfigID       int    `json:"ConfigID"`
	CurrencyISO    string `json:"CurrencyISO"`
	LocaleFormat   string `json:"LocaleFormat"`
}
