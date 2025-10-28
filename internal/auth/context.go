package auth

import (
	"context"
	"errors"
	"net/http"
	"strconv"
)

const userIDKey contextKey = "userID"

type contextKey string

// ContextWithUserID injects the authenticated user ID into the context.
func ContextWithUserID(ctx context.Context, id int) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

// UserIDFromContext extracts the user ID from the context if present.
func UserIDFromContext(ctx context.Context) (int, error) {
	value := ctx.Value(userIDKey)
	if value == nil {
		return 0, errors.New("user id not found in context")
	}
	id, ok := value.(int)
	if !ok {
		return 0, errors.New("user id has unexpected type")
	}
	return id, nil
}

// UserIDFromRequest extracts the user ID from the X-User-ID header.
func UserIDFromRequest(r *http.Request) (int, error) {
	raw := r.Header.Get("X-User-ID")
	if raw == "" {
		return 0, errors.New("missing X-User-ID header")
	}
	id, err := strconv.Atoi(raw)
	if err != nil {
		return 0, errors.New("invalid X-User-ID header")
	}
	return id, nil
}
