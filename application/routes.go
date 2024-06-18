package application

import (
	"net/http"

	"github.com/OPC-16/RMS-server/handler"
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

   a.router = router
}
