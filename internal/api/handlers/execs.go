package handlers

import (
	"fmt"
	"net/http"
)

func ExecsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		fmt.Fprintf(w, "Welcome to Executives Page... : GET Method")
	case http.MethodPost:
		fmt.Fprintf(w, "Welcome to Executives Page... : POST Method")
	case http.MethodPut:
		fmt.Fprintf(w, "Welcome to Executives Page... : PUT Method")
	case http.MethodPatch:
		fmt.Fprintf(w, "Welcome to Executives Page... : PATCH Method")
	case http.MethodDelete:
		fmt.Fprintf(w, "Welcome to Executives Page... : DELETE Method")
	default:
		fmt.Fprintf(w, "Invalid Request : %v", r.Method)
	}
}
