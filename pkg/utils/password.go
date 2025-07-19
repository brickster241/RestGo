package utils

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
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