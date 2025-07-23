package api

import (
	"context"
	"cloud.google.com/go/firestore"

	"pawtroli-be/internal/firebase"
	"pawtroli-be/internal/logger"
)

var firestoreClient *firestore.Client

func InitHandlers() {
	ctx := context.Background()
	client, err := firebase.App.Firestore(ctx)
	if err != nil {
		logger.LogErrorf("❌ Failed to init Firestore: %v", err)
		panic(err)
	}
	firestoreClient = client
	logger.LogInfo("✅ Firestore client initialized")
}