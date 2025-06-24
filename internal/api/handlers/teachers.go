package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	models "github.com/brickster241/rest-go/internal/models"
	"github.com/brickster241/rest-go/internal/repository/sqlconnect"
)

var mu_tchr = &sync.Mutex{}

func getTeachersHandler(w http.ResponseWriter, r *http.Request) {
	
	// Connect to DB
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		http.Error(w, "Error connecting DB.", http.StatusInternalServerError)
		return
	}
	defer db.Close()
	
	path := strings.TrimPrefix(r.URL.Path, "/teachers/")
	idStr := strings.TrimSuffix(path, "/")

	// Path is None
	if path == "" {
		query := "SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE 1=1"
		var args []interface{}
		
		// Filter based on different params
		query, args = addQueryFilters(r, query, args)

		// Will be of type param:asc or param:desc
		query = applySortingFilters(r, query)

		rows, err := db.Query(query, args...)
		if err != nil {
			fmt.Println(err)
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
		return
	}

	// Handle Path Parameters
	teacherId, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println(err)
		return
	}
	var tchr models.Teacher
	err = db.QueryRow(fmt.Sprintf("SELECT id, first_name, last_name, email, class, subject FROM teachers where id = %d", teacherId)).Scan(&tchr.ID, &tchr.FirstName, &tchr.LastName, &tchr.Email, &tchr.Class, &tchr.Subject)
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

func addTeacherHandler(w http.ResponseWriter, r *http.Request) {
	
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
		fmt.Println("Error in preparing DB Query :", err)
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

	fmt.Printf("Added %d Teachers.", len(addedTeachers))
	json.NewEncoder(w).Encode(resp)
}

func TeachersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTeachersHandler(w, r)
	case http.MethodPost:
		addTeacherHandler(w, r)
	case http.MethodPut:
		fmt.Fprintf(w, "Welcome to Teachers Page... : PUT Method")
	case http.MethodPatch:
		fmt.Fprintf(w, "Welcome to Teachers Page... : PATCH Method")
	case http.MethodDelete:
		fmt.Fprintf(w, "Welcome to Teachers Page... : DELETE Method")
	default:
		fmt.Fprintf(w, "Invalid Request : %v", r.Method)
	}
}