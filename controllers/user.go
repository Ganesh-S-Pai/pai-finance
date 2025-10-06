package controllers

import (
	// "net/http"

	"context"
	"net/http"
	"time"

	"github.com/Ganesh-S-Pai/pai-finance/models"

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
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	// Use per-request context with timeout
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := uc.UserColl.InsertOne(timeoutCtx, user)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to insert user: " + err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(iris.Map{"data": iris.Map{"user_id": result.InsertedID}, "message": "User created successfully!"})
}

// GetAllUsers → GET /users
func (uc *UserController) GetAllUsers(ctx iris.Context) {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := uc.UserColl.Find(timeoutCtx, bson.M{})
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to fetch users: " + err.Error()})
		return
	}
	defer cursor.Close(timeoutCtx)

	var users []models.User
	for cursor.Next(timeoutCtx) {
		var user models.User
		if err := cursor.Decode(&user); err == nil {
			users = append(users, user)
		}
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(iris.Map{"data": users, "message": "Users fetched successfully!"})
}

// GetUserByID → GET /users/{id}
func (uc *UserController) GetUserByID(ctx iris.Context) {
	id := ctx.Params().Get("id")

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid ID format"})
		return
	}

	var user models.User
	err = uc.UserColl.FindOne(timeoutCtx, bson.M{"_id": objectID}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(iris.Map{"error": "User not found"})
		return
	} else if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Database error: " + err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(iris.Map{"data": user, "message": "User fetched successfully!"})
}

// UpdateUserByID → PUT /users/{id}
func (uc *UserController) UpdateUserByID(ctx iris.Context) {
	id := ctx.Params().Get("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid ID format"})
		return
	}

	var updateData models.UserRequest
	if err := ctx.ReadJSON(&updateData); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"first_name": updateData.FirstName,
			"last_name":  updateData.LastName,
			"email":      updateData.Email,
			"password":   updateData.Password,
			"phone":      updateData.Phone,
			"dob":        updateData.DOB,
			"gender":     updateData.Gender,
			"updated_at": time.Now(),
		},
	}

	res, err := uc.UserColl.UpdateByID(ctxTimeout, objectID, update)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Database error: " + err.Error()})
		return
	}

	if res.MatchedCount == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(iris.Map{"error": "User not found"})
		return
	}

	var user models.User
	err = uc.UserColl.FindOne(ctxTimeout, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Database error: " + err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(iris.Map{"message": "User updated successfully!", "data": user})
}

// DeleteUserByID → DELETE /users/{id}
func (uc *UserController) DeleteUserByID(ctx iris.Context) {
	id := ctx.Params().Get("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid ID format"})
		return
	}

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := uc.UserColl.DeleteOne(ctxTimeout, bson.M{"_id": objectID})
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Database error: " + err.Error()})
		return
	}

	if res.DeletedCount == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(iris.Map{"error": "User not found"})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(iris.Map{"data": objectID, "message": "Deleted successfully"})
}
