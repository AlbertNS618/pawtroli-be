package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"pawtroli-be/internal/firebase"
	"pawtroli-be/internal/logger"
)

func VerifyToken(next http.Handler) http.Handler {
	logger.LogInfo("VerifyToken middleware initialized")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logger.LogDebug("VerifyToken middleware called")

		authHeader := r.Header.Get("Authorization")
		logger.LogDebugf("VerifyToken: Authorization header: %s", authHeader)

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			logger.LogWarning("VerifyToken: Missing or invalid Authorization header")
			http.Error(w, "Missing auth token", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		ctx := context.Background()
		client, err := firebase.App.Auth(ctx)
		if err != nil {
			logger.LogErrorf("VerifyToken: Failed to get auth client: %v", err)
			http.Error(w, "Failed to get auth client", http.StatusInternalServerError)
			return
		}

		token, err := client.VerifyIDToken(ctx, tokenStr)
		duration := time.Since(start)
		if err != nil {
			logger.LogErrorf("VerifyToken: Invalid token: %v", err)
			logger.LogAuthOperation("token_verification", "", false)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		logger.LogInfof("VerifyToken: Authenticated UID: %s (verification took %v)", token.UID, duration)
		logger.LogAuthOperation("token_verification", token.UID, true)
		ctx = context.WithValue(r.Context(), "uid", token.UID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
