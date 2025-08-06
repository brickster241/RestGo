package middlewares

import (
	"log"
	"net/http"
)

func XSS_MW(next http.Handler) http.Handler {
	log.Println("******* Initializing XSS_MW *******")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("+++++++ XSS_MW Ran +++++++")
		next.ServeHTTP(w, r)
		log.Println("------- Sending Response from XSS_MW -------")
	})
}