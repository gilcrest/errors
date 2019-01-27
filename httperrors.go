package errors

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

// hError represents an HTTP handler error. It provides methods for a HTTP status
// code and embeds the built-in error interface.
type hError interface {
	error
	Status() int
	ErrKind() string
	ErrCode() string
}

// HTTPErr represents an error with an associated HTTP status code.
type HTTPErr struct {
	HTTPStatusCode int
	Kind           Kind
	Code           Code
	Err            error
}

// Allows HTTPErr to satisfy the error interface.
func (hse HTTPErr) Error() string {
	return hse.Err.Error()
}

// SetErr creates an error type and adds it to the struct
func (hse *HTTPErr) SetErr(s string) {
	hse.Err = Str(s)
}

// ErrKind returns a string denoting the "kind" of error
func (hse HTTPErr) ErrKind() string {
	return hse.Kind.String()
}

// ErrCode returns a string denoting the "kind" of error
func (hse HTTPErr) ErrCode() string {
	return string(hse.Code)
}

// Status Returns an HTTP Status Code.
func (hse HTTPErr) Status() int {
	return hse.HTTPStatusCode
}

type errResponse struct {
	Error svcError `json:"error"`
}

type svcError struct {
	Kind    string `json:"kind"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// HTTPError takes a writer and an error, performs a type switch to
// determine if the type is an HTTPError (which meets the Error interface
// as defined in this package), then sends the Error as a response to the
// client. If the type does not meet the Error interface as defined in this
// package, then a proper error is still formed and sent to the client,
// however, the Kind and Code will be Unanticipated.
func HTTPError(w http.ResponseWriter, err error) {
	const op Op = "errors.httpError"

	if err != nil {
		// We perform a "type switch" https://tour.golang.org/methods/16
		// to determine the interface value type
		switch e := err.(type) {
		// If the interface value is of type Error (not a typical error, but
		// the Error interface defined above), then
		case hError:
			// We can retrieve the status here and write out a specific
			// HTTP status code.
			log.Printf("HTTP %d - %s", e.Status(), e)

			er := errResponse{
				Error: svcError{
					Kind:    e.ErrKind(),
					Code:    e.ErrCode(),
					Message: e.Error(),
				},
			}

			// Marshal errResponse struct to JSON for the response body
			errJSON, _ := json.MarshalIndent(er, "", "    ")

			sendError(w, string(errJSON), e.Status())

		default:
			// Any error types we don't specifically look out for default
			// to serving a HTTP 500
			cd := http.StatusInternalServerError
			er := errResponse{
				Error: svcError{
					Kind:    Unanticipated.String(),
					Code:    "Unanticipated",
					Message: "Unexpected error - contact support",
				},
			}

			log.Error().Msgf("Unknown Error - HTTP %d - %s", cd, err.Error())

			// Marshal errResponse struct to JSON for the response body
			errJSON, _ := json.MarshalIndent(er, "", "    ")

			sendError(w, string(errJSON), cd)
		}
	}
}

// Taken from standard library, but changed to send application/json as header
// Error replies to the request with the specified error message and HTTP code.
// It does not otherwise end the request; the caller should ensure no further
// writes are done to w.
// The error message should be json.
func sendError(w http.ResponseWriter, error string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(statusCode)
	fmt.Fprintln(w, error)
}
