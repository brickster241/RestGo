package sqlconnect

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strconv"

	"github.com/brickster241/rest-go/internal/models"
)

func GetTeachersDBHandler(query string, args []interface{}) ([]models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		log.Println("Error connecting DB : ", err)
		return []models.Teacher{}, err
	}
	defer db.Close()

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Println("DB Query Error : ", err)
		return []models.Teacher{}, err
	}

	defer rows.Close()

	// Fetch the teachers
	teacherList := make([]models.Teacher, 0)
	for rows.Next() {
		var teacher models.Teacher
		err = rows.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
		if err != nil {
			log.Println("Error fetching DB row : ", err)
			return []models.Teacher{}, err
		}
		teacherList = append(teacherList, teacher)
	}
	return teacherList, nil
}

func GetOneTeacherDBHandler(teacherId int) (models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		log.Println("Error connecting DB :", err)
		return models.Teacher{}, err
	}
	defer db.Close()

	var tchr models.Teacher
	err = db.QueryRow(fmt.Sprintf("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = %d", teacherId)).Scan(&tchr.ID, &tchr.FirstName, &tchr.LastName, &tchr.Email, &tchr.Class, &tchr.Subject)
	if err == sql.ErrNoRows {
		log.Println("Teacher not found :", err)
		return models.Teacher{}, err
	} else if err != nil {
		log.Println("DB Query error :", err)
		return models.Teacher{}, err
	}
	return tchr, nil
}

func PostTeachersDBHandler(newTeachers []models.Teacher) ([]models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		log.Println("Error connecting DB :", err)
		return nil, err
	}
	defer db.Close()

	// Prepare Query
	stmt, err := db.Prepare("INSERT INTO teachers (first_name, last_name, email, class, subject) VALUES($1,$2,$3,$4,$5)")
	if err != nil {
		log.Println("Error in preparing DB Query :", err)
		return nil, err
	}

	defer stmt.Close()

	addedTeachers := make([]models.Teacher, len(newTeachers))
	for i, newTeacher := range newTeachers {
		_, err := stmt.Exec(newTeacher.FirstName, newTeacher.LastName, newTeacher.Email, newTeacher.Class, newTeacher.Subject)
		if err != nil {
			log.Println("Error Inserting values in DB :", err)
			return nil, err
		}
		addedTeachers[i] = newTeacher
	}
	return addedTeachers, nil
}

func PutOneTeacherDBHandler(teacherId int, updatedTchr models.Teacher) error {
	db, err := ConnectDB()
	if err != nil {
		log.Println("Error connecting DB :", err)
		return err
	}
	defer db.Close()

	var existingTchr models.Teacher
	err = db.QueryRow(fmt.Sprintf("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = %d", teacherId)).Scan(&existingTchr.ID, &existingTchr.FirstName, &existingTchr.LastName, &existingTchr.Email, &existingTchr.Class, &existingTchr.Subject)
	if err == sql.ErrNoRows {
		log.Println("Teacher not found :", err)
		return err
	} else if err != nil {
		log.Println("DB Query error :", err)
		return err
	}

	_, err = db.Exec("UPDATE teachers SET first_name=$1, last_name=$2, email=$3, class=$4, subject=$5 WHERE id=$6", updatedTchr.FirstName, updatedTchr.LastName, updatedTchr.Email, updatedTchr.Class, updatedTchr.Subject, existingTchr.ID)
	if err != nil {
		log.Println("Error updating Teacher in DB :", err)
		return err
	}
	return nil
}

func PatchOneTeacherDBHandler(teacherId int, updates map[string]interface{}) (models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		log.Println("Error connecting DB :", err)
		return models.Teacher{}, err
	}
	defer db.Close()

	var existingTchr models.Teacher
	err = db.QueryRow(fmt.Sprintf("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = %d", teacherId)).Scan(&existingTchr.ID, &existingTchr.FirstName, &existingTchr.LastName, &existingTchr.Email, &existingTchr.Class, &existingTchr.Subject)
	if err == sql.ErrNoRows {
		log.Println("Teacher not found :", err)
		return models.Teacher{}, err
	} else if err != nil {
		log.Println("DB Query error :", err)
		return models.Teacher{}, err
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
		log.Println("Error Updating Teachers :", err)
		return models.Teacher{}, err
	}
	return existingTchr, nil
}

func PatchTeachersDBHandler(updates []map[string]interface{}) ([]models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		log.Println("Error connecting DB :", err)
		return nil, err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		log.Println("Error starting DB transaction :", err)
		return nil, err
	}

	var existingTchrs []models.Teacher
	for _, update := range updates {
		tchrIdStr, ok := update["id"].(string)
		if !ok {
			tx.Rollback()
			log.Println("Invalid Teacher ID :", err)
			return nil, err
		}

		tchrId, err := strconv.Atoi(tchrIdStr)
		if err != nil {
			tx.Rollback()
			log.Println("Cannot convert Teacher ID to int :", err)
			return nil, err
		}

		var existingTchr models.Teacher
		err = tx.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = $1", tchrId).Scan(&existingTchr.ID, &existingTchr.FirstName, &existingTchr.LastName, &existingTchr.Email, &existingTchr.Class, &existingTchr.Subject)
		if err == sql.ErrNoRows {
			tx.Rollback()
			log.Println("Teacher not found :", err)
			return nil, err
		} else if err != nil {
			tx.Rollback()
			log.Println("DB Query error :", err)
			return nil, err
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
						log.Println("Invalid Payload Request :", err)
						return nil, err
					}
					break
				}
			}
		}

		_, err = tx.Exec("UPDATE teachers SET first_name=$1, last_name=$2, email=$3, class=$4, subject=$5 WHERE id=$6", existingTchr.FirstName, existingTchr.LastName, existingTchr.Email, existingTchr.Class, existingTchr.Subject, existingTchr.ID)
		if err != nil {
			tx.Rollback()
			log.Println("Error Updating Teachers :", err)
			return nil, err
		}
		existingTchrs = append(existingTchrs, existingTchr)
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Error committing transaction :", err)
		return nil, err
	}
	return existingTchrs, nil
}

func DeleteOneTeacherDBHandler(teacherId int) error {
	db, err := ConnectDB()
	if err != nil {
		log.Println("Error connecting DB :", err)
		return err
	}
	defer db.Close()

	// Perform the delete operation
	res, err := db.Exec("DELETE FROM teachers WHERE id=$1", teacherId)
	if err != nil {
		log.Println("Error deleting Teacher :", err)
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Println("Error retrieving DB Delete result :", err)
		return err
	}

	// Operation was successful, but no rows affected i.e. invalid ID.
	if rowsAffected == 0 {
		log.Println("Teacher not Found :", err)
		return err
	}
	return nil
}

func DeleteTeachersDBHandler(ids []int) error {
	db, err := ConnectDB()
	if err != nil {
		log.Println("Error connecting DB :", err)
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		log.Println("Error starting DB transaction :", err)
		return err
	}

	// Iterate over all the IDs.
	for _, teacherId := range ids {

		// Perform the delete operation
		res, err := tx.Exec("DELETE FROM teachers WHERE id=$1", teacherId)
		if err != nil {
			tx.Rollback()
			log.Println("Error deleting Teacher :", err)
			return err
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			tx.Rollback()
			log.Println("Error retrieving DB Delete result :", err)
			return err
		}

		// Operation was successful, but no rows affected i.e. invalid ID.
		if rowsAffected == 0 {
			tx.Rollback()
			log.Printf("Teacher %d not Found.\n", teacherId)
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Println("Error committing transaction :", err)
		return err
	}
	return nil
}