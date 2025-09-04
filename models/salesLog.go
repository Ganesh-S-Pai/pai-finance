package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SalesLog struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Date    time.Time          `bson:"date" json:"date"`
	Opening int                `bson:"opening" json:"opening"`
	Inward  int                `bson:"inward" json:"inward"`
	Sales   int                `bson:"sales" json:"sales"`
	Outward int                `bson:"outward" json:"outward"`
 	Created time.Time          `bson:"created" json:"created"`
}