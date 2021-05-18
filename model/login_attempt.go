package model

type LoginAttempt struct {
	Email  string `json:"email"`
	Time   string `json:"time"`
	Status bool   `json:"status"`
}

type AttemptValidation struct {
	Success int
	Failed  int
}
