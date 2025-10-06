package controllers

import (
	"context"
	"net/http"
	"time"
	"vhiw-sales-log/models"

	"github.com/kataras/iris/v12"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SalesController struct {
	SalesColl *mongo.Collection
}

func NewSalesController(salesColl *mongo.Collection) *SalesController {
	return &SalesController{
		SalesColl: salesColl,
	}
}

// AddSalesLog → POST /salesLogs
func (sc *SalesController) AddSalesLog(ctx iris.Context) {
	var salesLog models.SalesLog
	if err := ctx.ReadJSON(&salesLog); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	// Use per-request context with timeout
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := sc.SalesColl.InsertOne(timeoutCtx, salesLog)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to insert sales-logs: " + err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(iris.Map{"inserted_id": result.InsertedID})
}

// GetSalesLogs → GET /salesLogs
func (sc *SalesController) GetSalesLogs(ctx iris.Context) {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := sc.SalesColl.Find(timeoutCtx, bson.M{})
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to fetch sales-logs: " + err.Error()})
		return
	}
	defer cursor.Close(timeoutCtx)

	var salesLogs []models.SalesLog
	for cursor.Next(timeoutCtx) {
		var salesLog models.SalesLog
		if err := cursor.Decode(&salesLog); err == nil {
			salesLogs = append(salesLogs, salesLog)
		}
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(salesLogs)
}

// GetSalesLogByID → GET /salesLogs/{id}
func (sc *SalesController) GetSalesLogByID(ctx iris.Context) {
	id := ctx.Params().Get("id")

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ✅ Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid ID format"})
		return
	}

	var salesLog models.SalesLog
	err = sc.SalesColl.FindOne(timeoutCtx, bson.M{"_id": objectID}).Decode(&salesLog)
	if err == mongo.ErrNoDocuments {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Sales log not found"})
		return
	} else if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Database error: " + err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(salesLog)
}

// UpdateSalesLogByID → PUT /salesLogs/{id}
func (sc *SalesController) UpdateSalesLogByID(ctx iris.Context) {
	id := ctx.Params().Get("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid ID format"})
		return
	}

	var updateData models.SalesLog
	if err := ctx.ReadJSON(&updateData); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"date":    updateData.Date,
			"opening": updateData.Opening,
			"inward":  updateData.Inward,
			"sales":   updateData.Sales,
			"outward": updateData.Outward,
		},
	}

	res, err := sc.SalesColl.UpdateByID(ctxTimeout, objectID, update)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Database error: " + err.Error()})
		return
	}

	if res.MatchedCount == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Sales log not found"})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(iris.Map{"message": "Updated successfully"})
}

// DeleteSalesLogByID → DELETE /salesLogs/{id}
func (sc *SalesController) DeleteSalesLogByID(ctx iris.Context) {
	id := ctx.Params().Get("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid ID format"})
		return
	}

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := sc.SalesColl.DeleteOne(ctxTimeout, bson.M{"_id": objectID})
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Database error: " + err.Error()})
		return
	}

	if res.DeletedCount == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Sales log not found"})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(iris.Map{"message": "Deleted successfully"})
}
