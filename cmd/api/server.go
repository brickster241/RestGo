package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"

	mw "github.com/brickster241/rest-go/internal/api/middlewares"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		fmt.Fprintf(w, "Welcome to Root Page... : GET Method")
	case http.MethodPost:
		fmt.Fprintf(w, "Welcome to Root Page... : POST Method")
	case http.MethodPut:
		fmt.Fprintf(w, "Welcome to Root Page... : PUT Method")
	case http.MethodPatch:
		fmt.Fprintf(w, "Welcome to Root Page... : PATCH Method")
	case http.MethodDelete:
		fmt.Fprintf(w, "Welcome to Root Page... : DELETE Method")
	default:
		fmt.Fprintf(w, "Invalid Request : %v", r.Method)
	}
}

func teachersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		fmt.Fprintf(w, "Welcome to Teachers Page... : GET Method")
	case http.MethodPost:
		fmt.Fprintf(w, "Welcome to Teachers Page... : POST Method")
	case http.MethodPut:
		fmt.Fprintf(w, "Welcome to Teachers Page... : PUT Method")
	case http.MethodPatch:
		fmt.Fprintf(w, "Welcome to Teachers Page... : PATCH Method")
	case http.MethodDelete:
		fmt.Fprintf(w, "Welcome to Teachers Page... : DELETE Method")
	default:
		fmt.Fprintf(w, "Invalid Request : %v", r.Method)
	}
}

func studentsHandler(w http.ResponseWriter, r *http.Request) {
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

func execsHandler(w http.ResponseWriter, r *http.Request) {
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

func main() {

	cert := "cert.pem"
	key := "key.pem"

	mux := http.NewServeMux()

	// Handle default route
	mux.HandleFunc("/", rootHandler)

	// Handle teachers route
	mux.HandleFunc("/teachers", teachersHandler)

	// Handle students route
	mux.HandleFunc("/students", studentsHandler)

	// Handle default route
	mux.HandleFunc("/execs", execsHandler)

	rl := mw.NewRateLimiter(5, time.Minute)
	hppOptions := mw.HPPOptions{
		CheckQuery: true,
		CheckBody: true,
		CheckBodyOnlyForContentType: "application/x-www-form-urlencoded",
		WhiteList: []string{"sortBy", "sortOrder", "name", "age", "class"},
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	secureMux := mw.Hpp(hppOptions)(rl.RateLimiterMW(mw.CompressionMW(mw.ResponseTimeMW(mw.SecurityHeadersMW(mw.CorsMW(mux))))))

	// Define Port and Start server
	port := ":3000"

	// Create custom server
	server := &http.Server{
		Addr: port,
		Handler: secureMux,
		TLSConfig: tlsConfig,
	}

	fmt.Println("Server running on Port :", port)
	err := server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatalln("Couldn't start server... :", err)
	}
}