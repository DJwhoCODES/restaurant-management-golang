package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Note struct {
	ID         primitive.ObjectID `bson:"_id"`
	Title      *string            `json:"title" validate:"required,min=2,max=100"`
	Text       *string            `json:"text" validate:"required,min=2,max=1000"`
	Created_At time.Time          `json:"created_at"`
	Updated_At time.Time          `json:"updated_at"`
	Note_Id    string             `json:"note_id"`
}
