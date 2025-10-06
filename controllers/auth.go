package controllers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/Ganesh-S-Pai/pai-finance/models"
	"github.com/Ganesh-S-Pai/pai-finance/utils"
	"github.com/kataras/iris/v12"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct {
	UserColl      *mongo.Collection
	TokenTTLHours int
}

type SignupRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	DOB       string `json:"dob"`
	Gender    string `json:"gender"`
	Phone     string `json:"phone"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewAuthController(userColl *mongo.Collection) *AuthController {
	return &AuthController{
		UserColl:      userColl,
		TokenTTLHours: 72,
	}
}

func (ac *AuthController) Signup(ctx iris.Context) {
	var req SignupRequest
	if err := ctx.ReadJSON(&req); err != nil {
		utils.SendResponseAndStop(ctx, http.StatusBadRequest, "error", "Invalid request payload", nil)
		return
	}

	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" || req.DOB == "" || req.Gender == "" || req.Phone == "" {
		utils.SendResponseAndStop(ctx, http.StatusBadRequest, "error", "first_name, last_name, DOB, gender, email, phone and password are required", nil)
		return
	}

	var dob time.Time
	p, err := time.Parse("2006-01-02", req.DOB)
	if err != nil {
		utils.SendResponseAndStop(ctx, http.StatusBadRequest, "error", "dob must be in YYYY-MM-DD format", nil)
		return
	}
	if p.After(time.Now()) {
		utils.SendResponseAndStop(ctx, http.StatusBadRequest, "error", "dob cannot be in the future", nil)
		return
	}
	dob = p

	cctx, cancel := context.WithTimeout(ctx.Request().Context(), 6*time.Second)
	defer cancel()

	count, err := ac.UserColl.CountDocuments(cctx, bson.M{"email": strings.ToLower(req.Email)})
	if err != nil {
		utils.SendResponseAndStop(ctx, http.StatusInternalServerError, "error", "database error", nil)
		return
	}
	if count > 0 {
		utils.SendResponseAndStop(ctx, http.StatusConflict, "error", "email already registered", nil)
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.SendResponseAndStop(ctx, http.StatusInternalServerError, "error", "failed to hash password", nil)
		return
	}

	user := models.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     strings.ToLower(req.Email),
		Password:  string(hashed),
		DOB:       dob,
		Gender:    req.Gender,
		Phone:     req.Phone,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	res, err := ac.UserColl.InsertOne(cctx, user)
	if err != nil {
		utils.SendResponseAndStop(ctx, http.StatusInternalServerError, "error", "failed to create user", nil)
		return
	}
	oid := res.InsertedID.(primitive.ObjectID)

	token, err := utils.CreateToken(oid, ac.TokenTTLHours)
	if err != nil {
		utils.SendResponseAndStop(ctx, http.StatusInternalServerError, "error", "failed to create token", nil)
		return
	}

	utils.SendResponse(ctx, http.StatusOK, "success", "Sign up successful!", iris.Map{
		"token": token,
		"user": iris.Map{
			"id":         oid.Hex(),
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"email":      user.Email,
			"phone":      user.Phone,
			"dob": func() string {
				if user.DOB.IsZero() {
					return ""
				}
				return user.DOB.Format("2006-01-02")
			}(),
			"gender": user.Gender,
		},
	})
}

func (ac *AuthController) Login(ctx iris.Context) {
	var req LoginRequest
	if err := ctx.ReadJSON(&req); err != nil {
		utils.SendResponseAndStop(ctx, http.StatusBadRequest, "error", "Invalid request payload", nil)
		return
	}

	if req.Email == "" || req.Password == "" {
		utils.SendResponseAndStop(ctx, http.StatusBadRequest, "error", "email and password required", nil)
		return
	}

	cctx, cancel := context.WithTimeout(ctx.Request().Context(), 6*time.Second)
	defer cancel()

	var user models.User
	err := ac.UserColl.FindOne(cctx, bson.M{"email": strings.ToLower(req.Email)}).Decode(&user)
	if err != nil {
		utils.SendResponseAndStop(ctx, http.StatusUnauthorized, "error", "Invalid credentials", nil)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		utils.SendResponseAndStop(ctx, http.StatusUnauthorized, "error", "Invalid credentials", nil)
		return
	}

	token, err := utils.CreateToken(user.ID, ac.TokenTTLHours)
	if err != nil {
		utils.SendResponseAndStop(ctx, http.StatusInternalServerError, "error", "Failed to create token", nil)
		return
	}

	utils.SendResponse(ctx, http.StatusOK, "success", "Login successful!", iris.Map{
		"token": token,
		"user": iris.Map{
			"id":         user.ID.Hex(),
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"email":      user.Email,
			"phone":      user.Phone,
			"dob": func() string {
				if user.DOB.IsZero() {
					return ""
				}
				return user.DOB.Format("2006-01-02")
			}(),
			"gender": user.Gender,
		},
	})
}

// Iris middleware: requires Authorization: Bearer <token>
// Usage: app.UseRouter(authMiddleware) or route.Use(authMiddleware)
func AuthMiddleware(userColl *mongo.Collection) iris.Handler {
	return func(ctx iris.Context) {
		auth := ctx.GetHeader("Authorization")
		if auth == "" {
			utils.SendResponse(ctx, http.StatusUnauthorized, "error", "missing authorization header", nil)
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.SendResponse(ctx, http.StatusUnauthorized, "error", "invalid authorization header", nil)
			return
		}
		tokenStr := parts[1]
		userID, err := utils.ValidateToken(tokenStr)
		if err != nil {
			utils.SendResponse(ctx, http.StatusUnauthorized, "error", "invalid token", nil)
			return
		}
		ctx.Values().Set("userID", userID)
		ctx.Next()
	}
}