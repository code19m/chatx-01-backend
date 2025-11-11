package httpx

import "net/http"

func HandleError(w http.ResponseWriter, err error) {
	// TODO: handle based on error type
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
