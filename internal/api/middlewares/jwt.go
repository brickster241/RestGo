package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/brickster241/rest-go/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
)

type ContextKey string

func JWT_MW(next http.Handler) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r* http.Request)  {
		// Fetch the cookies and check
		token, err := r.Cookie("Bearer")
		if err != nil {
			// Cookie not found.
			http.Error(w, "Authorization Header Missing.", http.StatusForbidden)
			return
		}

		jwtSecret := os.Getenv("JWT_SECRET")
		parsedToken, err := jwt.Parse(token.Value, func (token *jwt.Token) (interface {}, error) {
			
			// Don't forget to validate the algo is what you expect.
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				utils.ErrorHandler(fmt.Errorf("unexpected signing method : %v", token.Header["alg"]), "Unauthorized.") 
				return nil, err
			}
			return []byte(jwtSecret), nil
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
			 
		}
		if !parsedToken.Valid {
			http.Error(w, "Invalid Login Token", http.StatusUnauthorized)
			return
		}
		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid Login Token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ContextKey("role"), claims["role"])
		ctx = context.WithValue(ctx, ContextKey("expiresAt"), claims["exp"])
		ctx = context.WithValue(ctx, ContextKey("username"), claims["user"])
		ctx = context.WithValue(ctx, ContextKey("userId"), claims["uid"])

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}