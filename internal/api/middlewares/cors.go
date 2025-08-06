package middlewares

import (
	"log"
	"net/http"
)

var allowedOrigins = []string {
	"https://localhost:3000",
}

func CorsMW(next http.Handler) http.Handler {
	log.Println("******* Initializing CorsMW *******")
	
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("+++++++ CorsMW Ran +++++++")
		origin := r.Header.Get("Origin")

		// Only allow requests from specified urls' header.
		if !isOriginAllowed(origin) {
			http.Error(w, "Not Allowed by CORS.", http.StatusForbidden)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Expose-Headers", "Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle PreFlight check
		if r.Method == http.MethodOptions {
			return
		}

		next.ServeHTTP(w, r)
		log.Println("------- Sending Response from CorsMW -------")
	})
}

// Simple Linear Search to check whether origin is present in the allowed list.
func isOriginAllowed(origin string) bool {
	for _, allowedOrigin := range allowedOrigins {
		if origin == allowedOrigin {
			return true
		}
	}
	return false
}