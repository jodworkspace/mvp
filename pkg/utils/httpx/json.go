package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type JSON map[string]any

type ErrorResponse struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Details    JSON   `json:"details,omitempty"`
}

func ReadJSON(r *http.Request, data any) error {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

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

func WriteJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for k, v := range headers[0] {
			w.Header()[k] = v
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_, err = w.Write(out)
	return err
}

func ErrorJSON(w http.ResponseWriter, errResp ErrorResponse) error {
	if errResp.StatusCode == 0 {
		errResp.StatusCode = http.StatusInternalServerError
	}

	if errResp.Message == "" {
		errResp.Message = http.StatusText(errResp.StatusCode)
	}

	return WriteJSON(w, errResp.StatusCode, errResp)
}
