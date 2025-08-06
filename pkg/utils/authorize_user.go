package utils

import "errors"

type ContextKey string

func AuthorizeUser(usrRole string, allowedRoles ...string) (bool, error) {
	for _, allowedRole := range allowedRoles {
		if usrRole == allowedRole {
			return true, nil
		}
	}
	return false, errors.New("user not authorized")
}