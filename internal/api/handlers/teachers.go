package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	models "github.com/brickster241/rest-go/internal/models"
	"github.com/brickster241/rest-go/internal/repository/sqlconnect"
)

var (
	teachers = make(map[int]models.Teacher)
	mu_tchr = &sync.Mutex{}
)

func getTeachersHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/teachers/")
	idStr := strings.TrimSuffix(path, "/")

	// Path is None, handle query params
	if path == "" {
		// Filter based on First or Last Name
		firstName := r.URL.Query().Get("first_name")
		lastName := r.URL.Query().Get("last_name")

		teacherList := make([]models.Teacher, 0, len(teachers))
		for _, teacher := range teachers {
			if (firstName == "" || firstName == teacher.FirstName) && (lastName == "" || lastName == teacher.LastName) {
				teacherList = append(teacherList, teacher)
			}
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
		err := json.NewEncoder(w).Encode(resp)
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
	tchr, exists := teachers[teacherId]
	if !exists {
		http.Error(w, "Teacher Not Found", http.StatusNotFound)
		return
	}
	err = json.NewEncoder(w).Encode(tchr)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error Fetching teacher with Id %d", teacherId), http.StatusBadRequest)
		return
	}
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