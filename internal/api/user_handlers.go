package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"pawtroli-be/internal/logger"
	"pawtroli-be/internal/models"

	"cloud.google.com/go/firestore"
	"github.com/gorilla/mux"
)

func UserRoutes(r *mux.Router) {
	r.HandleFunc("/register", UserRegister).Methods("POST")
	r.HandleFunc("/login", UserLogin).Methods("POST")
}

// GET /register
func UserRegister(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logger.LogInfo("UserRegister called")

	user := new(models.User)
	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		logger.LogErrorf("Failed to decode user: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logger.LogInfof("Registering user: %+v", user)

	ctx := context.Background()
	docRef := firestoreClient.Collection("users").Doc(user.ID)

	_, err := docRef.Set(ctx, map[string]interface{}{
		"name":      user.Name,
		"email":     user.Email,
		"phone":     user.Phone,
		"role":      user.Role, // "user" or "admin"
		"createdAt": time.Now(),
	}, firestore.MergeAll)

	duration := time.Since(start)
	if err != nil {
		logger.LogErrorf("Failed to save user: %v", err)
		logger.LogFirestoreOperation("CREATE", "users", user.ID, false, duration)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.LogInfof("User registered: %s", user.ID)
	logger.LogFirestoreOperation("CREATE", "users", user.ID, true, duration)
	logger.LogHTTPRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusOK, duration)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// UserLogin handles authenticated requests to /login
func UserLogin(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logger.LogInfof("UserLogin called: method=%s, url=%s, remoteAddr=%s",
		r.Method, r.URL.Path, r.RemoteAddr)

	uid, ok := r.Context().Value("uid").(string)
	logger.LogInfof("User ID from context: %s", uid)

	if !ok || uid == "" {
		logger.LogWarning("Unauthorized access attempt to /login")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Fetch user data from Firestore
	ctx := context.Background()
	doc, err := firestoreClient.Collection("users").Doc(uid).Get(ctx)
	duration := time.Since(start)

	if err != nil {
		logger.LogErrorf("Failed to get user data: %v", err)
		logger.LogFirestoreOperation("READ", "users", uid, false, duration)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	data := doc.Data()
	name, _ := data["name"].(string)
	email, _ := data["email"].(string)
	phone, _ := data["phone"].(string)
	role, _ := data["role"].(string)

	logger.LogInfof("Authenticated user: %s", uid)
	logger.LogFirestoreOperation("READ", "users", uid, true, duration)
	logger.LogHTTPRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusOK, time.Since(start))
	logger.LogAuthOperation("login", uid, true)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "authenticated",
		"uid":    uid,
		"name":   name,
		"email":  email,
		"phone":  phone,
		"role":   role,
	})
}
