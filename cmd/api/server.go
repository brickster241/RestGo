package main

import (
	"crypto/tls"
	"embed"
	"log"
	"net/http"
	"os"
	"time"

	mw "github.com/brickster241/rest-go/internal/api/middlewares"
	"github.com/brickster241/rest-go/internal/api/router"
	"github.com/brickster241/rest-go/pkg/utils"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

//go:embed  .env
var envFile embed.FS

func loadEnvFromEmbeddedFile() {
	// Read embedded .env file
	content, err := envFile.ReadFile(".env")
	if err != nil {
		log.Fatalf("Error reading .env File : %v", err)
		return
	}

	// Create a temp file to load the env vars
	tempFile, err := os.CreateTemp("", ".env")
	if err != nil {
		log.Fatalf("Error creating temp .env File : %v", err)
		return
	}
	defer os.Remove(tempFile.Name())

	// Write content of embedded .env file to the temp file.
	_, err = tempFile.Write(content)
	if err != nil {
		log.Fatalf("Error writing to temp .env file: %v", err)
		return
	}

	err = tempFile.Close()
	if err != nil {
		log.Fatalf("Error closing temp File : %v", err)
		return
	}

	// Load env vars from the temp file
	err = godotenv.Load(tempFile.Name())
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
		return
	}
}

func main() {
	// Only in development for running source code.
	loadEnvFromEmbeddedFile()

	cert := os.Getenv("CERT_FILE")
	key := os.Getenv("KEY_FILE")

	
	rl := mw.NewRateLimiter(5, time.Minute)
	hppOptions := mw.HPPOptions{
		CheckQuery: true,
		CheckBody: true,
		CheckBodyOnlyForContentType: "application/x-www-form-urlencoded",
		WhiteList: []string{"sortBy", "sortOrder", "name", "age", "class"},
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// Proper Middleware order.
	jwt_MW := mw.ExcludePathsMW(mw.JWT_MW, "/execs/login", "/execs/forgotpassword", "/execs/resetpassword/reset")
	secureMux := utils.ApplyMiddleWares(router.MainRouter(), mw.Hpp(hppOptions), mw.SecurityHeadersMW, mw.CompressionMW, jwt_MW, mw.XSS_MW, mw.ResponseTimeMW, rl.RateLimiterMW, mw.CorsMW)
	// Define Port and Start server
	port := ":3000"

	// Create custom server
	server := &http.Server{
		Addr: port,
		Handler: secureMux,
		TLSConfig: tlsConfig,
	}

	log.Printf("Server running on Port %v\n", port)
	err := server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatalln("Couldn't start server... :", err)
	}
}