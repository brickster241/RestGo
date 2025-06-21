package middlewares

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"strings"
)

func CompressionMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts gzip encoding
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
		}

		// Set the response header
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()

		// Wrap the response writer
		w = &gzipResponseWriter{ResponseWriter: w, Writer: gz}

		next.ServeHTTP(w, r)
		fmt.Println("Sent Response from CompressionMW.")
	})
}

// Gzip Response writer wraps http.ResponseWriter to write gzipped responses.
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

func (g *gzipResponseWriter) Write (b []byte) (int, error) {
	return g.Writer.Write(b)
}