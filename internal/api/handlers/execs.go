package handlers

import (
	"crypto/sha256"
	"encoding/hex"
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
	"github.com/go-mail/mail/v2"
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
	exec, err := sqlconnect.LoginExecDBHandler(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Verify Password
	err = utils.VerifyPassword(exec.Password, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate Token
	tokenString, err := utils.SignToken(exec.ID, exec.Username, exec.Role)
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
		Expires: time.Now().Add(20 * time.Minute),
		SameSite: http.SameSiteStrictMode,
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

// POST /execs/logout
func LogoutExecHandler(w http.ResponseWriter, r *http.Request) {
	// Send Token as a response or as a cookie
	http.SetCookie(w, &http.Cookie{
		Name: "Bearer",
		Value: "",
		Path: "/",
		HttpOnly: true,
		Secure: true,
		Expires: time.Unix(0, 0),
		SameSite: http.SameSiteStrictMode,
	})

	// Response Body
	w.Header().Set("Content-Type", "application/json")
	resp := struct{
		Message string `json:"message"`
	}{
		Message: "Logged Out Successfully !",
	}
	json.NewEncoder(w).Encode(resp)
}

// POST /execs/{id}/updatepassword
func UpdateExecPasswordHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	execId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Exec Id.", http.StatusBadRequest)
		return
	}
	var req models.UpdatePasswordRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid Request Body.").Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.CurrentPassword == "" || req.NewPassword == "" {
		http.Error(w, "please enter password", http.StatusBadRequest)
		return
	}

	execName, execRole, err := sqlconnect.UpdateExecPasswordDBHandler(execId, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	token, err := utils.SignToken(execId, execName, execRole)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Updated Password. Failed to Create Token.").Error(), http.StatusInternalServerError)
		return
	}

	// Send Token as a response or as a cookie
	http.SetCookie(w, &http.Cookie{
		Name: "Bearer",
		Value: token,
		Path: "/",
		HttpOnly: true,
		Secure: true,
		Expires: time.Now().Add(20 * time.Minute),
		SameSite: http.SameSiteStrictMode,
	})

	// Response Body
	w.Header().Set("Content-Type", "application/json")
	resp := models.UpdatePasswordResponse{
		Token: token,
		PasswordUpdated: true,
	}
	json.NewEncoder(w).Encode(resp)

}

// POST /execs/forgotpassword
func ForgotExecPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid Request Body.").Error(), http.StatusBadRequest)
		return
	}
	mins, token, err := sqlconnect.ForgotExecPasswordDBHandler(req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send the reset email
	resetURL := fmt.Sprintf("https://localhost:3000/execs/resetpassword/reset/%s", token)
	msg := fmt.Sprintf("Forgot your password? Reset your password using following link: \n%s\n If you didn't request a password reset, please ignore this email. This link is only valid for %d mins.\n", resetURL, int(mins))
	m := mail.NewMessage()
	m.SetHeader("From", "schooladmin@school.com")
	m.SetHeader("To", req.Email)
	m.SetHeader("Subject", "Your password Reset Link")
	m.SetBody("text/plain", msg)
	d := mail.NewDialer("localhost", 1025, "", "")
	err = d.DialAndSend(m)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Failed to send password Reset Email.").Error(), http.StatusInternalServerError)
		return
	}
	// Respond with Success Message.
	fmt.Fprintf(w, "Password Reset Link sent to %s", req.Email)
}

// POST /execs/resetpassword/reset/{resetcode}
func ResetPasswordHandler(w http.ResponseWriter, r* http.Request) {
	
	token := r.PathValue("resetcode")
	type ResetPasswordRequest struct {
		NewPassword string `json:"new_password"`
		ConfirmPassword string `json:"confirm_password"`
	}

	var req ResetPasswordRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid Request Body.").Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.NewPassword == "" || req.ConfirmPassword == "" {
		http.Error(w, "Passwords should not be blank.", http.StatusBadRequest)
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		http.Error(w, "Passwords should match.", http.StatusBadRequest)
		return
	}

	bytes, err := hex.DecodeString(token)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Internal Server Error.").Error(), http.StatusInternalServerError)
		return
	}

	hashedToken := sha256.Sum256(bytes)
	hashedTokenString := hex.EncodeToString(hashedToken[:])

	// Hash the new Password
	hashedPwd, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Error Resetting Password.").Error(), http.StatusInternalServerError)
		return
	}

	err = sqlconnect.ResetPasswordDBHandler(hashedTokenString, hashedPwd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, "Password Reset Successfully.")
}
