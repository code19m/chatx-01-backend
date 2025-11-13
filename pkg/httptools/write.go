package httptools

import (
	"encoding/json"
	"net/http"
)

func WriteResponse(code int, w http.ResponseWriter, resp any) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")

	if resp == nil {
		return
	}

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
