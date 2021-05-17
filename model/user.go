package model

type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type EmailVerification struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}
