package sqlconnect

import (
	"database/sql"
	"fmt"
	"os"
)

func ConnectDB() (*sql.DB, error){
	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=require", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"))
	db, err := sql.Open("postgres", connectionString)

	if err != nil {
		return nil, err
	}
	fmt.Println("Connected to PostgreSQL.")
	return db, nil
}