package sqlconnect

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"

	"github.com/brickster241/rest-go/internal/models"
	"github.com/brickster241/rest-go/pkg/utils"
)

func GetStudentsDBHandler(query string, args []interface{}) ([]models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return []models.Student{}, utils.ErrorHandler(err, "Error connecting DB.")
	}
	defer db.Close()

	rows, err := db.Query(query, args...)
	if err != nil {
		return []models.Student{}, utils.ErrorHandler(err, "Error fetching Students.")
	}

	defer rows.Close()

	// Fetch the students
	studentList := make([]models.Student, 0)
	for rows.Next() {
		var student models.Student
		err = rows.Scan(&student.ID, &student.FirstName, &student.LastName, &student.Email, &student.Class)
		if err != nil {
			return []models.Student{}, utils.ErrorHandler(err, "Error fetching Students.")
		}
		studentList = append(studentList, student)
	}
	return studentList, nil
}

func GetOneStudentDBHandler(studentId int) (models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "Error connecting DB.")
	}
	defer db.Close()

	var sdnt models.Student
	err = db.QueryRow(fmt.Sprintf("SELECT id, first_name, last_name, email, class FROM students WHERE id = %d", studentId)).Scan(&sdnt.ID, &sdnt.FirstName, &sdnt.LastName, &sdnt.Email, &sdnt.Class)
	if err == sql.ErrNoRows {
		return models.Student{}, utils.ErrorHandler(err, fmt.Sprintf("Error fetching Student %d.", studentId))
	} else if err != nil {
		return models.Student{}, utils.ErrorHandler(err, fmt.Sprintf("Error fetching Student %d.", studentId))
	}
	return sdnt, nil
}

func PostStudentsDBHandler(newStudents []models.Student) ([]models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting DB.")
	}
	defer db.Close()

	// Prepare Query
	tx, err := db.Begin()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error Adding Students.")
	} 
	// stmt, err := db.Prepare("INSERT INTO students (first_name, last_name, email, class, subject) VALUES($1,$2,$3,$4,$5)")
	stmt, err := tx.Prepare(generateInsertQuery("students", models.Student{}))
	if err != nil {
		tx.Rollback()
		return nil, utils.ErrorHandler(err, "Error Adding students.")
	}

	defer stmt.Close()

	addedStudents := make([]models.Student, len(newStudents))
	for i, newStudent := range newStudents {
		values := getStructValues(newStudent)
		// _, err := stmt.Exec(newStudent.FirstName, newStudent.LastName, newStudent.Email, newStudent.Class)
		_, err := stmt.Exec(values...)
		if err != nil {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "Error Adding students.")
		}
		addedStudents[i] = newStudent
	}

	err = tx.Commit()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error Adding Students.")
	}
	return addedStudents, nil
}

func PutOneStudentDBHandler(studentId int, updatedSdnt models.Student) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "Error connecting DB.")
	}
	defer db.Close()

	var existingSdnt models.Student
	err = db.QueryRow(fmt.Sprintf("SELECT id, first_name, last_name, email, class FROM students WHERE id = %d", studentId)).Scan(&existingSdnt.ID, &existingSdnt.FirstName, &existingSdnt.LastName, &existingSdnt.Email, &existingSdnt.Class)
	if err == sql.ErrNoRows {
		return utils.ErrorHandler(err, fmt.Sprintf("Error updating Student %d.", studentId))
	} else if err != nil {
		return utils.ErrorHandler(err, fmt.Sprintf("Error updating Student %d.", studentId))
	}

	_, err = db.Exec("UPDATE students SET first_name=$1, last_name=$2, email=$3, class=$4 WHERE id=$5", updatedSdnt.FirstName, updatedSdnt.LastName, updatedSdnt.Email, updatedSdnt.Class, existingSdnt.ID)
	if err != nil {
		return utils.ErrorHandler(err, fmt.Sprintf("Error updating Student %d.", studentId))
	}
	return nil
}

func PatchOneStudentDBHandler(studentId int, updates map[string]interface{}) (models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "Error connecting DB.")
	}
	defer db.Close()

	var existingSdnt models.Student
	err = db.QueryRow(fmt.Sprintf("SELECT id, first_name, last_name, email, class FROM students WHERE id = %d", studentId)).Scan(&existingSdnt.ID, &existingSdnt.FirstName, &existingSdnt.LastName, &existingSdnt.Email, &existingSdnt.Class)
	if err == sql.ErrNoRows {
		return models.Student{}, utils.ErrorHandler(err, fmt.Sprintf("Error updating Student %d.", studentId))
	} else if err != nil {
		return models.Student{}, utils.ErrorHandler(err, fmt.Sprintf("Error updating Student %d.", studentId))
	}

	// Apply updates using reflect
	studentVal := reflect.ValueOf(&existingSdnt).Elem()
	studentValType := studentVal.Type()

	for k, v := range updates {
		for i := 0; i < studentVal.NumField(); i++ {
			field := studentValType.Field(i)
			json_field := field.Tag.Get("json")

			// Check whether such key exists in fields and set its value to v
			if json_field == k+",omitempty" && studentVal.Field(i).CanSet() {
				studentVal.Field(i).Set(reflect.ValueOf(v).Convert(studentVal.Field(i).Type()))
			}
		}
	}

	_, err = db.Exec("UPDATE students SET first_name=$1, last_name=$2, email=$3, class=$4 WHERE id=$5", existingSdnt.FirstName, existingSdnt.LastName, existingSdnt.Email, existingSdnt.Class, existingSdnt.ID)
	if err != nil {
		return models.Student{}, utils.ErrorHandler(err, fmt.Sprintf("Error updating Student %d.", studentId))
	}
	return existingSdnt, nil
}

func PatchStudentsDBHandler(updates []map[string]interface{}) ([]models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting DB.")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error updating Students.")
	}

	var existingSdnts []models.Student
	for _, update := range updates {
		sdntIdStr, ok := update["id"].(string)
		if !ok {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "Error updating Students.")
		}

		sdntId, err := strconv.Atoi(sdntIdStr)
		if err != nil {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "Error updating Students.")
		}

		var existingSdnt models.Student
		err = tx.QueryRow("SELECT id, first_name, last_name, email, class FROM students WHERE id = $1", sdntId).Scan(&existingSdnt.ID, &existingSdnt.FirstName, &existingSdnt.LastName, &existingSdnt.Email, &existingSdnt.Class)
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "Error updating Students.")
		} else if err != nil {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "Error updating Students.")
		}

		// apply updates using reflect
		studentVal := reflect.ValueOf(&existingSdnt).Elem()
		studentValType := studentVal.Type()

		for k, v := range update {
			if k == "id" {
				continue // Skip the id field.
			}
			for i := 0; i < studentVal.NumField(); i++ {
				field := studentValType.Field(i)
				json_field := field.Tag.Get("json")

				// Check whether such key exists in fields and set its value to v
				if json_field == k+",omitempty" && studentVal.Field(i).CanSet() {
					if reflect.ValueOf(v).Type().ConvertibleTo(studentVal.Field(i).Type()) {
						studentVal.Field(i).Set(reflect.ValueOf(v).Convert(studentVal.Field(i).Type()))
					} else {
						tx.Rollback()
						return nil, utils.ErrorHandler(err, "Error updating Students.")
					}
					break
				}
			}
		}

		_, err = tx.Exec("UPDATE students SET first_name=$1, last_name=$2, email=$3, class=$4 WHERE id=$5", existingSdnt.FirstName, existingSdnt.LastName, existingSdnt.Email, existingSdnt.Class, existingSdnt.ID)
		if err != nil {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "Error updating Students.")
		}
		existingSdnts = append(existingSdnts, existingSdnt)
	}

	err = tx.Commit()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error updating Students.")
	}
	return existingSdnts, nil
}

func DeleteOneStudentDBHandler(studentId int) error {
	db, err := ConnectDB()
	if err != nil {
		// log.Println("Error connecting DB :", err)
		return utils.ErrorHandler(err, "Error connecting DB.")
	}
	defer db.Close()

	// Perform the delete operation
	res, err := db.Exec("DELETE FROM students WHERE id=$1", studentId)
	if err != nil {
		return utils.ErrorHandler(err, fmt.Sprintf("Error deleting Student %d.", studentId))
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return utils.ErrorHandler(err, fmt.Sprintf("Error deleting Student %d.", studentId))
	}

	// Operation was successful, but no rows affected i.e. invalid ID.
	if rowsAffected == 0 {
		return utils.ErrorHandler(err, fmt.Sprintf("Error deleting Student %d.", studentId))
	}
	return nil
}

func DeleteStudentsDBHandler(ids []int) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "Error connecting DB.")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return utils.ErrorHandler(err, "Error deleting Students.")
	}

	// Iterate over all the IDs.
	for _, studentId := range ids {

		// Perform the delete operation
		res, err := tx.Exec("DELETE FROM students WHERE id=$1", studentId)
		if err != nil {
			tx.Rollback()
			return utils.ErrorHandler(err, "Error deleting Students.")
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			tx.Rollback()
			return utils.ErrorHandler(err, "Error deleting Students.")
		}

		// Operation was successful, but no rows affected i.e. invalid ID.
		if rowsAffected == 0 {
			tx.Rollback()
			return utils.ErrorHandler(err, "Error deleting Students.")
		}
	}
	err = tx.Commit()
	if err != nil {
		return utils.ErrorHandler(err, "Error deleting Students.")
	}
	return nil
}
