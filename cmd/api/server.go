package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	mw "github.com/brickster241/rest-go/internal/api/middlewares"
)

func main() {

	// Handle default route
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to ROOT Page...")
	})

	// Handle teachers route
	http.HandleFunc("/teachers", func(w http.ResponseWriter, r *http.Request) {
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
	})

	// Handle students route
	http.HandleFunc("/students", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to Students Page...")
	})

	// Handle default route
	http.HandleFunc("/execs", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to Executives Page...")
	})

	rl := mw.NewRateLimiter(5, time.Minute)
	hppOptions := mw.HPPOptions{
		CheckQuery: true,
		CheckBody: true,
		CheckBodyOnlyForContentType: "application/x-www-form-urlencoded",
		WhiteList: []string{"sortBy", "sortOrder", "name", "age", "class"},
	}

	// Define Port and Start server
	port := 3000

	fmt.Println("Server running on Port :", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatalln("Couldn't start server... :", err)
	}
}