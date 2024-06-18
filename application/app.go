package application

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
)

type App struct {
   router  *echo.Echo
   rdb     *redis.Client
   config  Config
}

func New(config Config) *App {
   app := &App{
      rdb: redis.NewClient(&redis.Options{
         Addr: config.RedisAddress,
      }),
      config: config,
   }

   app.loadRoutes()

   return app
}

func (a *App) Start(ctx context.Context) error {
   //checking if redis is connected, if not we don't even start the server
   err := a.rdb.Ping(ctx).Err()
   if err != nil {
      return fmt.Errorf("failed to connect to redis: %w", err)
   }

   defer func() {
      if err := a.rdb.Close(); err != nil {
         fmt.Println("Failed to close redis:", err)
      }
   }()

   fmt.Println("Starting the server...")

   serverPort := a.config.ServerPort
   serverPort = ":" + serverPort

   a.router.Logger.Fatal(a.router.Start(serverPort))

   return nil
}
