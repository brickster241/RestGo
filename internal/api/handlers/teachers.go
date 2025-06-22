package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	models "github.com/brickster241/rest-go/internal/models"
)

var (
	teachers = make(map[int]models.Teacher)
	mu_tchr = &sync.Mutex{}
	nextID = 1
)

// Initialize some dummy data
func init() {
	teachers[nextID] = models.Teacher{
		ID: nextID,
		FirstName: "John",
		LastName: "Doe",
		Class: "9A",
		Subject: "Math",
	}
	nextID++
	teachers[nextID] = models.Teacher{
		ID: nextID,
		FirstName: "Jack",
		LastName: "Smith",
		Class: "10C",
		Subject: "English",
	}
	nextID++
	teachers[nextID] = models.Teacher{
		ID: nextID,
		FirstName: "Jack",
		LastName: "Doe",
		Class: "9B",
		Subject: "Chemistry",
	}
	nextID++
}

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
	mu_tchr.Lock()
	defer mu_tchr.Unlock()

	var newTeachers []models.Teacher
	err := json.NewDecoder(r.Body).Decode(&newTeachers)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	addedTeachers := make([]models.Teacher, len(newTeachers))
	for i, newTeacher := range newTeachers {
		newTeacher.ID = nextID
		teachers[nextID] = newTeacher
		addedTeachers[i] = newTeacher
		nextID++
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