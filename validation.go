package errors

// MissingField is an error type that can be used when
// validating input
type MissingField string

func (e MissingField) Error() string {
	return string(e) + " is required"
}
