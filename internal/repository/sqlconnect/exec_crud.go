package sqlconnect

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/brickster241/rest-go/internal/models"
	"github.com/brickster241/rest-go/pkg/utils"
)

func GetExecsDBHandler(query string, args []interface{}) ([]models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		return []models.Exec{}, utils.ErrorHandler(err, "Error connecting DB.")
	}
	defer db.Close()

	rows, err := db.Query(query, args...)
	if err != nil {
		return []models.Exec{}, utils.ErrorHandler(err, "Error fetching Execs.")
	}

	defer rows.Close()

	// Fetch the execs
	execList := make([]models.Exec, 0)
	for rows.Next() {
		var exec models.Exec
		err = rows.Scan(&exec.ID, &exec.FirstName, &exec.LastName, &exec.Email, &exec.Username, &exec.UserCreatedAt, &exec.InactiveStatus, &exec.Role)
		if err != nil {
			return []models.Exec{}, utils.ErrorHandler(err, "Error fetching Execs.")
		}
		execList = append(execList, exec)
	}
	return execList, nil
}

func GetOneExecDBHandler(execId int) (models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Error connecting DB.")
	}
	defer db.Close()

	var exec models.Exec
	err = db.QueryRow(fmt.Sprintf("SELECT id, first_name, last_name, email, username, user_created_at, inactive_status, role FROM execs WHERE id = %d", execId)).Scan(&exec.ID, &exec.FirstName, &exec.LastName, &exec.Email, &exec.Username, &exec.UserCreatedAt, &exec.InactiveStatus, &exec.Role)
	if err == sql.ErrNoRows {
		return models.Exec{}, utils.ErrorHandler(err, fmt.Sprintf("Error fetching Exec %d.", execId))
	} else if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, fmt.Sprintf("Error fetching Exec %d.", execId))
	}
	return exec, nil
}

func PostExecsDBHandler(newExecs []models.Exec) ([]models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting DB.")
	}
	defer db.Close()

	// Prepare Query
	tx, err := db.Begin()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error Adding Execs.")
	} 
	stmt, err := tx.Prepare(generateInsertQuery("execs", models.Exec{}))
	if err != nil {
		tx.Rollback()
		return nil, utils.ErrorHandler(err, "Error Adding execs.")
	}

	defer stmt.Close()

	addedExecs := make([]models.Exec, len(newExecs))
	for i, newExec := range newExecs {

		hashPassword, err := utils.HashPassword(newExec.Password)
		if err != nil {
			return nil, err
		}
		newExec.Password = hashPassword

		values := getStructValues(newExec)
		_, err = stmt.Exec(values...)
		if err != nil {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "Error Adding execs.")
		}
		addedExecs[i] = newExec
	}

	err = tx.Commit()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error Adding Execs.")
	}
	return addedExecs, nil
}

func PatchOneExecDBHandler(execId int, updates map[string]interface{}) (models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Error connecting DB.")
	}
	defer db.Close()

	var existingExec models.Exec
	err = db.QueryRow(fmt.Sprintf("SELECT id, first_name, last_name, email, username FROM execs WHERE id = %d", execId)).Scan(&existingExec.ID, &existingExec.FirstName, &existingExec.LastName, &existingExec.Email, &existingExec.Username)
	if err == sql.ErrNoRows {
		return models.Exec{}, utils.ErrorHandler(err, fmt.Sprintf("Error updating Exec %d.", execId))
	} else if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, fmt.Sprintf("Error updating Exec %d.", execId))
	}

	// Apply updates using reflect
	execVal := reflect.ValueOf(&existingExec).Elem()
	execValType := execVal.Type()

	for k, v := range updates {
		for i := 0; i < execVal.NumField(); i++ {
			field := execValType.Field(i)
			json_field := field.Tag.Get("json")

			// Check whether such key exists in fields and set its value to v
			if json_field == k+",omitempty" && execVal.Field(i).CanSet() {
				execVal.Field(i).Set(reflect.ValueOf(v).Convert(execVal.Field(i).Type()))
			}
		}
	}

	_, err = db.Exec("UPDATE execs SET first_name=$1, last_name=$2, email=$3, username=$4 WHERE id=$5", existingExec.FirstName, existingExec.LastName, existingExec.Email, existingExec.Username, existingExec.ID)
	if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, fmt.Sprintf("Error updating Exec %d.", execId))
	}
	return existingExec, nil
}

func PatchExecsDBHandler(updates []map[string]interface{}) ([]models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting DB.")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error updating Execs.")
	}

	var existingExecs []models.Exec
	for _, update := range updates {
		execIdStr, ok := update["id"].(string)
		if !ok {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "Error updating Execs.")
		}

		execId, err := strconv.Atoi(execIdStr)
		if err != nil {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "Error updating Execs.")
		}

		var existingExec models.Exec
		err = tx.QueryRow("SELECT id, first_name, last_name, email, username FROM execs WHERE id = $1", execId).Scan(&existingExec.ID, &existingExec.FirstName, &existingExec.LastName, &existingExec.Email, &existingExec.Username)
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "Error updating Execs.")
		} else if err != nil {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "Error updating Execs.")
		}

		// apply updates using reflect
		execVal := reflect.ValueOf(&existingExec).Elem()
		execValType := execVal.Type()

		for k, v := range update {
			if k == "id" {
				continue // Skip the id field.
			}
			for i := 0; i < execVal.NumField(); i++ {
				field := execValType.Field(i)
				json_field := field.Tag.Get("json")

				// Check whether such key exists in fields and set its value to v
				if json_field == k+",omitempty" && execVal.Field(i).CanSet() {
					if reflect.ValueOf(v).Type().ConvertibleTo(execVal.Field(i).Type()) {
						execVal.Field(i).Set(reflect.ValueOf(v).Convert(execVal.Field(i).Type()))
					} else {
						tx.Rollback()
						return nil, utils.ErrorHandler(err, "Error updating Execs.")
					}
					break
				}
			}
		}

		_, err = tx.Exec("UPDATE execs SET first_name=$1, last_name=$2, email=$3, username=$4 WHERE id=$5", existingExec.FirstName, existingExec.LastName, existingExec.Email, existingExec.Username, existingExec.ID)
		if err != nil {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "Error updating Execs.")
		}
		existingExecs = append(existingExecs, existingExec)
	}

	err = tx.Commit()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error updating Execs.")
	}
	return existingExecs, nil
}

func DeleteOneExecDBHandler(execId int) error {
	db, err := ConnectDB()
	if err != nil {
		// log.Println("Error connecting DB :", err)
		return utils.ErrorHandler(err, "Error connecting DB.")
	}
	defer db.Close()

	// Perform the delete operation
	res, err := db.Exec("DELETE FROM execs WHERE id=$1", execId)
	if err != nil {
		return utils.ErrorHandler(err, fmt.Sprintf("Error deleting Exec %d.", execId))
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return utils.ErrorHandler(err, fmt.Sprintf("Error deleting Exec %d.", execId))
	}

	// Operation was successful, but no rows affected i.e. invalid ID.
	if rowsAffected == 0 {
		return utils.ErrorHandler(err, fmt.Sprintf("Error deleting Exec %d.", execId))
	}
	return nil
}

func LoginExecDBHandler(req models.Exec) (models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Internal Server Error.")
	}

	defer db.Close()

	exec := models.Exec{}
	err = db.QueryRow("SELECT id, first_name, last_name, email, username, password, inactive_status, role from execs WHERE username=$1", req.Username).Scan(&exec.ID, &exec.FirstName, &exec.LastName, &exec.Email, &exec.Username, &exec.Password, &exec.InactiveStatus, &exec.Role)
	if err == sql.ErrNoRows {
		return models.Exec{}, utils.ErrorHandler(err, "Incorrect Username / Password.")
	}
	if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Internal Server Error.")
	}

	if exec.InactiveStatus {
		return models.Exec{}, utils.ErrorHandler(errors.New("account is inactive"), "Account is inactive.")
	}
	return exec, nil
}

func UpdateExecPasswordDBHandler(execId int, req models.UpdatePasswordRequest) (string, string, error) {
	db, err := ConnectDB()
	if err != nil {
		return "", "", utils.ErrorHandler(err, "Internal Server Error.")
	}

	defer db.Close()
	var execName string
	var execPwd string
	var execRole string

	err = db.QueryRow("SELECT username, password, role FROM execs WHERE id=$1", execId).Scan(&execName, &execPwd, &execRole)
	if err != nil {
		return "", "", utils.ErrorHandler(err, "User Not Found.")
	}
	err = utils.VerifyPassword(execPwd, req.CurrentPassword)
	if err != nil {
		return "", "", err
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return "", "", err
	}
	_, err = db.Exec("UPDATE execs SET password=$1, password_changed_at=$2 WHERE id=$3", hashedPassword, time.Now(), execId)
	if err != nil {
		return "", "", utils.ErrorHandler(err, "Failed to Update Password.")
	}
	return execName, execRole, nil
}

func ForgotExecPasswordDBHandler(execEmail string) (time.Duration, string, error) {
	db, err := ConnectDB()
	if err != nil {
		return 0, "", utils.ErrorHandler(err, "Internal Server Error.")
	}
	defer db.Close()

	var exec models.Exec
	err = db.QueryRow("SELECT id FROM execs WHERE email=$1", execEmail).Scan(&exec.ID)
	if err != nil {
		return 0, "", utils.ErrorHandler(err, "User Not Found.")
	}

	duration, err := strconv.Atoi(os.Getenv("RESET_TOKEN_EXP_DURATION"))
	if err != nil {
		return 0, "", utils.ErrorHandler(err, "Some error occured.")
	}
	mins := time.Duration(duration)
	expiry := time.Now().Add(mins * time.Minute)
	tokenBytes := make([]byte, 32)
	_, err = rand.Read(tokenBytes)
	if err != nil {
		return 0, "", utils.ErrorHandler(err, "Failed to send Password reset email.")
	}

	token := hex.EncodeToString(tokenBytes)
	hashedToken := sha256.Sum256(tokenBytes)
	hashedTokenString := hex.EncodeToString(hashedToken[:])
	_, err = db.Exec("UPDATE execs SET password_reset_token=$1, password_token_expires=$2 WHERE id=$3", hashedTokenString, expiry, exec.ID)
	if err != nil {
		return 0, "", utils.ErrorHandler(err, "Failed to send Password reset email.")
	}
	return mins, token, nil
}