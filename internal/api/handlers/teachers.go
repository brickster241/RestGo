package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	models "github.com/brickster241/rest-go/internal/models"
	"github.com/brickster241/rest-go/internal/repository/sqlconnect"
)

var mu_tchr = &sync.Mutex{}

// GET teachers/
func GetTeachersHandler(w http.ResponseWriter, r *http.Request) {
	
	query := "SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE 1=1"
	var args []interface{}
	
	// Filter based on different params
	query, args = addQueryFilters(r, query, args)

	// Will be of type param:asc or param:desc
	query = applySortingFilters(r, query)

	// Connect to DB
	teacherList, err := sqlconnect.GetTeachersDBHandler(query, args)
	if err != nil {
		log.Println("Error fetching Teachers : ", err)
		http.Error(w, "Error fetching Teachers", http.StatusInternalServerError)
		return
	}

	resp := struct {
		Status string    `json:"status"`
		Count  int       `json:"count"`
		Data   []models.Teacher `json:"data"`
	}{
		Status: "success",
		Count:  len(teacherList),
		Data:   teacherList,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, "Error occured while fetching Teacher List.", http.StatusBadRequest)
	}
}

// GET /teachers/{id}
func GetOneTeacherHandler(w http.ResponseWriter, r *http.Request) {
	
	idStr := r.PathValue("id")
	// Handle Path Parameters
	teacherId, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println("Invalid Teacher ID :", err)
		http.Error(w, "Invalid teacher ID.", http.StatusBadRequest)
		return
	}

	// Connect to DB
	tchr, err := sqlconnect.GetOneTeacherDBHandler(teacherId)
	if err != nil {
		log.Println("Error fetching Teacher :", err)
		http.Error(w, "Error Fetching Teacher.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(tchr)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error Fetching teacher with Id %d", teacherId), http.StatusBadRequest)
		return
	}
}

func applySortingFilters(r *http.Request, query string) string {
	sortParams := r.URL.Query()["sortby"]
	addQuery := " ORDER BY"
	if len(sortParams) > 0 {
		for i, sortParam := range sortParams {
			parts := strings.Split(sortParam, ":")
			if len(parts) != 2 {
				continue
			}

			field, order := parts[0], parts[1]
			if !isValidOrder(order) || !isValidSortField(field) {
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

func isValidSortField(field string) bool {
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

func addQueryFilters(r *http.Request, query string, args []interface{}) (string, []interface{}) {
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
	err := json.NewDecoder(r.Body).Decode(&newTeachers)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	// Connect to DB
	addedTeachers, err := sqlconnect.PostTeachersDBHandler(newTeachers)
	if err != nil {
		log.Println("Error adding Teachers : ", err)
		http.Error(w, "Error adding Teachers.", http.StatusInternalServerError)
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

	log.Printf("Added %d Teachers.", len(addedTeachers))
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
		log.Println("Invalid teacher ID :", err)
		http.Error(w, "Invalid teacher ID.", http.StatusBadRequest)
		return
	}

	var updatedTchr models.Teacher
	err = json.NewDecoder(r.Body).Decode(&updatedTchr)
	if err != nil {
		http.Error(w, "Invalid Teacher Payload.", http.StatusBadRequest)
		return
	}

	// Connect to DB
	err = sqlconnect.PutOneTeacherDBHandler(teacherId, updatedTchr)
	if err != nil {
		log.Println("Error updating Teacher : ", err)
		http.Error(w, "Error updating Teacher", http.StatusInternalServerError)
		return
	}

	// Set the Headers
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Updated Teacher with id : %d", teacherId)
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
		log.Println(err)
		http.Error(w, "Invalid teacher ID.", http.StatusBadRequest)
		return
	}

	// Get specific patch keys
	var updates map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Invalid Payload Request.", http.StatusBadRequest)
		return
	}

	// Connect to DB
	existingTchr, err := sqlconnect.PatchOneTeacherDBHandler(teacherId, updates)
	if err != nil {
		log.Println("Error updating Teacher :", err)
		http.Error(w, "Error updating Teacher", http.StatusInternalServerError)
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
		http.Error(w, "Invalid Payload Request.", http.StatusBadRequest)
		return
	}

	existingTchrs, err := sqlconnect.PatchTeachersDBHandler(updates)
	if err != nil {
		log.Println("Error updating Teachers :", err)
		http.Error(w, "Error updating Teachers.", http.StatusInternalServerError)
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
		log.Println(err)
		http.Error(w, "Invalid teacher ID.", http.StatusBadRequest)
		return
	}

	// Connect to DB
	err = sqlconnect.DeleteOneTeacherDBHandler(teacherId)
	if err != nil {
		log.Println("Error Deleting teacher :", err)
		http.Error(w, "Error Deleting teacher.", http.StatusInternalServerError)
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
		http.Error(w, "Invalid Payload Request.", http.StatusBadRequest)
		return
	}

	// Connect to DB
	err = sqlconnect.DeleteTeachersDBHandler(ids)
	if err != nil {
		log.Println("Error Deleting teachers :", err)
		http.Error(w, "Error Deleting teachers.", http.StatusInternalServerError)
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

