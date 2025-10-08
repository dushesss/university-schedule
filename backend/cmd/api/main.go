package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	addr := getenv("HTTP_ADDR", ":8080")

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "schedule-analyzer",
			"status":  "ok",
		})
	})

	log.Println("API запущен на", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
