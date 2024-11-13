package onfido

import "fmt"

var ErrInvalidId = &OnfidoError{Type: "validation_error", Message: "id is required"}

// ------------------------------------------------------------------
//                          ONFIDO ERROR
// ------------------------------------------------------------------

type OnfidoError struct {
	Type    string         `json:"type,omitempty"`
	Message string         `json:"message,omitempty"`
	Fields  map[string]any `json:"fields,omitempty"`
}

func (e OnfidoError) Error() string {
	// build a string representation of the Error
	msg := "OnfidoError - "
	if e.Type != "" {
		msg += fmt.Sprintf("Type: %s\n", e.Type)
	}

	if e.Message != "" {
		msg += fmt.Sprintf("\tMessage: %s\n", e.Message)
	}

	if len(e.Fields) > 0 {
		msg += "\tFields:\t"
		for k, v := range e.Fields {
			msg += fmt.Sprintf("%s - %v\n\t\t", k, v)
		}
	}
	return msg
}
