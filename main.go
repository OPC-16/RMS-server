package main

import (
	"context"
	"log"

	"github.com/OPC-16/RMS-server/application"
)

func main() {
   app := application.New(application.LoadConfig())

   ctx := context.Background()

   err := app.Start(ctx)
   if err != nil {
      log.Fatal("Failed to start the app:", err)
   }
}
