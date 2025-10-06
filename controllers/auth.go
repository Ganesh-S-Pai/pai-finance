package controllers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/kataras/iris/v12"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"vhiw-sales-log/models"
	"vhiw-sales-log/utils"
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
		ctx.StopWithJSON(http.StatusBadRequest, iris.Map{"error": "invalid request payload"})
		return
	}

	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" || req.DOB == "" || req.Gender == "" || req.Phone == "" {
		ctx.StopWithJSON(http.StatusBadRequest, iris.Map{"error": "first_name, last_name, DOB, gender, email and password are required"})
		return
	}

	var dob time.Time
	p, err := time.Parse("2006-01-02", req.DOB)
	if err != nil {
		ctx.StopWithJSON(http.StatusBadRequest, iris.Map{"error": "dob must be in YYYY-MM-DD format"})
		return
	}
	if p.After(time.Now()) {
		ctx.StopWithJSON(http.StatusBadRequest, iris.Map{"error": "dob cannot be in the future"})
		return
	}
	dob = p

	cctx, cancel := context.WithTimeout(ctx.Request().Context(), 6*time.Second)
	defer cancel()

	count, err := ac.UserColl.CountDocuments(cctx, bson.M{"email": strings.ToLower(req.Email)})
	if err != nil {
		ctx.StopWithJSON(http.StatusInternalServerError, iris.Map{"error": "database error"})
		return
	}
	if count > 0 {
		ctx.StopWithJSON(http.StatusConflict, iris.Map{"error": "email already registered"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		ctx.StopWithJSON(http.StatusInternalServerError, iris.Map{"error": "failed to hash password"})
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
		ctx.StopWithJSON(http.StatusInternalServerError, iris.Map{"error": "failed to create user"})
		return
	}
	oid := res.InsertedID.(primitive.ObjectID)

	token, err := utils.CreateToken(oid, ac.TokenTTLHours)
	if err != nil {
		ctx.StopWithJSON(http.StatusInternalServerError, iris.Map{"error": "failed to create token"})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(iris.Map{
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
		ctx.StopWithJSON(http.StatusBadRequest, iris.Map{"error": "invalid request payload"})
		return
	}

	if req.Email == "" || req.Password == "" {
		ctx.StopWithJSON(http.StatusBadRequest, iris.Map{"error": "email and password required"})
		return
	}

	cctx, cancel := context.WithTimeout(ctx.Request().Context(), 6*time.Second)
	defer cancel()

	var user models.User
	err := ac.UserColl.FindOne(cctx, bson.M{"email": strings.ToLower(req.Email)}).Decode(&user)
	if err != nil {
		ctx.StopWithJSON(http.StatusUnauthorized, iris.Map{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		ctx.StopWithJSON(http.StatusUnauthorized, iris.Map{"error": "invalid credentials"})
		return
	}

	token, err := utils.CreateToken(user.ID, ac.TokenTTLHours)
	if err != nil {
		ctx.StopWithJSON(http.StatusInternalServerError, iris.Map{"error": "failed to create token"})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(iris.Map{
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
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(iris.Map{"error": "missing authorization header"})
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(iris.Map{"error": "invalid authorization header"})
			return
		}
		tokenStr := parts[1]
		userID, err := utils.ValidateToken(tokenStr)
		if err != nil {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(iris.Map{"error": "invalid token"})
			return
		}
		// optionally: load user from DB and attach to ctx.Values()
		ctx.Values().Set("userID", userID)
		ctx.Next()
	}
}

// Helper to extract userID from ctx (when used in other handlers)
func GetUserIDFromCtx(ctx iris.Context) (primitive.ObjectID, bool) {
	v := ctx.Values().Get("userID")
	if v == nil {
		return primitive.NilObjectID, false
	}
	oid, ok := v.(primitive.ObjectID)
	return oid, ok
}
