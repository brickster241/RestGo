package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"

	models "github.com/brickster241/rest-go/internal/models"
	"github.com/brickster241/rest-go/internal/repository/sqlconnect"
	"github.com/brickster241/rest-go/pkg/utils"
)

var mu_sdnt = &sync.Mutex{}

// GET students/
func GetStudentsHandler(w http.ResponseWriter, r *http.Request) {
	
	query := "SELECT id, first_name, last_name, email, class FROM students WHERE 1=1"
	var args []interface{}
	
	// Filter based on different params
	query, args = addQueryFilters(r, query, args)

	// Will be of type param:asc or param:desc
	query = applySortingFilters(r, query)

	// Connect to DB
	studentList, err := sqlconnect.GetStudentsDBHandler(query, args)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := struct {
		Status string    `json:"status"`
		Count  int       `json:"count"`
		Data   []models.Student `json:"data"`
	}{
		Status: "success",
		Count:  len(studentList),
		Data:   studentList,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GET /students/{id}
func GetOneStudentHandler(w http.ResponseWriter, r *http.Request) {
	
	idStr := r.PathValue("id")
	// Handle Path Parameters
	studentId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid Student ID.").Error(), http.StatusBadRequest)
		return
	}

	// Connect to DB
	sdnt, err := sqlconnect.GetOneStudentDBHandler(studentId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sdnt)
}

func applySortingFiltersStudent(r *http.Request, query string) string {
	sortParams := r.URL.Query()["sortby"]
	addQuery := " ORDER BY"
	if len(sortParams) > 0 {
		for i, sortParam := range sortParams {
			parts := strings.Split(sortParam, ":")
			if len(parts) != 2 {
				continue
			}

			field, order := parts[0], parts[1]
			if !isValidOrder(order) || !isValidSortFieldStudent(field) {
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

func isValidSortFieldStudent(field string) bool {
	validFields := map[string]bool{
		"first_name": true,
		"last_name": true,
		"email": false,
		"class": true,
	}
	return validFields[field]
}

func addQueryFiltersStudent(r *http.Request, query string, args []interface{}) (string, []interface{}) {
	params := []string{
		"first_name",
		"last_name",
		"email",
		"class",
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

// POST /students/
func PostStudentsHandler(w http.ResponseWriter, r *http.Request) {
	
	// Mutex variables
	mu_sdnt.Lock()
	defer mu_sdnt.Unlock()

	var newStudents []models.Student
	var rawStudents []map[string]interface{}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading Request Body.", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &rawStudents)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid Request Body.").Error(), http.StatusBadRequest)
		return
	}

	// Check whether there are unallowed fields.
	fields := utils.GetFieldNames(models.Student{})
	allowedFields := make(map[string]struct{})
	for _, field := range fields {
		allowedFields[field] = struct{}{}
	}
	for _, student := range rawStudents {
		for key := range student {
			_, ok := allowedFields[key]
			if !ok {
				http.Error(w, "Unacceptable Field found in request.", http.StatusBadRequest)
				return
			}
		}
	}

	err = json.Unmarshal(body, &newStudents)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid Request Body.").Error(), http.StatusBadRequest)
		return
	}
	
	// Check whether all fields are non empty.
	for _, student := range newStudents {
		err := utils.CheckBlankFields(student)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}

	// Connect to DB
	addedStudents, err := sqlconnect.PostStudentsDBHandler(newStudents)
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
		Data   []models.Student `json:"data"`
	}{
		Status: "success",
		Count:  len(addedStudents),
		Data:   addedStudents,
	}
	json.NewEncoder(w).Encode(resp)
}

// PUT /students/{id}
func PutOneStudentHandler(w http.ResponseWriter, r *http.Request) {
	
	// Mutex variables
	mu_sdnt.Lock()
	defer mu_sdnt.Unlock()

	idStr := r.PathValue("id")

	// Handle Path Parameters
	studentId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid Student ID.").Error(), http.StatusBadRequest)
		return
	}

	var updatedSdnt models.Student
	err = json.NewDecoder(r.Body).Decode(&updatedSdnt)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid Student Payload.").Error(), http.StatusBadRequest)
		return
	}

	// Connect to DB
	err = sqlconnect.PutOneStudentDBHandler(studentId, updatedSdnt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the Headers
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedSdnt)
}

// PATCH /students/{id}
func PatchOneStudentHandler(w http.ResponseWriter, r *http.Request) {
	
	// Mutex variables
	mu_sdnt.Lock()
	defer mu_sdnt.Unlock()

	idStr := r.PathValue("id")

	// Handle Path Parameters
	studentId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid Student ID.").Error(), http.StatusBadRequest)
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
	existingSdnt, err := sqlconnect.PatchOneStudentDBHandler(studentId, updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send back content
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingSdnt)
}

// PATCH /students/{id}
func PatchStudentsHandler(w http.ResponseWriter, r *http.Request) {
	
	// Mutex variables
	mu_sdnt.Lock()
	defer mu_sdnt.Unlock()

	// Get specific patch keys
	var updates []map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid Payload Request.").Error(), http.StatusBadRequest)
		return
	}

	existingSdnts, err := sqlconnect.PatchStudentsDBHandler(updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Response
	w.WriteHeader(http.StatusNoContent)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingSdnts)
}

// DELETE /students/{id}
func DeleteOneStudentHandler(w http.ResponseWriter, r *http.Request) {
	// Mutex variables
	mu_sdnt.Lock()
	defer mu_sdnt.Unlock()

	idStr := r.PathValue("id")

	// Handle Path Parameters
	studentId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid student ID.").Error(), http.StatusBadRequest)
		return
	}

	// Connect to DB
	err = sqlconnect.DeleteOneStudentDBHandler(studentId)
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
		Status: "Student successfully deleted.",
		ID: studentId,
	}
	json.NewEncoder(w).Encode(resp)
}

// DELETE /students/
func DeleteStudentsHandler(w http.ResponseWriter, r *http.Request) {
	// Mutex variables
	mu_sdnt.Lock()
	defer mu_sdnt.Unlock()

	// Get specific ids to delete
	var ids []int
	err := json.NewDecoder(r.Body).Decode(&ids)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid Payload Request.").Error(), http.StatusBadRequest)
		return
	}

	// Connect to DB
	err = sqlconnect.DeleteStudentsDBHandler(ids)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Response
	w.WriteHeader(http.StatusNoContent)
	w.Header().Set("Content-Type", "application/json")
	resp := struct{
		Status string `json:"status"`
		DeletedIDs []int `json:"deleted_ids"`
	}{
		Status: "Students successfully deleted.",
		DeletedIDs: ids,
	}
	json.NewEncoder(w).Encode(resp)

}

