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

var mu_tchr = &sync.Mutex{}

// GET teachers/
func GetTeachersHandler(w http.ResponseWriter, r *http.Request) {
	
	query := "SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE 1=1"
	var args []interface{}
	
	// Filter based on different params
	query, args = addQueryFiltersTeacher(r, query, args)

	// Will be of type param:asc or param:desc
	query = applySortingFiltersTeacher(r, query)

	page, limit := utils.GetPaginationParams(r)
	offset := (page - 1) * limit
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	// Connect to DB
	teacherList, totalTeachers, err := sqlconnect.GetTeachersDBHandler(query, args)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := struct {
		Status string    `json:"status"`
		Count  int       `json:"count"`
		Page int `json:"page"`
		PageSize int `json:"page_size"`
		Data   []models.Teacher `json:"data"`
	}{
		Status: "success",
		Count:  totalTeachers,
		Page: page,
		PageSize: limit,
		Data:   teacherList,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GET /teachers/{id}
func GetOneTeacherHandler(w http.ResponseWriter, r *http.Request) {
	
	idStr := r.PathValue("id")
	// Handle Path Parameters
	teacherId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid Teacher ID.").Error(), http.StatusBadRequest)
		return
	}

	// Connect to DB
	tchr, err := sqlconnect.GetOneTeacherDBHandler(teacherId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tchr)
}

// GET /teachers/{id}/students
func GetStudentsByTeacherIDHandler(w http.ResponseWriter, r *http.Request) {
	
	var students []models.Student
	idStr := r.PathValue("id")
	
	// Handle Path Parameters
	teacherId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid Teacher ID.").Error(), http.StatusBadRequest)
		return
	}

	students, err = sqlconnect.GetStudentsByTeachersIDDBHandler(teacherId, students)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	resp := struct {
		Status string `json:"status"`
		TeacherID int `json:"teacherId"`
		Count int `json:"count"`
		Data []models.Student `json:"data"`
	}{
		Status: "success",
		TeacherID: teacherId,
		Count: len(students),
		Data: students,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GET /teachers/{id}/studentcount
func GetStudentCountByTeacherIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	
	// Handle Path Parameters
	teacherId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid Teacher ID.").Error(), http.StatusBadRequest)
		return
	}

	studentCount, err := sqlconnect.GetStudentCountByTeacherIDDBHandler(teacherId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := struct {
		Status string `json:"status"`
		TeacherId int `json:"teacherID"`
		StudentCount int `json:"studentCount"`
	}{
		Status: "success",
		TeacherId: teacherId,
		StudentCount: studentCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func applySortingFiltersTeacher(r *http.Request, query string) string {
	sortParams := r.URL.Query()["sortby"]
	addQuery := " ORDER BY"
	if len(sortParams) > 0 {
		for i, sortParam := range sortParams {
			parts := strings.Split(sortParam, ":")
			if len(parts) != 2 {
				continue
			}

			field, order := parts[0], parts[1]
			if !isValidOrder(order) || !isValidSortFieldTeacher(field) {
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

func isValidSortFieldTeacher(field string) bool {
	validFields := map[string]bool{
		"first_name": true,
		"last_name": true,
		"email": false,
		"class": true,
		"subject": false,
	}
	return validFields[field]
}

func isValidOrder(order string) bool {
	return order == "asc" || order == "desc"
}

func addQueryFiltersTeacher(r *http.Request, query string, args []interface{}) (string, []interface{}) {
	params := []string{
		"first_name",
		"last_name",
		"email",
		"class",
		"subject",
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

// POST /teachers/
func PostTeachersHandler(w http.ResponseWriter, r *http.Request) {
	
	// Mutex variables
	mu_tchr.Lock()
	defer mu_tchr.Unlock()

	var newTeachers []models.Teacher
	var rawTeachers []map[string]interface{}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading Request Body.", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &rawTeachers)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid Request Body.").Error(), http.StatusBadRequest)
		return
	}

	// Check whether there are unallowed fields.
	fields := utils.GetFieldNames(models.Teacher{})
	allowedFields := make(map[string]struct{})
	for _, field := range fields {
		allowedFields[field] = struct{}{}
	}
	for _, teacher := range rawTeachers {
		for key := range teacher {
			_, ok := allowedFields[key]
			if !ok {
				http.Error(w, "Unacceptable Field found in request.", http.StatusBadRequest)
				return
			}
		}
	}

	err = json.Unmarshal(body, &newTeachers)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid Request Body.").Error(), http.StatusBadRequest)
		return
	}
	
	// Check whether all fields are non empty.
	for _, teacher := range newTeachers {
		err := utils.CheckBlankFields(teacher)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}

	// Connect to DB
	addedTeachers, err := sqlconnect.PostTeachersDBHandler(newTeachers)
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
		Data   []models.Teacher `json:"data"`
	}{
		Status: "success",
		Count:  len(addedTeachers),
		Data:   addedTeachers,
	}
	json.NewEncoder(w).Encode(resp)
}

// PUT /teachers/{id}
func PutOneTeacherHandler(w http.ResponseWriter, r *http.Request) {
	
	// Mutex variables
	mu_tchr.Lock()
	defer mu_tchr.Unlock()

	idStr := r.PathValue("id")

	// Handle Path Parameters
	teacherId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid Teacher ID.").Error(), http.StatusBadRequest)
		return
	}

	var updatedTchr models.Teacher
	err = json.NewDecoder(r.Body).Decode(&updatedTchr)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid Teacher Payload.").Error(), http.StatusBadRequest)
		return
	}

	// Connect to DB
	err = sqlconnect.PutOneTeacherDBHandler(teacherId, updatedTchr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the Headers
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTchr)
}

// PATCH /teachers/{id}
func PatchOneTeacherHandler(w http.ResponseWriter, r *http.Request) {
	
	// Mutex variables
	mu_tchr.Lock()
	defer mu_tchr.Unlock()

	idStr := r.PathValue("id")

	// Handle Path Parameters
	teacherId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid Teacher ID.").Error(), http.StatusBadRequest)
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
	existingTchr, err := sqlconnect.PatchOneTeacherDBHandler(teacherId, updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send back content
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingTchr)
}

// PATCH /teachers/{id}
func PatchTeachersHandler(w http.ResponseWriter, r *http.Request) {
	
	// Mutex variables
	mu_tchr.Lock()
	defer mu_tchr.Unlock()

	// Get specific patch keys
	var updates []map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid Payload Request.").Error(), http.StatusBadRequest)
		return
	}

	existingTchrs, err := sqlconnect.PatchTeachersDBHandler(updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Response
	w.WriteHeader(http.StatusNoContent)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingTchrs)
}

// DELETE /teachers/{id}
func DeleteOneTeacherHandler(w http.ResponseWriter, r *http.Request) {
	// Mutex variables
	mu_tchr.Lock()
	defer mu_tchr.Unlock()

	idStr := r.PathValue("id")

	// Handle Path Parameters
	teacherId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid teacher ID.").Error(), http.StatusBadRequest)
		return
	}

	// Connect to DB
	err = sqlconnect.DeleteOneTeacherDBHandler(teacherId)
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
		Status: "Teacher successfully deleted.",
		ID: teacherId,
	}
	json.NewEncoder(w).Encode(resp)
}

// DELETE /teachers/
func DeleteTeachersHandler(w http.ResponseWriter, r *http.Request) {
	// Mutex variables
	mu_tchr.Lock()
	defer mu_tchr.Unlock()

	// Get specific ids to delete
	var ids []int
	err := json.NewDecoder(r.Body).Decode(&ids)
	if err != nil {
		http.Error(w, utils.ErrorHandler(err, "Invalid Payload Request.").Error(), http.StatusBadRequest)
		return
	}

	// Connect to DB
	err = sqlconnect.DeleteTeachersDBHandler(ids)
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
		Status: "Teachers successfully deleted.",
		DeletedIDs: ids,
	}
	json.NewEncoder(w).Encode(resp)

}

