package jobs

type EmailPayload struct {
	To      string `json:"to" validate:"required,email"`
	Subject string `json:"subject" validate:"required,min=3,max=200"`
	Body    string `json:"body" validate:"required,min=5"`
}