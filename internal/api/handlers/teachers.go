package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"

	models "github.com/brickster241/rest-go/internal/models"
	"github.com/brickster241/rest-go/internal/repository/sqlconnect"
)

var mu_tchr = &sync.Mutex{}

// GET teachers/
func GetTeachersHandler(w http.ResponseWriter, r *http.Request) {
	
	// Connect to DB
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		http.Error(w, "Error connecting DB.", http.StatusInternalServerError)
		return
	}
	defer db.Close()
	
	query := "SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE 1=1"
	var args []interface{}
	
	// Filter based on different params
	query, args = addQueryFilters(r, query, args)

	// Will be of type param:asc or param:desc
	query = applySortingFilters(r, query)

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Println(err)
		http.Error(w, "DB Query error.", http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	// Fetch the teachers
	teacherList := make([]models.Teacher, 0)
	for rows.Next() {
		var teacher models.Teacher
		err = rows.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
		if err != nil {
			http.Error(w, "Error fetching DB row value.", http.StatusInternalServerError)
			return
		}
		teacherList = append(teacherList, teacher)
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
	// Connect to DB
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		http.Error(w, "Error connecting DB.", http.StatusInternalServerError)
		return
	}
	defer db.Close()
	
	idStr := r.PathValue("id")
	// Handle Path Parameters
	teacherId, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid teacher ID.", http.StatusBadRequest)
		return
	}
	var tchr models.Teacher
	err = db.QueryRow(fmt.Sprintf("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = %d", teacherId)).Scan(&tchr.ID, &tchr.FirstName, &tchr.LastName, &tchr.Email, &tchr.Class, &tchr.Subject)
	if err == sql.ErrNoRows {
		http.Error(w, "Teacher not found.", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "DB Query error.", http.StatusInternalServerError)
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
func PostTeacherHandler(w http.ResponseWriter, r *http.Request) {
	
	// Mutex variables
	mu_tchr.Lock()
	defer mu_tchr.Unlock()

	// Connect to DB
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		http.Error(w, "Error connecting DB.", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Prepare Query
	stmt, err := db.Prepare("INSERT INTO teachers (first_name, last_name, email, class, subject) VALUES($1,$2,$3,$4,$5)")
	if err != nil {
		http.Error(w, "Error in preparing DB query.", http.StatusInternalServerError)
		log.Println("Error in preparing DB Query :", err)
		return
	}

	defer stmt.Close()
	
	
	var newTeachers []models.Teacher
	err = json.NewDecoder(r.Body).Decode(&newTeachers)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	addedTeachers := make([]models.Teacher, len(newTeachers))
	for i, newTeacher := range newTeachers {
		_, err := stmt.Exec(newTeacher.FirstName, newTeacher.LastName, newTeacher.Email, newTeacher.Class, newTeacher.Subject)
		if err != nil {
			http.Error(w, "Error Inserting values in DB.", http.StatusInternalServerError)
			return
		}
		addedTeachers[i] = newTeacher
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

	// Connect to DB
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		http.Error(w, "Error connecting DB.", http.StatusInternalServerError)
		return
	}
	defer db.Close()
	idStr := r.PathValue("id")

	// Handle Path Parameters
	teacherId, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid teacher ID.", http.StatusBadRequest)
		return
	}

	var updatedTchr models.Teacher
	err = json.NewDecoder(r.Body).Decode(&updatedTchr)
	if err != nil {
		http.Error(w, "Invalid Teacher Payload.", http.StatusBadRequest)
		return
	}

	var existingTchr models.Teacher
	err = db.QueryRow(fmt.Sprintf("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = %d", teacherId)).Scan(&existingTchr.ID, &existingTchr.FirstName, &existingTchr.LastName, &existingTchr.Email, &existingTchr.Class, &existingTchr.Subject)
	if err == sql.ErrNoRows {
		http.Error(w, "Teacher not found.", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "DB Query error.", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("UPDATE teachers SET first_name=$1, last_name=$2, email=$3, class=$4, subject=$5 WHERE id=$6", updatedTchr.FirstName, updatedTchr.LastName, updatedTchr.Email, updatedTchr.Class, updatedTchr.Subject, existingTchr.ID)
	if err != nil {
		http.Error(w, "Error updating Teacher in DB.", http.StatusInternalServerError)
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

	// Connect to DB
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		http.Error(w, "Error connecting DB.", http.StatusInternalServerError)
		return
	}
	defer db.Close()

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

	var existingTchr models.Teacher
	err = db.QueryRow(fmt.Sprintf("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = %d", teacherId)).Scan(&existingTchr.ID, &existingTchr.FirstName, &existingTchr.LastName, &existingTchr.Email, &existingTchr.Class, &existingTchr.Subject)
	if err == sql.ErrNoRows {
		http.Error(w, "Teacher not found.", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "DB Query error.", http.StatusInternalServerError)
		return
	}

	// apply updates using reflect
	teacherVal := reflect.ValueOf(&existingTchr).Elem()
	teacherValType := teacherVal.Type()
	
	for k, v := range updates {
		for i := 0; i < teacherVal.NumField(); i++ {
			field := teacherValType.Field(i)
			json_field := field.Tag.Get("json")

			// Check whether such key exists in fields and set its value to v
			if json_field == k + ",omitempty"  && teacherVal.Field(i).CanSet() {
				teacherVal.Field(i).Set(reflect.ValueOf(v).Convert(teacherVal.Field(i).Type()))
			}
		}
	}

	_, err = db.Exec("UPDATE teachers SET first_name=$1, last_name=$2, email=$3, class=$4, subject=$5 WHERE id=$6", existingTchr.FirstName, existingTchr.LastName, existingTchr.Email, existingTchr.Class, existingTchr.Subject, existingTchr.ID)
	if err != nil {
		http.Error(w, "Error Updating Teachers.", http.StatusInternalServerError)
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

	// Connect to DB
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		http.Error(w, "Error connecting DB.", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Get specific patch keys
	var updates []map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Invalid Payload Request.", http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Error starting DB transaction.", http.StatusInternalServerError)
		return
	}

	var existingTchrs []models.Teacher
	for _, update := range updates {
		tchrIdStr, ok := update["id"].(string)
		if !ok {
			tx.Rollback()
			http.Error(w, "Invalid Teacher ID.", http.StatusBadRequest)
			return
		}

		tchrId, err := strconv.Atoi(tchrIdStr)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Cannot convert Teacher ID to int.", http.StatusBadRequest)
			return
		}

		var existingTchr models.Teacher
		err = tx.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = $1", tchrId).Scan(&existingTchr.ID, &existingTchr.FirstName, &existingTchr.LastName, &existingTchr.Email, &existingTchr.Class, &existingTchr.Subject)
		if err == sql.ErrNoRows {
			tx.Rollback()
			http.Error(w, "Teacher not found.", http.StatusNotFound)
			return
		} else if err != nil {
			tx.Rollback()
			http.Error(w, "DB Query error.", http.StatusInternalServerError)
			return
		}

		// apply updates using reflect
		teacherVal := reflect.ValueOf(&existingTchr).Elem()
		teacherValType := teacherVal.Type()
		
		for k, v := range update {
			if k == "id" {
				continue	// Skip the id field.
			}
			for i := 0; i < teacherVal.NumField(); i++ {
				field := teacherValType.Field(i)
				json_field := field.Tag.Get("json")

				// Check whether such key exists in fields and set its value to v
				if json_field == k + ",omitempty"  && teacherVal.Field(i).CanSet() {
					if reflect.ValueOf(v).Type().ConvertibleTo(teacherVal.Field(i).Type()) {
						teacherVal.Field(i).Set(reflect.ValueOf(v).Convert(teacherVal.Field(i).Type()))
					} else {
						tx.Rollback()
						http.Error(w, "Invalid Payload Request.", http.StatusBadRequest)
						return
					}
					break
				}
			}
		}

		_, err = tx.Exec("UPDATE teachers SET first_name=$1, last_name=$2, email=$3, class=$4, subject=$5 WHERE id=$6", existingTchr.FirstName, existingTchr.LastName, existingTchr.Email, existingTchr.Class, existingTchr.Subject, existingTchr.ID)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Error Updating Teachers.", http.StatusInternalServerError)
			return
		}
		existingTchrs = append(existingTchrs, existingTchr)
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, "Error committing transaction.", http.StatusInternalServerError)
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

	// Connect to DB
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		http.Error(w, "Error connecting DB.", http.StatusInternalServerError)
		return
	}
	defer db.Close()
	idStr := r.PathValue("id")

	// Handle Path Parameters
	teacherId, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid teacher ID.", http.StatusBadRequest)
		return
	}

	// Perform the delete operation
	res, err := db.Exec("DELETE FROM teachers WHERE id=$1", teacherId)
	if err != nil {
		http.Error(w, "Error deleting Teacher.", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		http.Error(w, "Error retrieving DB Delete result.", http.StatusInternalServerError)
		return
	}

	// Operation was successful, but no rows affected i.e. invalid ID.
	if rowsAffected == 0 {
		http.Error(w, "Teacher not Found.", http.StatusNotFound)
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

	// Connect to DB
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		http.Error(w, "Error connecting DB.", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Get specific ids to delete
	var ids []int
	err = json.NewDecoder(r.Body).Decode(&ids)
	if err != nil {
		http.Error(w, "Invalid Payload Request.", http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Error starting DB transaction.", http.StatusInternalServerError)
		return
	}

	// Iterate over all the IDs.
	for _, teacherId := range ids {

		// Perform the delete operation
		res, err := tx.Exec("DELETE FROM teachers WHERE id=$1", teacherId)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Error deleting Teacher.", http.StatusInternalServerError)
			return
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			tx.Rollback()
			http.Error(w, "Error retrieving DB Delete result.", http.StatusInternalServerError)
			return
		}

		// Operation was successful, but no rows affected i.e. invalid ID.
		if rowsAffected == 0 {
			tx.Rollback()
			http.Error(w, fmt.Sprintf("Teacher %d not Found.", teacherId), http.StatusNotFound)
			return
		}
	}
	err = tx.Commit()
	if err != nil {
		http.Error(w, "Error committing transaction.", http.StatusInternalServerError)
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

