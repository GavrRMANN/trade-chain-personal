package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"
)

type contextKey string

const UserIDKey contextKey = "user_id"
const claimsKey contextKey = "claims"

var ErrInvalidAuthHeader = errors.New("invalid authorization header")

func getUserIDFromRequest(r *http.Request) (string, error) {
	claims, err := getClaimsFromRequest(r)
	if err != nil || claims == nil {
		return "", err
	}
	return claims.UserID, nil
}

func getClaimsFromRequest(r *http.Request) (*Claims, error) {
	authHeader := r.Header.Get("Authorization")

	if authHeader == "" {
		return nil, nil
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, ErrInvalidAuthHeader
	}

	claims, err := ValidateToken(parts[1])
	if err != nil {
		return nil, err
	}

	return claims, nil
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := getClaimsFromRequest(r)

		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		if claims == nil || claims.UserID == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(
			r.Context(),
			UserIDKey,
			claims.UserID,
		)
		ctx = context.WithValue(ctx, claimsKey, claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := getUserIDFromRequest(r)

		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// Нет токена — гость.
		if userID == "" {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(
			r.Context(),
			UserIDKey,
			userID,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// TokenExpirationFromContext возвращает время окончания JWT авторизованного запроса.
func TokenExpirationFromContext(ctx context.Context) (time.Time, bool) {
	claims, ok := ctx.Value(claimsKey).(*Claims)
	if !ok || claims.ExpiresAt == nil {
		return time.Time{}, false
	}
	return claims.ExpiresAt.Time, true
}
