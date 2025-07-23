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

func PetRoutes(r *mux.Router) {
	pets := r.PathPrefix("/pets").Subrouter()
	pets.HandleFunc("/{id}", GetPet).Methods("GET")
	pets.HandleFunc("", CreatePet).Methods("POST")
	pets.HandleFunc("/{petId}/activate", ActivatePet).Methods("PATCH")
	pets.HandleFunc("/{petId}/updates", CreatePetUpdate).Methods("POST")
	pets.HandleFunc("/{petId}/updates", GetPetUpdates).Methods("GET")
	pets.HandleFunc("/{petId}/delete", DeletePet).Methods("DELETE")
}

// POST /pets
func CreatePet(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logger.LogInfo("CreatePet called")

	pet := new(models.Pet)

	if err := json.NewDecoder(r.Body).Decode(pet); err != nil {
		logger.LogErrorf("Failed to decode pet: %v", err)
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
	logger.LogInfof("Creating pet: %+v", pet)
	pet.CreatedAt = time.Now()

	// Use provided PetID
	docRef := firestoreClient.Collection("pets").Doc(pet.PetID)

	// Write the document
	_, err := docRef.Set(context.Background(), pet)
	duration := time.Since(start)
	if err != nil {
		logger.LogErrorf("Failed to save pet: %v", err)
		logger.LogFirestoreOperation("CREATE", "pets", pet.PetID, false, duration)
		http.Error(w, "Error saving pet", http.StatusInternalServerError)
		return
	}

	logger.LogInfof("Pet saved with ID: %s", pet.PetID)
	logger.LogFirestoreOperation("CREATE", "pets", pet.PetID, true, duration)
	logger.LogHTTPRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusOK, time.Since(start))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pet)
}

// GET /petId
func GetPet(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	petID := mux.Vars(r)["id"]
	logger.LogInfof("Fetching pet with ID: %s", petID)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Query Firestore for the pet document
	doc, err := firestoreClient.Collection("pets").Doc(petID).Get(ctx)
	if err != nil {
		logger.LogErrorf("Error fetching pet: %v", err)
		http.Error(w, "Failed to fetch pet", http.StatusInternalServerError)
		return
	}

	// Check if document exists
	if !doc.Exists() {
		logger.LogErrorf("Pet not found: %s", petID)
		http.Error(w, "Pet not found", http.StatusNotFound)
		return
	}

	// Map the document data to our Pet model
	var pet models.Pet
	if err := doc.DataTo(&pet); err != nil {
		logger.LogErrorf("Error mapping data to pet model: %v", err)
		http.Error(w, "Error processing pet data", http.StatusInternalServerError)
		return
	}

	// Add the ID to the model
	pet.PetID = petID

	// Log success
	logger.LogInfof("Successfully fetched pet %s in %v", petID, time.Since(startTime))

	// Return the pet as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(pet); err != nil {
		logger.LogErrorf("Error encoding pet to JSON: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func ActivatePet(w http.ResponseWriter, r *http.Request) {
	petId := mux.Vars(r)["petId"]

	// 1) Decode JSON payload
	var req struct {
		CheckIn  string `json:"checkIn"`
		CheckOut string `json:"checkOut"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.LogErrorf("Failed to decode ActivatePet body: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	logger.LogInfof("ActivatePet called for petId=%s, payload=%+v", petId, req)

	// 2) Parse ISO-8601 timestamps
	checkInTime, err := time.Parse(time.RFC3339, req.CheckIn)
	if err != nil {
		logger.LogErrorf("Invalid checkIn format: %v", err)
		http.Error(w, "Invalid checkIn timestamp", http.StatusBadRequest)
		return
	}
	checkOutTime, err := time.Parse(time.RFC3339, req.CheckOut)
	if err != nil {
		logger.LogErrorf("Invalid checkOut format: %v", err)
		http.Error(w, "Invalid checkOut timestamp", http.StatusBadRequest)
		return
	}

	// 3) Perform the Firestore update
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = firestoreClient.Collection("pets").Doc(petId).Update(ctx, []firestore.Update{
		{Path: "active", Value: true},
		{Path: "checkIn", Value: checkInTime},
		{Path: "checkOut", Value: checkOutTime},
	})
	if err != nil {
		logger.LogErrorf("Failed to activate pet: %v", err)
		http.Error(w, "Failed to activate pet", http.StatusInternalServerError)
		return
	}

	logger.LogInfof("Successfully activated pet: %s", petId)
	w.WriteHeader(http.StatusNoContent)
}

// DELETE /pets/{petId}/delete
func DeletePet(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	petId := mux.Vars(r)["petId"]
	logger.LogInfof("DeletePet called for petId: %s", petId)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := firestoreClient.Collection("pets").Doc(petId).Delete(ctx)
	duration := time.Since(start)
	if err != nil {
		logger.LogErrorf("Failed to delete pet: %v", err)
		logger.LogFirestoreOperation("DELETE", "pets", petId, false, duration)
		http.Error(w, "Failed to delete pet", http.StatusInternalServerError)
		return
	}

	logger.LogInfof("Successfully deleted pet: %s", petId)
	logger.LogFirestoreOperation("DELETE", "pets", petId, true, duration)
	logger.LogHTTPRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusNoContent, duration)

	w.WriteHeader(http.StatusNoContent)
}
