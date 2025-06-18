package main

import (
	"fmt"
	"net/http"
)

func main() {

	// Add functions which handle all http methods.
	
	http.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Handling incoming orders....")
	})

	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Handling users....")
	})

	// Define Port and Start server
	port := 3000

	fmt.Println("Server running on Port :", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}