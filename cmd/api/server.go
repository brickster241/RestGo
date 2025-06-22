package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	mw "github.com/brickster241/rest-go/internal/api/middlewares"
	"github.com/brickster241/rest-go/internal/api/router"
	"github.com/brickster241/rest-go/pkg/utils"
)


func main() {

	cert := "cert.pem"
	key := "key.pem"

	
	/* rl := mw.NewRateLimiter(5, time.Minute)
	hppOptions := mw.HPPOptions{
		CheckQuery: true,
		CheckBody: true,
		CheckBodyOnlyForContentType: "application/x-www-form-urlencoded",
		WhiteList: []string{"sortBy", "sortOrder", "name", "age", "class"},
	}*/

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// Proper Middleware order.
	// secureMux := applyMiddleWares(mux, mw.Hpp(hppOptions), mw.CompressionMW, mw.SecurityHeadersMW, mw.ResponseTimeMW, rl.RateLimiterMW, mw.CorsMW)

	secureMux := utils.ApplyMiddleWares(router.Router(), mw.SecurityHeadersMW)
	// Define Port and Start server
	port := ":3000"

	// Create custom server
	server := &http.Server{
		Addr: port,
		Handler: secureMux,
		TLSConfig: tlsConfig,
	}

	fmt.Printf("Server running on Port %v\n", port)
	err := server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatalln("Couldn't start server... :", err)
	}
}