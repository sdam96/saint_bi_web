package models

type User struct {
	ID         int
	Username   string
	Password   string
	FirstLogin bool
}

type Connection struct {
	ID             int
	Alias          string
	ApiURL         string
	ApiUser        string
	ApiPassword    string
	RefreshSeconds int
	ConfigID       int
}
