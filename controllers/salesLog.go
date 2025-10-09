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

	utils.SendResponse(ctx, http.StatusOK, "success", "Fetched sales-logs successfully!", salesLogs)
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

// AddSalesLog → POST /salesLogs
func (sc *SalesController) AddSalesLog(ctx iris.Context) {
	var salesLog models.SalesLogRequest
	if err := ctx.ReadJSON(&salesLog); err != nil {
		utils.SendResponse(ctx, http.StatusBadRequest, "error", "Invalid request payload", nil)
		return
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := sc.SalesColl.CountDocuments(timeoutCtx, bson.M{"date": salesLog.Date})
	if err != nil {
		utils.SendResponseAndStop(ctx, http.StatusInternalServerError, "error", "database error", nil)
		return
	}
	if count > 0 {
		utils.SendResponseAndStop(ctx, http.StatusConflict, "error", "Date already added!", nil)
		return
	}

	startOfPrevDay := time.Date(salesLog.Date.Year(), salesLog.Date.Month(), salesLog.Date.Day()-1, 0, 0, 0, 0, salesLog.Date.Location())
	endOfPrevDay := startOfPrevDay.Add(24 * time.Hour)

	filter := bson.M{
		"date": bson.M{
			"$gte": startOfPrevDay,
			"$lt":  endOfPrevDay,
		},
	}

	var prevSalesLog models.SalesLog
	err = sc.SalesColl.FindOne(timeoutCtx, filter).Decode(&prevSalesLog)
	if err != nil && err != mongo.ErrNoDocuments {
		utils.SendResponse(ctx, http.StatusInternalServerError, "error", "Failed to fetch previous day log: "+err.Error(), nil)
		return
	}

	physical := prevSalesLog.Physical + salesLog.Inward - salesLog.Outward - salesLog.Sales

	difference := salesLog.System - prevSalesLog.Physical

	insert := models.SalesLog{
		Date:       salesLog.Date,
		Opening:    prevSalesLog.Physical,
		Inward:     salesLog.Inward,
		Sales:      salesLog.Sales,
		Outward:    salesLog.Outward,
		Physical:   physical,
		System:     salesLog.System,
		Difference: difference,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	result, err := sc.SalesColl.InsertOne(timeoutCtx, insert)
	if err != nil {
		utils.SendResponse(ctx, http.StatusInternalServerError, "error", "Failed to insert sales-logs: "+err.Error(), nil)
		return
	}

	utils.SendResponse(ctx, http.StatusOK, "success", "Sales log created successfully!", iris.Map{"inserted_id": result.InsertedID})
}

// UpdateSalesLogByID → PUT /salesLogs/{id}
func (sc *SalesController) UpdateSalesLogByID(ctx iris.Context) {
	id := ctx.Params().Get("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.SendResponse(ctx, http.StatusBadRequest, "error", "Invalid ID format", nil)
		return
	}

	var updateData models.SalesLogUpdateRequest
	if err := ctx.ReadJSON(&updateData); err != nil {
		utils.SendResponse(ctx, http.StatusBadRequest, "error", "Invalid request payload"+err.Error(), nil)
		return
	}

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	startOfPrevDay := time.Date(updateData.Date.Year(), updateData.Date.Month(), updateData.Date.Day()-1, 0, 0, 0, 0, updateData.Date.Location())
	endOfPrevDay := startOfPrevDay.Add(24 * time.Hour)

	filter := bson.M{
		"date": bson.M{
			"$gte": startOfPrevDay,
			"$lt":  endOfPrevDay,
		},
	}

	var prevSalesLog models.SalesLog
	err = sc.SalesColl.FindOne(ctxTimeout, filter).Decode(&prevSalesLog)
	if err != nil && err != mongo.ErrNoDocuments {
		utils.SendResponse(ctx, http.StatusInternalServerError, "error", "Failed to fetch previous day log: "+err.Error(), nil)
		return
	}

	var currentSalesLog models.SalesLog
	err = sc.SalesColl.FindOne(ctxTimeout, bson.M{"_id": objectID}).Decode(&currentSalesLog)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.SendResponse(ctx, http.StatusNotFound, "error", "Sales log not found", nil)
			return
		}
		utils.SendResponse(ctx, http.StatusInternalServerError, "error", "Failed to fetch current log: "+err.Error(), nil)
		return
	}

	physical := prevSalesLog.Physical + updateData.Inward - updateData.Outward - updateData.Sales

	difference := currentSalesLog.System - prevSalesLog.Physical

	update := bson.M{
		"$set": bson.M{
			"inward":     updateData.Inward,
			"sales":      updateData.Sales,
			"outward":    updateData.Outward,
			"physical":   physical,
			"difference": difference,
			"updated_at": time.Now(),
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
