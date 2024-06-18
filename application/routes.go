package application

import (
	"fmt"
	"net/http"

	"github.com/OPC-16/RMS-server/handler"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (a *App) loadRoutes() {
   router := echo.New()

   router.Use(middleware.Logger())
   router.Use(middleware.Recover())

   // Middleware to inject Redis Client into context
   router.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
      return func(c echo.Context) error {
         c.Set("redis", a.rdb)
         return next(c)
      }
   })

   // setting up a hello world route for "/"
   router.GET("/", func(c echo.Context) error {
      return c.String(http.StatusOK, "Hello, World!")
   })

   // setting up the main routes
   router.POST("/signup", handler.Signup)
   router.POST("/login", handler.Login)
   router.POST("/uploadResume", handler.UploadResume, jwtMiddleware)

   a.router = router
}

// a middleware to verify JWT tokens for protected routes
func jwtMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        authHeader := c.Request().Header.Get("Authorization")
        if authHeader == "" {
            return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized, (empty token)"})
        }

        tokenString := authHeader[len("Bearer "):]

        // Validate token
        jwtToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            // Check token signing method
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
            }
            return []byte("secret"), nil
        })
        if err != nil || !jwtToken.Valid {
            return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
        }

        // extract claims
        claims, ok := jwtToken.Claims.(jwt.MapClaims)
        if !ok || !jwtToken.Valid {
            return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
        }

        //check UserType
        if claims["usertype"] != "Applicant" {
            return c.JSON(http.StatusForbidden, map[string]string{"error": "Forbidden"})
        }

        // Token is valid and user type is 'Applicant', proceed to the next handler
        return next(c)
    }
}
