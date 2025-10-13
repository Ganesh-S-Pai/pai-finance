package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DateInfo struct {
	Type  string     `bson:"type" json:"type"`
	Start *time.Time `bson:"start,omitempty" json:"start,omitempty"`
	End   *time.Time `bson:"end" json:"end,omitempty"`
}

type AirbnbStatement struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Transaction string             `bson:"transaction" json:"transaction"`
	Type        string             `bson:"type" json:"type"`
	Date        DateInfo           `bson:"date_info" json:"date_info"`
	Nights      int                `bson:"nights,omitempty" json:"nights"`
	Amount      float64            `bson:"amount" json:"amount"`
	Remark      string             `bson:"remark,omitempty" json:"remark"`
	CreatedUser string             `bson:"created_user" json:"created_user"`
	UpdatedUser string             `bson:"updated_user" json:"updated_user"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// type AirbnbStatementRequest struct {
// 	Transaction string             `bson:"transaction" json:"transaction"`
// 	Type        string             `bson:"type" json:"type"`
// 	Date        DateInfo           `bson:"date_info" json:"date_info"`
// 	Amount      string             `bson:"amount" json:"amount"`
// 	Remark      string             `bson:"remark" json:"remark"`
// 	CreatedUser string             `bson:"created_user" json:"created_user"`
// 	UpdatedUser string             `bson:"updated_name" json:"updated_user"`
// 	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
// 	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
// }
