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

func ChatRoutes(r *mux.Router) {
	chats := r.PathPrefix("/chats").Subrouter()
	chats.HandleFunc("", CreateChatRoom).Methods("POST")
	chats.HandleFunc("/{roomId}/messages", SendMessage).Methods("POST")
	chats.HandleFunc("/{roomId}/messages", GetMessages).Methods("GET")
}

// POST /chats/{roomId}
func CreateChatRoom(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	roomId := mux.Vars(r)["roomId"]
	logger.LogInfof("CreateChatRoom called for roomId: %s", roomId)

	room := new(models.ChatRoom)
	if err := json.NewDecoder(r.Body).Decode(room); err != nil {
		logger.LogErrorf("Failed to decode chat room: %v", err)
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
	room.CreatedAt = time.Now()

	// Check if chat room already exists
	docRef := firestoreClient.Collection("chats").Doc(roomId)
	docSnap, err := docRef.Get(context.Background())
	if err == nil && docSnap.Exists() {
		logger.LogInfof("Chat room already exists: %s", roomId)
		room.ID = roomId
		logger.LogHTTPRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusOK, time.Since(start))
		json.NewEncoder(w).Encode(room)
		return
	}

	_, err = docRef.Set(context.Background(), room)
	duration := time.Since(start)
	if err != nil {
		logger.LogErrorf("Failed to create chat room: %v", err)
		logger.LogFirestoreOperation("CREATE", "chats", roomId, false, duration)
		http.Error(w, "Error creating chat room", http.StatusInternalServerError)
		return
	}
	room.ID = roomId
	logger.LogInfof("Chat room created: %s", roomId)
	logger.LogFirestoreOperation("CREATE", "chats", roomId, true, duration)
	logger.LogHTTPRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusOK, time.Since(start))
	json.NewEncoder(w).Encode(room)
}

// POST /chats/{roomId}/messages
func SendMessage(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	roomId := mux.Vars(r)["roomId"]
	logger.LogInfof("SendMessage called for roomId: %s", roomId)

	msg := new(models.Message)
	if err := json.NewDecoder(r.Body).Decode(msg); err != nil {
		logger.LogErrorf("Failed to decode message: %v", err)
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
	msg.Timestamp = time.Now()
	msg.RoomID = roomId

	doc, _, err := firestoreClient.Collection("chats").Doc(roomId).Collection("messages").Add(context.Background(), msg)
	duration := time.Since(start)
	if err != nil {
		logger.LogErrorf("Failed to send message: %v", err)
		logger.LogFirestoreOperation("CREATE", "chats/"+roomId+"/messages", "", false, duration)
		http.Error(w, "Error sending message", http.StatusInternalServerError)
		return
	}
	msg.ID = doc.ID
	logger.LogInfof("Message sent with ID: %s in roomId: %s", msg.ID, roomId)
	logger.LogFirestoreOperation("CREATE", "chats/"+roomId+"/messages", msg.ID, true, duration)
	logger.LogHTTPRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusOK, time.Since(start))
	json.NewEncoder(w).Encode(msg)
}

type MessageResponse struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	SenderID  string `json:"senderId"`
	RoomID    string `json:"roomId"`
	Timestamp string `json:"timestamp"`
}

// GET /chats/{roomId}/messages
func GetMessages(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	roomId := mux.Vars(r)["roomId"]
	logger.LogInfof("GetMessages called for roomId: %s", roomId)

	docs, err := firestoreClient.Collection("chats").Doc(roomId).Collection("messages").OrderBy("timestamp", firestore.Asc).Documents(context.Background()).GetAll()
	duration := time.Since(start)
	if err != nil {
		logger.LogErrorf("Failed to fetch messages: %v", err)
		logger.LogFirestoreOperation("READ", "chats/"+roomId+"/messages", "", false, duration)
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}

	loc, _ := time.LoadLocation("Asia/Jakarta") // UTC+7

	var messages []MessageResponse
	for _, doc := range docs {
		m := new(models.Message)
		doc.DataTo(m)
		m.ID = doc.Ref.ID
		// Convert timestamp to UTC+7 before formatting
		messages = append(messages, MessageResponse{
			ID:        m.ID,
			Content:   m.Content,
			SenderID:  m.SenderID,
			RoomID:    m.RoomID,
			Timestamp: m.Timestamp.In(loc).Format(time.RFC3339),
		})
		logger.LogDebugf("Message: %+v", m)
	}
	logger.LogInfof("Fetched %d messages for roomId: %s", len(messages), roomId)
	logger.LogFirestoreOperation("READ", "chats/"+roomId+"/messages", "", true, duration)
	logger.LogHTTPRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
