package models

import "time"

type User struct {
	ID        string    `firestore:"-"` // use for document ID
	Name      string    `firestore:"name"`
	Email     string    `firestore:"email"`
	Phone     string    `firestore:"phone"`
	Role      string    `firestore:"role"` // "user" or "admin"
	CreatedAt time.Time `firestore:"createdAt"`
}

type Pet struct {
	PetID     string    `json:"petId" firestore:"petId"` // <-- store PetID in a petId field
	Name      string    `json:"name" firestore:"name"`
	Type      string    `json:"type" firestore:"type"`
	Gender    string    `json:"gender" firestore:"gender"`
	Age       int       `json:"age" firestore:"age"`
	Color     string    `json:"color" firestore:"color"`
	Allergy   string    `json:"allergy" firestore:"allergy"`
	Other     string    `json:"other" firestore:"other"`
	OwnerID   string    `json:"ownerId" firestore:"ownerId"`
	ImageURL  string    `json:"imageUrl" firestore:"imageUrl"`
	Active    bool      `json:"active" firestore:"active"`
	Status    string    `json:"status" firestore:"status"`
	CheckIn   time.Time `json:"checkIn" firestore:"checkIn"`
	CheckOut  time.Time `json:"checkOut" firestore:"checkOut"`
	CreatedAt time.Time `json:"createdAt" firestore:"createdAt"`
}

type PetUpdate struct {
	ID          string    `json:"id" firestore:"-"`
	Caption     string    `json:"caption" firestore:"caption"`
	Description string    `json:"description" firestore:"description"`
	ImageURL    string    `json:"imageUrl" firestore:"imageUrl"`
	PetID       string    `json:"petId" firestore:"petId"` // already stored
	Timestamp   time.Time `json:"timestamp" firestore:"timestamp"`
}

type ChatRoom struct {
	ID        string    `firestore:"-"`       // use for document ID
	UserIDs   []string  `firestore:"userIds"` // participants' user IDs
	CreatedAt time.Time `firestore:"createdAt"`
}

type Message struct {
	ID        string    `firestore:"-"` // use for document ID
	RoomID    string    `firestore:"roomId"`
	SenderID  string    `firestore:"senderId"`
	Content   string    `firestore:"content"`
	Timestamp time.Time `firestore:"timestamp"`
}
