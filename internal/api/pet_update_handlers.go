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
	"google.golang.org/api/iterator"
)

// POST /pets/{petId}/updates
func CreatePetUpdate(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	petId := mux.Vars(r)["petId"]
	logger.LogInfof("CreatePetUpdate called for petId: %s", petId)

	update := new(models.PetUpdate)
	update.PetID = petId // Set the petId from the URL
	if err := json.NewDecoder(r.Body).Decode(update); err != nil {
		logger.LogErrorf("Failed to decode pet update: %v", err)
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
	update.Timestamp = time.Now()
	petStatus := update.Caption

	// 1) Add the pet update
	_, _, err := firestoreClient.Collection("pet_updates").Add(context.Background(), update)
	duration := time.Since(start)
	if err != nil {
		logger.LogErrorf("Failed to add pet update: %v", err)
		logger.LogFirestoreOperation("CREATE", "pet_updates", "", false, duration)
		http.Error(w, "Failed to add update", http.StatusInternalServerError)
		return
	}

	// 2) ALSO update the pet's status field in the pets collection
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	_, err = firestoreClient.Collection("pets").Doc(petId).Update(ctx2, []firestore.Update{
		{Path: "status", Value: petStatus},
	})
	if err != nil {
		logger.LogErrorf("Failed to update pet status: %v", err)
		// we don't abort the request, we just log it
	}

	logger.LogInfof("Pet update added and status set to %q for petId: %s", petStatus, petId)
	logger.LogFirestoreOperation("CREATE", "pet_updates", "", true, duration)
	logger.LogHTTPRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusCreated, time.Since(start))
	w.WriteHeader(http.StatusCreated)
}

// GET /pets/{petId}/updates
func GetPetUpdates(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	petId := mux.Vars(r)["petId"]
	logger.LogInfof("GetPetUpdates called for petId: %s", petId)

	ctx := context.Background()
	iter := firestoreClient.Collection("pet_updates").Where("petId", "==", petId).Documents(ctx)
	defer iter.Stop()

	var updates []models.PetUpdate
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			logger.LogErrorf("Error fetching pet updates: %v", err)
			http.Error(w, "Failed to fetch updates", http.StatusInternalServerError)
			return
		}
		var update models.PetUpdate
		if err := doc.DataTo(&update); err != nil {
			logger.LogErrorf("Error decoding pet update: %v", err)
			continue
		}
		update.ID = doc.Ref.ID
		updates = append(updates, update)
		logger.LogInfof("Fetched update: %+v", update)
	}
	logger.LogInfof("Fetched %d updates for petId: %s", len(updates), petId)
	logger.LogHTTPRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updates)
}
