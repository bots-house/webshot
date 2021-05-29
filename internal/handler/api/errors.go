package api

import "encoding/json"

type HTTPError struct {
	Err  error
	Code int
}

func (err *HTTPError) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Err string `json:"details"`
	}{
		Err: err.Error(),
	})
}

func httpError(err error, code int) *HTTPError {
	return &HTTPError{Err: err, Code: code}
}

func (err *HTTPError) Error() string {
	return err.Err.Error()
}
