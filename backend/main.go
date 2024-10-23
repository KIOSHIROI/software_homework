package main

import (
	"backend/models"
	"backend/routers"
	"fmt"
	"net/http"
	"time"
)

func init() {
	models.Setup()
}

func main() {
	server := &http.Server{
		Addr:         ":8000",
		Handler:      routers.InitRouter(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	fmt.Println("Successful Setup.")
	server.ListenAndServe()
}
