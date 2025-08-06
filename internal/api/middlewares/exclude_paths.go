package middlewares

import (
	"log"
	"net/http"
	"strings"
)

func ExcludePathsMW(middleware func(http.Handler) http.Handler, excludedPaths ... string) func (http.Handler) http.Handler {
	log.Println("******* Initializing ExcludePathsMW *******")
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, path := range excludedPaths {
				if strings.HasPrefix(r.URL.Path, path) {
					next.ServeHTTP(w, r)
					return
				}
			}
			middleware(next).ServeHTTP(w, r)
		})

	}
}