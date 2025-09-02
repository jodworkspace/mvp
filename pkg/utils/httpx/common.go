package httpx

import "net/http"

func NoContent(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusNoContent)
	_, err := w.Write([]byte{})
	return err
}
