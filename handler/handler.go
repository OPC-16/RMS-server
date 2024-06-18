package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/OPC-16/RMS-server/model"
	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
)

func Signup(c echo.Context) error {
   //retrieve the redis client from the context
   rdb, ok := c.Get("redis").(*redis.Client)
   if !ok {
      return c.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to get Redis client from context"})
   }

   user := new (model.User)

   // Bind the request body to the user instance
   if err := c.Bind(user); err != nil {
      return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid Input"})
   }

   //perform basic validation
   if user.Name == "" || user.Password == "" {
      return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name and Password are required"})
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
