package handlers

import (
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	models "github.com/brickster241/rest-go/internal/models"
	"github.com/brickster241/rest-go/internal/repository/sqlconnect"
	"github.com/brickster241/rest-go/pkg/utils"
	"golang.org/x/crypto/argon2"
)

var mu_exec = &sync.Mutex{}

// GET execs/
func GetExecsHandler(w http.ResponseWriter, r *http.Request) {
	
	query := "SELECT id, first_name, last_name, email, username, user_created_at, inactive_status, role FROM execs WHERE 1=1"
	var args []interface{}
	
	// Filter based on different params
	query, args = addQueryFiltersExec(r, query, args)

	// Will be of type param:asc or param:desc
	query = applySortingFiltersExec(r, query)

	// Connect to DB
	execList, err := sqlconnect.GetExecsDBHandler(query, args)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := struct {
		Status string    `json:"status"`
		Count  int       `json:"count"`
		Data   []models.Exec `json:"data"`
	}{
		Status: "success",
		Count:  len(execList),
		Data:   execList,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GET /execs/{id}
func GetOneExecHandler(w http.ResponseWriter, r *http.Request) {
	
	idStr := r.PathValue("id")
	// Handle Path Parameters
	execId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid Exec ID.").Error(), http.StatusBadRequest)
		return
	}

	// Connect to DB
	exec, err := sqlconnect.GetOneExecDBHandler(execId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exec)
}

func applySortingFiltersExec(r *http.Request, query string) string {
	sortParams := r.URL.Query()["sortby"]
	addQuery := " ORDER BY"
	if len(sortParams) > 0 {
		for i, sortParam := range sortParams {
			parts := strings.Split(sortParam, ":")
			if len(parts) != 2 {
				continue
			}

			field, order := parts[0], parts[1]
			if !isValidOrder(order) || !isValidSortFieldExec(field) {
				continue
			}
			// To ensure to incorporate multiple sorting values
			if i > 0 {
				addQuery += ","
			}
			addQuery += fmt.Sprintf(" %s %s", field, order)
		}
		if addQuery != " ORDER BY" {
			query += addQuery
		}
	}
	return query
}

func isValidSortFieldExec(field string) bool {
	validFields := map[string]bool{
		"first_name": true,
		"last_name": true,
		"email": false,
	}
	return validFields[field]
}

func addQueryFiltersExec(r *http.Request, query string, args []interface{}) (string, []interface{}) {
	params := []string{
		"first_name",
		"last_name",
		"email",
		"username",
		"inactive_status",
		"role",
	}

	for _, param := range params {
		value := r.URL.Query().Get(param)
		if value != "" {
			query += fmt.Sprintf(" AND %s=$%d", param, len(args)+1)
			args = append(args, value)
		}
	}
	return query, args
}

// POST /execs/
func PostExecsHandler(w http.ResponseWriter, r *http.Request) {
	
	// Mutex variables
	mu_exec.Lock()
	defer mu_exec.Unlock()

	var newExecs []models.Exec
	var rawExecs []map[string]interface{}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading Request Body.", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &rawExecs)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid Request Body.").Error(), http.StatusBadRequest)
		return
	}

	// Check whether there are unallowed fields.
	fields := utils.GetFieldNames(models.Exec{})
	allowedFields := make(map[string]struct{})
	for _, field := range fields {
		allowedFields[field] = struct{}{}
	}
	for _, exec := range rawExecs {
		for key := range exec {
			_, ok := allowedFields[key]
			if !ok {
				http.Error(w, "Unacceptable Field found in request.", http.StatusBadRequest)
				return
			}
		}
	}

	err = json.Unmarshal(body, &newExecs)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid Request Body.").Error(), http.StatusBadRequest)
		return
	}
	
	// Check whether all fields are non empty.
	for _, exec := range newExecs {
		err := utils.CheckBlankFields(exec)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}

	// Connect to DB
	addedExecs, err := sqlconnect.PostExecsDBHandler(newExecs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Set the Headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	resp := struct {
		Status string    `json:"status"`
		Count  int       `json:"count"`
		Data   []models.Exec `json:"data"`
	}{
		Status: "success",
		Count:  len(addedExecs),
		Data:   addedExecs,
	}
	json.NewEncoder(w).Encode(resp)
}

// PATCH /execs/{id}
func PatchOneExecHandler(w http.ResponseWriter, r *http.Request) {
	
	// Mutex variables
	mu_exec.Lock()
	defer mu_exec.Unlock()

	idStr := r.PathValue("id")

	// Handle Path Parameters
	execId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid Exec ID.").Error(), http.StatusBadRequest)
		return
	}

	// Get specific patch keys
	var updates map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid Payload Request.").Error(), http.StatusBadRequest)
		return
	}

	// Connect to DB
	existingExec, err := sqlconnect.PatchOneExecDBHandler(execId, updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send back content
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingExec)
}

// PATCH /execs/{id}
func PatchExecsHandler(w http.ResponseWriter, r *http.Request) {
	
	// Mutex variables
	mu_exec.Lock()
	defer mu_exec.Unlock()

	// Get specific patch keys
	var updates []map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid Payload Request.").Error(), http.StatusBadRequest)
		return
	}

	existingExecs, err := sqlconnect.PatchExecsDBHandler(updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Response
	w.WriteHeader(http.StatusNoContent)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingExecs)
}

// DELETE /execs/{id}
func DeleteOneExecHandler(w http.ResponseWriter, r *http.Request) {
	// Mutex variables
	mu_exec.Lock()
	defer mu_exec.Unlock()

	idStr := r.PathValue("id")

	// Handle Path Parameters
	execId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid exec ID.").Error(), http.StatusBadRequest)
		return
	}

	// Connect to DB
	err = sqlconnect.DeleteOneExecDBHandler(execId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	w.Header().Set("Content-Type", "application/json")
	resp := struct{
		Status string `json:"status"`
		ID int `json:"id"`
	}{
		Status: "Exec successfully deleted.",
		ID: execId,
	}
	json.NewEncoder(w).Encode(resp)
}

// POST /execs/login
func LoginExecHandler(w http.ResponseWriter, r *http.Request) {
	var req models.Exec

	// Data Validation
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid Request Body.").Error(), http.StatusBadRequest)
		return
	}

	defer r.Body.Close()
	
	if req.Username == "" || req.Password == "" {
		http.Error(w, utils.ErrorHandler(errors.New("username/password cannot be empty"), "username/password cannot be empty").Error(), http.StatusBadRequest)
		return
	}

	// Search for user if user actually exists
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Error Logging In. Try AGAIN.").Error(), http.StatusInternalServerError)
		return
	}

	defer db.Close()

	exec := models.Exec{}
	err = db.QueryRow("SELECT id, first_name, last_name, email, username, password, inactive_status, role from execs WHERE username=$1", req.Username).Scan(&exec.ID, &exec.FirstName, &exec.LastName, &exec.Email, &exec.Username, &exec.Password, &exec.InactiveStatus, &exec.Role)
	if err == sql.ErrNoRows {
		http.Error(w, utils.ErrorHandler(err, "User not Found.").Error(), http.StatusInternalServerError)
		return
	}
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Error Logging in. TRY AGAIN.").Error(), http.StatusInternalServerError)
		return
	}
	
	// Is user active
	if exec.InactiveStatus {
		http.Error(w, utils.ErrorHandler(errors.New("account is inactive"), "Account is inactive.").Error(), http.StatusForbidden)
	}

	// Verify Password
	parts := strings.Split(exec.Password, ".")
	if len(parts) != 2 {
		http.Error(w, utils.ErrorHandler(errors.New("invalid encoded hash format"), "Some error occured.").Error(), http.StatusForbidden)
		return
	}
	hashedSaltBase64 := parts[0]
	hashedPwdBase64 := parts[1]

	salt, err := base64.StdEncoding.DecodeString(hashedSaltBase64)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Some error occured").Error(), http.StatusForbidden)
		return
	}
	hashedPwd, err := base64.StdEncoding.DecodeString(hashedPwdBase64)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Some error occured").Error(), http.StatusForbidden)
		return
	}

	hash := argon2.IDKey([]byte(req.Password), salt, 1, 64*1024, 4, 32)
	if len(hash) != len(hashedPwd) {
		http.Error(w, utils.ErrorHandler(err, "Incorrect Password, Try AGAIN.").Error(), http.StatusForbidden)
		return
	}
	if subtle.ConstantTimeCompare(hash, hashedPwd) != 1 {
		http.Error(w, utils.ErrorHandler(err, "Incorrect Password, Try AGAIN.").Error(), http.StatusForbidden)
		return
	}

	// Generate Token
	tokenString, err := utils.SignToken(req.ID, req.Username, req.Role)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Could not create Login Token. Internal error.").Error(), http.StatusInternalServerError)
		return
	}

	// Send Token as a response or as a cookie
	http.SetCookie(w, &http.Cookie{
		Name: "Bearer",
		Value: tokenString,
		Path: "/",
		HttpOnly: true,
		Secure: true,
		Expires: time.Now().Add(20 * time.Second),
	})

	// Response Body
	w.Header().Set("Content-Type", "application/json")
	resp := struct{
		Token string `json:"token"`
	}{
		Token: tokenString,
	}
	json.NewEncoder(w).Encode(resp)
}
