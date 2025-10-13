package controllers

import (
	"context"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/Ganesh-S-Pai/pai-finance/models"
	"github.com/Ganesh-S-Pai/pai-finance/utils"
	"github.com/kataras/iris/v12"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AirbnbController struct {
	AirbnbColl *mongo.Collection
}

func NewAirbnbController(airbnbColl *mongo.Collection) *AirbnbController {
	return &AirbnbController{
		AirbnbColl: airbnbColl,
	}
}

func (ac *AirbnbController) GetAllStatements(ctx iris.Context) {
	page, _ := ctx.URLParamInt("page")
	limit, _ := ctx.URLParamInt("limit")
	search := ctx.URLParam("search")

	if limit <= 0 {
		limit = 10
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{}
	if search != "" {
		filter = bson.M{
			"$or": []bson.M{
				{"transaction": bson.M{"$regex": search, "$options": "i"}},
				{"type": bson.M{"$regex": search, "$options": "i"}},
				{"remark": bson.M{"$regex": search, "$options": "i"}},
			},
		}
	}

	total, err := ac.AirbnbColl.CountDocuments(timeoutCtx, filter)
	if err != nil {
		utils.SendResponse(ctx, http.StatusInternalServerError, "error", "Failed to count statements: "+err.Error(), nil)
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	if page <= 0 || page > totalPages {
		page = totalPages
	}

	if page < 1 {
		page = 1
	}

	skip := (page - 1) * limit
	skip64 := int64(skip)
	limit64 := int64(limit)

	findOptions := &options.FindOptions{
		Skip:  &skip64,
		Limit: &limit64,
		Sort:  bson.D{{Key: "created_at", Value: 1}},
	}

	cursor, err := ac.AirbnbColl.Find(timeoutCtx, filter, findOptions)
	if err != nil {
		utils.SendResponse(ctx, http.StatusInternalServerError, "error", "Failed to fetch statements: "+err.Error(), nil)
		return
	}
	defer cursor.Close(timeoutCtx)

	var statements []models.AirbnbStatement
	if err := cursor.All(timeoutCtx, &statements); err != nil {
		utils.SendResponse(ctx, http.StatusInternalServerError, "error", "Failed to decode statements: "+err.Error(), nil)
		return
	}

	allCursor, err := ac.AirbnbColl.Find(timeoutCtx, bson.M{})
	if err != nil {
		utils.SendResponse(ctx, http.StatusInternalServerError, "error", "Failed to fetch all statements: "+err.Error(), nil)
		return
	}
	defer allCursor.Close(timeoutCtx)

	var all []models.AirbnbStatement
	if err := allCursor.All(timeoutCtx, &all); err != nil {
		utils.SendResponse(ctx, http.StatusInternalServerError, "error", "Failed to decode all statements: "+err.Error(), nil)
		return
	}

	var airbnbTotal, directTotal float64
	expenseTotals := map[string]float64{
		"rent":          0,
		"electricity":   0,
		"water":         0,
		"internet":      0,
		"house_keeping": 0,
		"miscellaneous": 0,
	}

	for _, s := range all {
		transaction := strings.ToLower(s.Transaction)
		typ := strings.ToLower(s.Type)

		switch transaction {
		case "booking":
			switch typ {
			case "airbnb":
				airbnbTotal += float64(s.Amount)
			case "direct":
				directTotal += float64(s.Amount)
			}
		case "expense":
			switch typ {
			case "rent":
				expenseTotals["rent"] += float64(s.Amount)
			case "electricity":
				expenseTotals["electricity"] += float64(s.Amount)
			case "water":
				expenseTotals["water"] += float64(s.Amount)
			case "internet":
				expenseTotals["internet"] += float64(s.Amount)
			case "house_keeping":
				expenseTotals["house_keeping"] += float64(s.Amount)
			case "miscellaneous":
				expenseTotals["miscellaneous"] += float64(s.Amount)
			default:
				expenseTotals["miscellaneous"] += float64(s.Amount)
			}
		}
	}

	totalBooking := airbnbTotal + directTotal
	totalExpense := 0.0
	for _, v := range expenseTotals {
		totalExpense += v
	}
	netBalance := totalBooking - totalExpense

	summary := iris.Map{
		"bookings": iris.Map{
			"airbnb_total": airbnbTotal,
			"direct_total": directTotal,
			"total":        totalBooking,
		},
		"expenses": iris.Map{
			"rent":          expenseTotals["rent"],
			"electricity":   expenseTotals["electricity"],
			"water":         expenseTotals["water"],
			"internet":      expenseTotals["internet"],
			"house_keeping": expenseTotals["house_keeping"],
			"miscellaneous": expenseTotals["miscellaneous"],
			"total":         math.Round(totalExpense*100) / 100,
		},
		"net_balance": netBalance,
	}

	utils.SendResponse(ctx, http.StatusOK, "success", "Statements fetched successfully!", iris.Map{
		"items":       statements,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
		"summary":     summary,
	})
}

func (ac *AirbnbController) AddStatement(ctx iris.Context) {
	var statement models.AirbnbStatement
	if err := ctx.ReadJSON(&statement); err != nil {
		utils.SendResponse(ctx, http.StatusBadRequest, "error", "Invalid request payload", nil)
		return
	}

	now := time.Now().UTC()
	statement.CreatedAt = now
	statement.UpdatedAt = now

	if strings.ToLower(statement.Transaction) == "booking" &&
		statement.Date.Type == "range" &&
		statement.Date.Start != nil &&
		statement.Date.End != nil {

		duration := statement.Date.End.Sub(*statement.Date.Start)
		nights := int(duration.Hours() / 24)
		if nights < 0 {
			nights = 0
		}
		statement.Nights = nights
	} else {
		statement.Nights = 0
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := ac.AirbnbColl.InsertOne(timeoutCtx, statement)
	if err != nil {
		utils.SendResponse(ctx, http.StatusInternalServerError, "error", "Failed to insert statement: "+err.Error(), nil)
		return
	}

	utils.SendResponse(ctx, http.StatusOK, "success", "Statement created successfully!", iris.Map{
		"id": result.InsertedID,
	})
}

func (ac *AirbnbController) DeleteStatementByID(ctx iris.Context) {
	id := ctx.Params().Get("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.SendResponse(ctx, http.StatusBadRequest, "error", "Invalid ID", nil)
		return
	}

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := ac.AirbnbColl.DeleteOne(ctxTimeout, bson.M{"_id": objectID})
	if err != nil {
		utils.SendResponse(ctx, http.StatusInternalServerError, "error", "Database error: "+err.Error(), nil)
		return
	}

	if res.DeletedCount == 0 {
		utils.SendResponse(ctx, http.StatusNotFound, "error", "Statement not found", nil)
		return
	}

	utils.SendResponse(ctx, http.StatusOK, "success", "Statement deleted successfully!", iris.Map{"id": objectID})
}

func (ac *AirbnbController) UpdateStatementByID(ctx iris.Context) {
	id := ctx.Params().Get("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.SendResponse(ctx, http.StatusBadRequest, "error", "Invalid ID", nil)
		return
	}

	var statement models.AirbnbStatement
	if err := ctx.ReadJSON(&statement); err != nil {
		utils.SendResponse(ctx, http.StatusBadRequest, "error", "Invalid request payload", nil)
		return
	}

	now := time.Now().UTC()
	statement.UpdatedAt = now

	if strings.ToLower(statement.Transaction) == "booking" &&
		statement.Date.Type == "range" &&
		statement.Date.Start != nil &&
		statement.Date.End != nil {

		duration := statement.Date.End.Sub(*statement.Date.Start)
		nights := int(duration.Hours() / 24)
		if nights < 0 {
			nights = 0
		}
		statement.Nights = nights
	} else {
		statement.Nights = 0
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"transaction":  statement.Transaction,
			"type":         statement.Type,
			"date_info":    statement.Date,
			"nights":       statement.Nights,
			"amount":       statement.Amount,
			"remark":       statement.Remark,
			"updated_user": statement.UpdatedUser,
			"updated_at":   statement.UpdatedAt,
		},
	}

	result, err := ac.AirbnbColl.UpdateByID(timeoutCtx, objectID, update)
	if err != nil {
		utils.SendResponse(ctx, http.StatusInternalServerError, "error", "Failed to update statement: "+err.Error(), nil)
		return
	}

	if result.MatchedCount == 0 {
		utils.SendResponse(ctx, http.StatusNotFound, "error", "Statement not found", nil)
		return
	}

	var updatedStatement models.AirbnbStatement
	err = ac.AirbnbColl.FindOne(timeoutCtx, bson.M{"_id": objectID}).Decode(&updatedStatement)
	if err != nil {
		utils.SendResponse(ctx, http.StatusInternalServerError, "error", "Database error: "+err.Error(), nil)
		return
	}

	utils.SendResponse(ctx, http.StatusOK, "success", "Statement updated successfully!", updatedStatement)
}
