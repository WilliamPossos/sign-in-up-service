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

type EmailVerification struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type OperationResult struct {
	Message string `json:"message"`
}
