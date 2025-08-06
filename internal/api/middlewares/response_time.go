package middlewares

import (
	"log"
	"net/http"
	"time"
)

func ResponseTimeMW(next http.Handler) http.Handler {
	log.Println("******* Initializing ResponseTimeMW *******")
	
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("+++++++ ResponseTimeMW Ran +++++++")
	
		start := time.Now()
		rw := &responseTimeWriter{ResponseWriter: w, status: http.StatusOK}
		
		// Calculate the duration
		duration := time.Since(start)
		w.Header().Set("X-Response-Time", duration.String())
		next.ServeHTTP(rw, r)
		
		
		duration = time.Since(start)
		// Log the request details
		log.Printf("Method: %s, URL: %s, Status: %d, Duration: %v\n", r.Method, r.URL, rw.status, duration.String())
		log.Println("------- Sending Response from ResponseTimeMW -------")
	})
}

// Response Writer
type responseTimeWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseTimeWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}