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

type UserController struct {
	UserColl *mongo.Collection
}

func NewUserController(userColl *mongo.Collection) *UserController {
	return &UserController{
		UserColl: userColl,
	}
}

// AddUser → POST /users
func (uc *UserController) AddUser(ctx iris.Context) {
	var user models.UserRequest
	if err := ctx.ReadJSON(&user); err != nil {
		utils.SendResponse(ctx, http.StatusBadRequest, "error", "Invalid request payload", nil)
		return
	}

	// Use per-request context with timeout
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := uc.UserColl.InsertOne(timeoutCtx, user)
	if err != nil {
		utils.SendResponse(ctx, http.StatusInternalServerError, "error", "Failed to insert user: "+err.Error(), nil)
		return
	}

	utils.SendResponse(ctx, http.StatusOK, "success", "User created successfully!", iris.Map{"user_id": result.InsertedID})
}

// GetAllUsers → GET /users
func (uc *UserController) GetAllUsers(ctx iris.Context) {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := uc.UserColl.Find(timeoutCtx, bson.M{})
	if err != nil {
		utils.SendResponse(ctx, http.StatusInternalServerError, "error", "Failed to fetch users: "+err.Error(), nil)
		return
	}
	defer cursor.Close(timeoutCtx)

	var users []models.UserResponse
	for cursor.Next(timeoutCtx) {
		var user models.UserResponse
		if err := cursor.Decode(&user); err == nil {
			users = append(users, user)
		}
	}

	utils.SendResponse(ctx, http.StatusOK, "success", "Users fetched successfully!", users)
}

// GetUserByID → GET /users/{id}
func (uc *UserController) GetUserByID(ctx iris.Context) {
	id := ctx.Params().Get("id")

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.SendResponse(ctx, http.StatusBadRequest, "error", "Invalid ID", nil)
		return
	}

	var user models.UserResponse
	err = uc.UserColl.FindOne(timeoutCtx, bson.M{"_id": objectID}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		utils.SendResponse(ctx, http.StatusNotFound, "error", "User not found", nil)
		return
	} else if err != nil {
		utils.SendResponse(ctx, http.StatusInternalServerError, "error", "Database error: "+err.Error(), nil)
		return
	}

	utils.SendResponse(ctx, http.StatusOK, "success", "User fetched successfully!", user)
}

// UpdateUserByID → PUT /users/{id}
func (uc *UserController) UpdateUserByID(ctx iris.Context) {
	id := ctx.Params().Get("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.SendResponse(ctx, http.StatusBadRequest, "error", "Invalid ID", nil)
		return
	}

	var updateData models.UserRequest
	if err := ctx.ReadJSON(&updateData); err != nil {
		utils.SendResponse(ctx, http.StatusBadRequest, "error", err.Error(), nil)
		return
	}

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"first_name": updateData.FirstName,
			"last_name":  updateData.LastName,
			"email":      updateData.Email,
			"phone":      updateData.Phone,
			"dob":        updateData.DOB,
			"gender":     updateData.Gender,
			"updated_at": time.Now(),
		},
	}

	res, err := uc.UserColl.UpdateByID(ctxTimeout, objectID, update)
	if err != nil {
		utils.SendResponse(ctx, http.StatusInternalServerError, "error", "Database error: "+err.Error(), nil)
		return
	}

	if res.MatchedCount == 0 {
		utils.SendResponse(ctx, http.StatusNotFound, "error", "User not found", nil)
		return
	}

	var user models.User
	err = uc.UserColl.FindOne(ctxTimeout, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		utils.SendResponse(ctx, http.StatusInternalServerError, "error", "Database error: "+err.Error(), nil)
		return
	}

	utils.SendResponse(ctx, http.StatusOK, "success", "User updated successfully!", user)
}

// DeleteUserByID → DELETE /users/{id}
func (uc *UserController) DeleteUserByID(ctx iris.Context) {
	id := ctx.Params().Get("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.SendResponse(ctx, http.StatusBadRequest, "error", "Invalid ID", nil)
		return
	}

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := uc.UserColl.DeleteOne(ctxTimeout, bson.M{"_id": objectID})
	if err != nil {
		utils.SendResponse(ctx, http.StatusInternalServerError, "error", "Database error: "+err.Error(), nil)
		return
	}

	if res.DeletedCount == 0 {
		utils.SendResponse(ctx, http.StatusNotFound, "error", "User not found", nil)
		return
	}

	utils.SendResponse(ctx, http.StatusOK, "success", "User deleted successfully!", iris.Map{"user_id": objectID})
}
