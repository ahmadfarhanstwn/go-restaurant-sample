package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Note struct {
	ID primitive.ObjectID `bson:"_id"`
	Text string `json:"text"`
	Title string `json:"title"`
	Created_At time.Time `json:"created_at"`
	Updated_At time.Time `json:"updated_at"`
	Note_ID string `json:"note_id"`
}