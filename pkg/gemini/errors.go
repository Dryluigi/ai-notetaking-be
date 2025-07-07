package gemini

import "fmt"

type EmbedErrorType int

const (
	ErrTypeUnknown EmbedErrorType = iota
	ErrTypeInvalidAPIKey
	ErrTypeHTTPStatus
	ErrTypeRequestFailed
	ErrTypeJSONUnmarshal
	ErrTypeMarshalRequest
)

type EmbedError struct {
	Type    EmbedErrorType
	Message string
	Err     error
}

func (e *EmbedError) Error() string {
	return fmt.Sprintf("[%v] %s: %v", e.Type, e.Message, e.Err)
}

func (e *EmbedError) Unwrap() error {
	return e.Err
}
