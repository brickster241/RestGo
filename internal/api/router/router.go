package router

import (
	"net/http"

	"github.com/brickster241/rest-go/internal/api/handlers"
)

func Router() *http.ServeMux {
	mux := http.NewServeMux()

	// Handle default route
	mux.HandleFunc("GET /", handlers.RootHandler)

	// Handle teachers route
	mux.HandleFunc("GET /teachers", handlers.GetTeachersHandler)
	mux.HandleFunc("POST /teachers", handlers.PostTeachersHandler)
	mux.HandleFunc("PATCH /teachers", handlers.PatchTeachersHandler)
	mux.HandleFunc("DELETE /teachers", handlers.DeleteTeachersHandler)
	mux.HandleFunc("GET /teachers/{id}", handlers.GetOneTeacherHandler)
	mux.HandleFunc("PUT /teachers/{id}", handlers.PutOneTeacherHandler)
	mux.HandleFunc("PATCH /teachers/{id}", handlers.PatchOneTeacherHandler)
	mux.HandleFunc("DELETE /teachers/{id}", handlers.DeleteOneTeacherHandler)
	mux.HandleFunc("GET /teachers/{id}/students", handlers.GetStudentsByTeacherIDHandler)
	mux.HandleFunc("GET /teachers/{id}/studentcount", handlers.GetStudentCountByTeacherIDHandler)

	// Handle students route
	mux.HandleFunc("GET /students", handlers.GetStudentsHandler)
	mux.HandleFunc("POST /students", handlers.PostStudentsHandler)
	mux.HandleFunc("PATCH /students", handlers.PatchStudentsHandler)
	mux.HandleFunc("DELETE /students", handlers.DeleteStudentsHandler)
	mux.HandleFunc("GET /students/{id}", handlers.GetOneStudentHandler)
	mux.HandleFunc("PUT /students/{id}", handlers.PutOneStudentHandler)
	mux.HandleFunc("PATCH /students/{id}", handlers.PatchOneStudentHandler)
	mux.HandleFunc("DELETE /students/{id}", handlers.DeleteOneStudentHandler)

	// Handle default route
	mux.HandleFunc("GET /execs", handlers.ExecsHandler)

	return mux
}