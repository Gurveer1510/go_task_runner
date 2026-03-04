package jobs

import (
	"encoding/json"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/go-task-runner/internal/models"
)

func ValidateJob(job *models.Job, v *validator.Validate) error {
	switch job.Type {
	case "email":
		var emailPayload EmailPayload

		if err := json.Unmarshal(job.Payload, &emailPayload); err != nil {
			return errors.New("invalid email payload format")
		}

		if err := v.Struct(emailPayload); err != nil {
			validationErrors := err.(validator.ValidationErrors)

			for _, e := range validationErrors {
				switch e.Field() {
				case "To":
					return errors.New("invalid email address")
				case "Subject":
					return errors.New("subject must be between 3 and 200 characters")
				case "Body":
					return errors.New("body must be at least 5 characters")
				}
			}
		}
	default:
		return errors.New("unsupported job type")
	}
	return nil
}
