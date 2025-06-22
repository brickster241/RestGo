package handlers

import (
	"fmt"
	"net/http"
)

func StudentsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		fmt.Fprintf(w, "Welcome to Students Page... : GET Method")
	case http.MethodPost:
		fmt.Fprintf(w, "Welcome to Students Page... : POST Method")
	case http.MethodPut:
		fmt.Fprintf(w, "Welcome to Students Page... : PUT Method")
	case http.MethodPatch:
		fmt.Fprintf(w, "Welcome to Students Page... : PATCH Method")
	case http.MethodDelete:
		fmt.Fprintf(w, "Welcome to Students Page... : DELETE Method")
	default:
		fmt.Fprintf(w, "Invalid Request : %v", r.Method)
	}
}