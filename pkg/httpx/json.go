package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type ErrorResponse struct {
	StatusCode int            `json:"status_code"`
	Message    string         `json:"message"`
	Details    map[string]any `json:"details,omitempty"`
}

type JSON map[string]any

func ReadJSON(r *http.Request, data any) error {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	// json.Unmarshal require reading the body into memory first using io.ReadAll
	err := decoder.Decode(data)
	if err != nil {
		return err
	}

	err = decoder.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must have a single JSON value")
	}

	return nil
}

func WriteJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) (int, error) {
	out, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return 0, err
	}

	if len(headers) > 0 {
		for k, v := range headers[0] {
			w.Header()[k] = v
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return w.Write(out)
}

func ErrorJSON(w http.ResponseWriter, errResp *ErrorResponse) (int, error) {
	if errResp.StatusCode == 0 {
		errResp.StatusCode = http.StatusInternalServerError
	}

	if errResp.Message == "" {
		errResp.Message = http.StatusText(errResp.StatusCode)
	}

	return WriteJSON(w, errResp.StatusCode, errResp)
}
