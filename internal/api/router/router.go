package router

import (
	"net/http"

	"github.com/brickster241/rest-go/internal/api/handlers"
)

func Router() *http.ServeMux {
	mux := http.NewServeMux()

	// Handle default route
	mux.HandleFunc("/", handlers.RootHandler)

	// Handle teachers route
	mux.HandleFunc("/teachers/", handlers.TeachersHandler)

	// Handle students route
	mux.HandleFunc("/students", handlers.StudentsHandler)

	// Handle default route
	mux.HandleFunc("/execs", handlers.ExecsHandler)

	return mux
}