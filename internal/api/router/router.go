package router

import (
	"net/http"
)

func MainRouter() *http.ServeMux {

	tRouter := teachersRouter()
	sRouter := studentsRouter()
	eRouter := execsRouter()
	
	// Chaining Routers
	sRouter.Handle("/", eRouter)
	tRouter.Handle("/", sRouter)
	return tRouter
}