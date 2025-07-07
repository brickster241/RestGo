package sqlconnect

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"

	"github.com/brickster241/rest-go/internal/models"
	"github.com/brickster241/rest-go/pkg/utils"
)

func GetTeachersDBHandler(query string, args []interface{}) ([]models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return []models.Teacher{}, utils.ErrorHandler(err, "Error connecting DB.")
	}
	defer db.Close()

	rows, err := db.Query(query, args...)
	if err != nil {
		return []models.Teacher{}, utils.ErrorHandler(err, "Error fetching Teachers.")
	}

	defer rows.Close()

	// Fetch the teachers
	teacherList := make([]models.Teacher, 0)
	for rows.Next() {
		var teacher models.Teacher
		err = rows.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
		if err != nil {
			return []models.Teacher{}, utils.ErrorHandler(err, "Error fetching Teachers.")
		}
		teacherList = append(teacherList, teacher)
	}
	return teacherList, nil
}

func GetOneTeacherDBHandler(teacherId int) (models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, "Error connecting DB.")
	}
	defer db.Close()

	var tchr models.Teacher
	err = db.QueryRow(fmt.Sprintf("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = %d", teacherId)).Scan(&tchr.ID, &tchr.FirstName, &tchr.LastName, &tchr.Email, &tchr.Class, &tchr.Subject)
	if err == sql.ErrNoRows {
		return models.Teacher{}, utils.ErrorHandler(err, fmt.Sprintf("Error fetching Teacher %d.", teacherId))
	} else if err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, fmt.Sprintf("Error fetching Teacher %d.", teacherId))
	}
	return tchr, nil
}

func PostTeachersDBHandler(newTeachers []models.Teacher) ([]models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting DB.")
	}
	defer db.Close()

	// Prepare Query
	tx, err := db.Begin()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error Adding Teachers.")
	} 
	// stmt, err := db.Prepare("INSERT INTO teachers (first_name, last_name, email, class, subject) VALUES($1,$2,$3,$4,$5)")
	stmt, err := tx.Prepare(generateInsertQuery("teachers", models.Teacher{}))
	if err != nil {
		tx.Rollback()
		return nil, utils.ErrorHandler(err, "Error Adding teachers.")
	}

	defer stmt.Close()

	addedTeachers := make([]models.Teacher, len(newTeachers))
	for i, newTeacher := range newTeachers {
		values := getStructValues(newTeacher)
		// _, err := stmt.Exec(newTeacher.FirstName, newTeacher.LastName, newTeacher.Email, newTeacher.Class, newTeacher.Subject)
		_, err := stmt.Exec(values...)
		if err != nil {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "Error Adding teachers.")
		}
		addedTeachers[i] = newTeacher
	}

	err = tx.Commit()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error Adding Teachers.")
	}
	return addedTeachers, nil
}

func PutOneTeacherDBHandler(teacherId int, updatedTchr models.Teacher) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "Error connecting DB.")
	}
	defer db.Close()

	var existingTchr models.Teacher
	err = db.QueryRow(fmt.Sprintf("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = %d", teacherId)).Scan(&existingTchr.ID, &existingTchr.FirstName, &existingTchr.LastName, &existingTchr.Email, &existingTchr.Class, &existingTchr.Subject)
	if err == sql.ErrNoRows {
		return utils.ErrorHandler(err, fmt.Sprintf("Error updating Teacher %d.", teacherId))
	} else if err != nil {
		return utils.ErrorHandler(err, fmt.Sprintf("Error updating Teacher %d.", teacherId))
	}

	_, err = db.Exec("UPDATE teachers SET first_name=$1, last_name=$2, email=$3, class=$4, subject=$5 WHERE id=$6", updatedTchr.FirstName, updatedTchr.LastName, updatedTchr.Email, updatedTchr.Class, updatedTchr.Subject, existingTchr.ID)
	if err != nil {
		return utils.ErrorHandler(err, fmt.Sprintf("Error updating Teacher %d.", teacherId))
	}
	return nil
}

func PatchOneTeacherDBHandler(teacherId int, updates map[string]interface{}) (models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, "Error connecting DB.")
	}
	defer db.Close()

	var existingTchr models.Teacher
	err = db.QueryRow(fmt.Sprintf("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = %d", teacherId)).Scan(&existingTchr.ID, &existingTchr.FirstName, &existingTchr.LastName, &existingTchr.Email, &existingTchr.Class, &existingTchr.Subject)
	if err == sql.ErrNoRows {
		return models.Teacher{}, utils.ErrorHandler(err, fmt.Sprintf("Error updating Teacher %d.", teacherId))
	} else if err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, fmt.Sprintf("Error updating Teacher %d.", teacherId))
	}

	// Apply updates using reflect
	teacherVal := reflect.ValueOf(&existingTchr).Elem()
	teacherValType := teacherVal.Type()

	for k, v := range updates {
		for i := 0; i < teacherVal.NumField(); i++ {
			field := teacherValType.Field(i)
			json_field := field.Tag.Get("json")

			// Check whether such key exists in fields and set its value to v
			if json_field == k+",omitempty" && teacherVal.Field(i).CanSet() {
				teacherVal.Field(i).Set(reflect.ValueOf(v).Convert(teacherVal.Field(i).Type()))
			}
		}
	}

	_, err = db.Exec("UPDATE teachers SET first_name=$1, last_name=$2, email=$3, class=$4, subject=$5 WHERE id=$6", existingTchr.FirstName, existingTchr.LastName, existingTchr.Email, existingTchr.Class, existingTchr.Subject, existingTchr.ID)
	if err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, fmt.Sprintf("Error updating Teacher %d.", teacherId))
	}
	return existingTchr, nil
}

func PatchTeachersDBHandler(updates []map[string]interface{}) ([]models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting DB.")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error updating Teachers.")
	}

	var existingTchrs []models.Teacher
	for _, update := range updates {
		tchrIdStr, ok := update["id"].(string)
		if !ok {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "Error updating Teachers.")
		}

		tchrId, err := strconv.Atoi(tchrIdStr)
		if err != nil {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "Error updating Teachers.")
		}

		var existingTchr models.Teacher
		err = tx.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = $1", tchrId).Scan(&existingTchr.ID, &existingTchr.FirstName, &existingTchr.LastName, &existingTchr.Email, &existingTchr.Class, &existingTchr.Subject)
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "Error updating Teachers.")
		} else if err != nil {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "Error updating Teachers.")
		}

		// apply updates using reflect
		teacherVal := reflect.ValueOf(&existingTchr).Elem()
		teacherValType := teacherVal.Type()

		for k, v := range update {
			if k == "id" {
				continue // Skip the id field.
			}
			for i := 0; i < teacherVal.NumField(); i++ {
				field := teacherValType.Field(i)
				json_field := field.Tag.Get("json")

				// Check whether such key exists in fields and set its value to v
				if json_field == k+",omitempty" && teacherVal.Field(i).CanSet() {
					if reflect.ValueOf(v).Type().ConvertibleTo(teacherVal.Field(i).Type()) {
						teacherVal.Field(i).Set(reflect.ValueOf(v).Convert(teacherVal.Field(i).Type()))
					} else {
						tx.Rollback()
						return nil, utils.ErrorHandler(err, "Error updating Teachers.")
					}
					break
				}
			}
		}

		_, err = tx.Exec("UPDATE teachers SET first_name=$1, last_name=$2, email=$3, class=$4, subject=$5 WHERE id=$6", existingTchr.FirstName, existingTchr.LastName, existingTchr.Email, existingTchr.Class, existingTchr.Subject, existingTchr.ID)
		if err != nil {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "Error updating Teachers.")
		}
		existingTchrs = append(existingTchrs, existingTchr)
	}

	err = tx.Commit()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error updating Teachers.")
	}
	return existingTchrs, nil
}

func DeleteOneTeacherDBHandler(teacherId int) error {
	db, err := ConnectDB()
	if err != nil {
		// log.Println("Error connecting DB :", err)
		return utils.ErrorHandler(err, "Error connecting DB.")
	}
	defer db.Close()

	// Perform the delete operation
	res, err := db.Exec("DELETE FROM teachers WHERE id=$1", teacherId)
	if err != nil {
		return utils.ErrorHandler(err, fmt.Sprintf("Error deleting Teacher %d.", teacherId))
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return utils.ErrorHandler(err, fmt.Sprintf("Error deleting Teacher %d.", teacherId))
	}

	// Operation was successful, but no rows affected i.e. invalid ID.
	if rowsAffected == 0 {
		return utils.ErrorHandler(err, fmt.Sprintf("Error deleting Teacher %d.", teacherId))
	}
	return nil
}

func DeleteTeachersDBHandler(ids []int) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "Error connecting DB.")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return utils.ErrorHandler(err, "Error deleting Teachers.")
	}

	// Iterate over all the IDs.
	for _, teacherId := range ids {

		// Perform the delete operation
		res, err := tx.Exec("DELETE FROM teachers WHERE id=$1", teacherId)
		if err != nil {
			tx.Rollback()
			return utils.ErrorHandler(err, "Error deleting Teachers.")
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			tx.Rollback()
			return utils.ErrorHandler(err, "Error deleting Teachers.")
		}

		// Operation was successful, but no rows affected i.e. invalid ID.
		if rowsAffected == 0 {
			tx.Rollback()
			return utils.ErrorHandler(err, "Error deleting Teachers.")
		}
	}
	err = tx.Commit()
	if err != nil {
		return utils.ErrorHandler(err, "Error deleting Teachers.")
	}
	return nil
}