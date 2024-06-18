package application

import (
	"os"
)

type Config struct {
   RedisAddress string
   ServerPort   string     //input to Echo server for its port number is in string
}

func LoadConfig() Config {
   //default values
   cfg := Config{
      RedisAddress: "localhost:6379",
      ServerPort: "3000",
   }

   //loading any config from env varibles
   if redisAddr, exists := os.LookupEnv("REDIS_ADDR"); exists {
      cfg.RedisAddress = redisAddr
   }
   if serverPort, exists := os.LookupEnv("SERVER_PORT"); exists {
      cfg.ServerPort = serverPort
   }

   return cfg
}
