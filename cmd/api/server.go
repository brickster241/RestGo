package main

import (
	"crypto/tls"
	"log"
	"net/http"

	mw "github.com/brickster241/rest-go/internal/api/middlewares"
	"github.com/brickster241/rest-go/internal/api/router"
	"github.com/brickster241/rest-go/internal/repository/sqlconnect"
	"github.com/brickster241/rest-go/pkg/utils"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)


func main() {
	// Load the .env file
	godotenv.Load()

	_, err := sqlconnect.ConnectDB()

	if err != nil {
		log.Println("Error connecting to DB :", err)
		return
	}

	cert := "server.crt"
	key := "server.key"

	
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

	jwt_MW := mw.ExcludePathsMW(mw.JWT_MW, "/execs/login", "/execs/forgotpassword")
	secureMux := utils.ApplyMiddleWares(router.MainRouter(), jwt_MW, mw.SecurityHeadersMW)
	// Define Port and Start server
	port := ":3000"

	// Create custom server
	server := &http.Server{
		Addr: port,
		Handler: secureMux,
		TLSConfig: tlsConfig,
	}

	log.Printf("Server running on Port %v\n", port)
	err = server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatalln("Couldn't start server... :", err)
	}
}