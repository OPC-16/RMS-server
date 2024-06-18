package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/OPC-16/RMS-server/model"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func Signup(c echo.Context) error {
   //retrieve the redis client from the context
   rdb, ok := c.Get("redis").(*redis.Client)
   if !ok {
      return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get Redis client from context"})
   }

   user := new (model.User)

   // Bind the request body to the user instance
   if err := c.Bind(user); err != nil {
      return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid Input"})
   }

   //perform basic validation
   if user.Name == "" || user.Email == "" || user.Password == "" {
      return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name, Email and Password are required"})
   }

   // Serialize user to JSON
   userJson, err := json.Marshal(user)
   if err != nil {
      return fmt.Errorf("could not marshal user: %v", err)
   }

   // Extract the context.Context from echo.Context
   ctx := c.Request().Context()
   // Store JSON string in Redis with email as key
   err = rdb.Set(ctx, "user:" + user.Email, userJson, 0).Err()
   if err != nil {
      return c.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to store user data in Redis"})
   }

   //respond with success
   return c.JSON(http.StatusOK, map[string]string{"message": "User signed up successfully"})
}

func Login(c echo.Context) error {
   type LoginRequest struct {
      Email    string `json:"email"`
      Password string `json:"password"`
   }

   req := new(LoginRequest)
   if err := c.Bind(req); err != nil {
      return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid Input"})
   }

   //Retrieve user details from redis
   ctx := c.Request().Context()
   rdb, ok := c.Get("redis").(*redis.Client)
   if !ok {
      return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get Redis client from context"})
   }

   userJson, err := rdb.Get(ctx, "user:" + req.Email).Result()
   if err != nil {
      return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user details"})
   }

   //unmarshal user json
   var user model.User
   if err := json.Unmarshal([]byte(userJson), &user); err != nil {
      return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to unmarshal user data"})
   }

   //compare the passwords
   if req.Password != user.Password {
      return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid Credentials"})
   }

   //create JWT token
   token := jwt.New(jwt.SigningMethodHS256)
   claims := token.Claims.(jwt.MapClaims)
   claims["email"] = user.Email
   claims["exp"] = time.Now().Add(time.Hour * 24).Unix() // Token expires in 24 hours

   //sign the token with a secret
   tokenString, err := token.SignedString([]byte("secret"))
   if err != nil {
      return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate the token"})
   }

   // respond with JWT token
   return c.JSON(http.StatusOK, map[string]string{"token": tokenString})
}

func UploadResume(c echo.Context) error {

}
