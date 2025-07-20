package router

import (
	"net/http"

	"github.com/brickster241/rest-go/internal/api/handlers"
)

func execsRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// Handle Exec Routes
	mux.HandleFunc("GET /execs/", handlers.GetExecsHandler)
	mux.HandleFunc("POST /execs", handlers.PostExecsHandler)
	mux.HandleFunc("PATCH /execs", handlers.PatchExecsHandler)
	mux.HandleFunc("GET /execs/{id}", handlers.GetOneExecHandler)
	mux.HandleFunc("PATCH /execs/{id}", handlers.PatchOneExecHandler)
	mux.HandleFunc("DELETE /execs/{id}", handlers.DeleteOneExecHandler)

	mux.HandleFunc("POST /execs/{id}/updatepassword", handlers.UpdateExecPasswordHandler)
	mux.HandleFunc("POST /execs/login", handlers.LoginExecHandler)
	mux.HandleFunc("POST /execs/logout", handlers.LogoutExecHandler)
	mux.HandleFunc("POST /execs/forgotpassword", handlers.PostExecsHandler)
	mux.HandleFunc("POST /execs/resetpassword/reset/{resetcode}", handlers.PostExecsHandler)

	return mux
}