package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

func VerifyPassword(execPassword string, reqPassword string) error {
	parts := strings.Split(execPassword, ".")
	if len(parts) != 2 {
		return ErrorHandler(errors.New("invalid encoded hash format"), "Internal Server error.")
	}
	hashedSaltBase64 := parts[0]
	hashedPwdBase64 := parts[1]

	salt, err := base64.StdEncoding.DecodeString(hashedSaltBase64)
	if err != nil {
		return ErrorHandler(err, "Internal Server error.")
	}
	hashedPwd, err := base64.StdEncoding.DecodeString(hashedPwdBase64)
	if err != nil {
		return ErrorHandler(err, "Internal Server error.")
	}

	hash := argon2.IDKey([]byte(reqPassword), salt, 1, 64*1024, 4, 32)
	if len(hash) != len(hashedPwd) {
		return ErrorHandler(err, "Incorrect Username / Password.")
	}
	if subtle.ConstantTimeCompare(hash, hashedPwd) != 1 {
		return ErrorHandler(err, "Incorrect Username / Password.")
	}
	return nil
}

func HashPassword(newExecPassword string) (string, error) {
	if newExecPassword == "" {
		return "", ErrorHandler(errors.New("password is blank"), "Password cannot be Empty")
	}
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", ErrorHandler(errors.New("failed to generate salt"), "Error adding Execs.")
	}

	// Hash the Password
	hash := argon2.IDKey([]byte(newExecPassword), salt, 1, 64*1024, 4, 32)
	saltBase64 := base64.StdEncoding.EncodeToString(salt)
	hashBase64 := base64.StdEncoding.EncodeToString(hash)
	encodedHash := fmt.Sprintf("%s.%s", saltBase64, hashBase64)
	newExecPassword = encodedHash
	return encodedHash, nil
}