package response

import (
	"fmt"
	"github.com/go-playground/validator/v10"
)

type Response struct {
	Errors []string `json:"error,omitempty"`
}

func Error(errorMessages ...string) Response {
	return Response{Errors: errorMessages}
}

func ValidateError(errs validator.ValidationErrors) Response {
	var validationErrorsMessages []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			validationErrorsMessages = append(validationErrorsMessages, fmt.Sprintf("%s is required", err.Field()))
		case "url":
			validationErrorsMessages = append(validationErrorsMessages, fmt.Sprintf("%s is not a URL", err.Field()))
		default:
			validationErrorsMessages = append(validationErrorsMessages, fmt.Sprintf("%s is not a valid value", err.Field()))
		}
	}

	return Response{Errors: validationErrorsMessages}
}
