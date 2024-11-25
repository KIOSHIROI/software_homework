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
	localIP := "0.0.0.0"

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:8000", localIP),
		Handler:      routers.InitRouter(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("Server is running on http://" + localIP + ":8000")
	fmt.Println("Successful Setup.")
	server.ListenAndServe()
}
