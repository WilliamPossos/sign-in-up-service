package model

type SignUpDto struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignInDto struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	VerificationCode string `json:"code"`
}

type UserDynamoDb struct {
	Username         string `json:"username"`
	Email            string `json:"email"`
	Password         string `json:"password"`
	VerificationCode string `json:"code"`
}

type EmailConfig struct {
	Email    string `json:"email"`
	HtmlBody string `json:"body"`
	Subject  string `json:"subject"`
}

type OperationResult struct {
	OkMessage string `json:"okMessage"`
}
