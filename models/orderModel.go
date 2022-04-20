package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
	ID primitive.ObjectID `bson:"_id"`
	OrderDate time.Time `json:"order_date" validate:"required"`
	Created_At time.Time `json:"created_at"`
	Updated_At time.Time `json:"updated_at"`
	OrderID string `json:"order_id"`
	TableID *string `json:"table_id" validate:"required"`
}