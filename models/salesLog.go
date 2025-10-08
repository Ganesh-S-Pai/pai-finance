package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SalesLog struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Date       time.Time          `bson:"date" json:"date"`
	Opening    int                `bson:"opening" json:"opening"`
	Inward     int                `bson:"inward" json:"inward"`
	Sales      int                `bson:"sales" json:"sales"`
	Outward    int                `bson:"outward" json:"outward"`
	Physical   int                `bson:"physical" json:"physical"`
	System     int                `bson:"system" json:"system"`
	Difference int                `bson:"difference" json:"difference"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

type SalesLogRequest struct {
	Date    time.Time `bson:"date" json:"date"`
	Inward  int       `bson:"inward" json:"inward"`
	Sales   int       `bson:"sales" json:"sales"`
	Outward int       `bson:"outward" json:"outward"`
	System  int       `bson:"system" json:"system"`
}

type SalesLogUpdateRequest struct {
	Date       time.Time `bson:"date" json:"date"`
	Inward     int       `bson:"inward" json:"inward"`
	Sales      int       `bson:"sales" json:"sales"`
	Outward    int       `bson:"outward" json:"outward"`
	System     int       `bson:"system" json:"system"`
	Difference int       `bson:"difference" json:"difference"`
	UpdatedAt  time.Time `bson:"updated_at" json:"updated_at"`
}
