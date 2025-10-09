package main

import (
	"log"
	"net/http"
	"os"

	"github.com/WWoi/web-parcer/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	r.GET("/art/:articul", handlers.SearchByArticul)

	server := http.Server{
		Addr:    os.Getenv("SVR_PORT"),
		Handler: r,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
