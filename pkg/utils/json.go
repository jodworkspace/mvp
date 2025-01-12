package utils

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

func ReadJSON(w http.ResponseWriter, r *http.Request, data any) error {
	// r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	decoder := json.NewDecoder(r.Body)
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

func WriteJSON(w http.ResponseWriter, r *http.Request, status int, data any, headers ...http.Header) error {
	out, err := json.MarshalIndent(data, "", "\t")
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

func ErrorJSON(w http.ResponseWriter, r *http.Request, data *ErrorResponse) error {
	if data.StatusCode == 0 {
		data.StatusCode = http.StatusInternalServerError
	}
	err := WriteJSON(w, r, data.StatusCode, data)
	return err
}
