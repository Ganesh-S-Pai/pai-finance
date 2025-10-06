package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/Ganesh-S-Pai/pai-finance/models"
	"github.com/Ganesh-S-Pai/pai-finance/utils"

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
		utils.SendResponse(ctx, http.StatusBadRequest, "error", "Invalid request payload", nil)
		return
	}

	// Use per-request context with timeout
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := sc.SalesColl.InsertOne(timeoutCtx, salesLog)
	if err != nil {
		utils.SendResponse(ctx, http.StatusInternalServerError, "error", "Failed to insert sales-logs: "+err.Error(), nil)
		return
	}

	utils.SendResponse(ctx, http.StatusOK, "success", "Sales log created successfully!", iris.Map{"inserted_id": result.InsertedID})
}

// GetSalesLogs → GET /salesLogs
func (sc *SalesController) GetSalesLogs(ctx iris.Context) {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := sc.SalesColl.Find(timeoutCtx, bson.M{})
	if err != nil {
		utils.SendResponse(ctx, http.StatusInternalServerError, "error", "Failed to fetch sales-logs: "+err.Error(), nil)
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

	utils.SendResponse(ctx, http.StatusOK, "success", "Fetched sales-logs successfully!: ", salesLogs)
}

// GetSalesLogByID → GET /salesLogs/{id}
func (sc *SalesController) GetSalesLogByID(ctx iris.Context) {
	id := ctx.Params().Get("id")

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ✅ Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.SendResponse(ctx, http.StatusBadRequest, "error", "Invalid ID format", nil)
		return
	}

	var salesLog models.SalesLog
	err = sc.SalesColl.FindOne(timeoutCtx, bson.M{"_id": objectID}).Decode(&salesLog)
	if err == mongo.ErrNoDocuments {
		utils.SendResponse(ctx, http.StatusNotFound, "error", "Sales log not found", nil)
		return
	} else if err != nil {
		utils.SendResponse(ctx, http.StatusInternalServerError, "error", "Database error: "+err.Error(), nil)
		return
	}

	utils.SendResponse(ctx, http.StatusOK, "success", "Fetched sales-log successfully!", salesLog)
}

// UpdateSalesLogByID → PUT /salesLogs/{id}
func (sc *SalesController) UpdateSalesLogByID(ctx iris.Context) {
	id := ctx.Params().Get("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.SendResponse(ctx, http.StatusBadRequest, "error", "Invalid ID format", nil)
		return
	}

	var updateData models.SalesLog
	if err := ctx.ReadJSON(&updateData); err != nil {
		utils.SendResponse(ctx, http.StatusBadRequest, "error", "Invalid request payload"+err.Error(), nil)
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
		utils.SendResponse(ctx, http.StatusInternalServerError, "error", "Database error: "+err.Error(), nil)
		return
	}

	if res.MatchedCount == 0 {
		utils.SendResponse(ctx, http.StatusNotFound, "error", "Sales log not found", nil)
		return
	}

	utils.SendResponse(ctx, http.StatusOK, "success", "Updated successfully", iris.Map{"id": objectID})
}

// DeleteSalesLogByID → DELETE /salesLogs/{id}
func (sc *SalesController) DeleteSalesLogByID(ctx iris.Context) {
	id := ctx.Params().Get("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.SendResponse(ctx, http.StatusBadRequest, "error", "Invalid ID format", nil)
		return
	}

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := sc.SalesColl.DeleteOne(ctxTimeout, bson.M{"_id": objectID})
	if err != nil {
		utils.SendResponse(ctx, http.StatusInternalServerError, "error", "Database error: "+err.Error(), nil)
		return
	}

	if res.DeletedCount == 0 {
		utils.SendResponse(ctx, http.StatusNotFound, "error", "Sales log not found", nil)
		return
	}

	utils.SendResponse(ctx, http.StatusOK, "success", "Deleted successfully", iris.Map{"id": objectID})
}
