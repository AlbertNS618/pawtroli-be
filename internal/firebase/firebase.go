package firebase

import (
	"cloud.google.com/go/firestore"
	"context"

	"pawtroli-be/internal/logger"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

var App *firebase.App

func InitFirebase() {
	logger.LogInfo("Initializing Firebase...")
	opt := option.WithCredentialsFile("C:/Users/Asus/Documents/Skripsi/Pawtroli/pawtroli-be/configs/firebase_config.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		logger.LogErrorf("Error initializing firebase: %v", err)
		panic(err)
	}
	App = app
	logger.LogInfo("Firebase initialized successfully.")
}

var firestoreClient *firestore.Client

func InitHandlers() {
	ctx := context.Background()
	client, err := App.Firestore(ctx)
	if err != nil {
		logger.LogErrorf("❌ Failed to init Firestore: %v", err)
		panic(err)
	}
	firestoreClient = client
	logger.LogInfo("✅ Firestore client initialized")
}
