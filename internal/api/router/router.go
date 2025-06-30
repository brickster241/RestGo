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
	mux.HandleFunc("GET /teachers/", handlers.GetTeachersHandler)
	mux.HandleFunc("POST /teachers/", handlers.PostTeacherHandler)
	mux.HandleFunc("PATCH /teachers/", handlers.PatchTeachersHandler)
	mux.HandleFunc("DELETE /teachers/", handlers.DeleteTeachersHandler)
	mux.HandleFunc("GET /teachers/{id}", handlers.GetOneTeacherHandler)
	mux.HandleFunc("PUT /teachers/{id}", handlers.PutOneTeacherHandler)
	mux.HandleFunc("PATCH /teachers/{id}", handlers.PatchOneTeacherHandler)
	mux.HandleFunc("DELETE /teachers/{id}", handlers.DeleteOneTeacherHandler)

	// Handle students route
	mux.HandleFunc("/students", handlers.StudentsHandler)

	// Handle default route
	mux.HandleFunc("/execs", handlers.ExecsHandler)

	return mux
}