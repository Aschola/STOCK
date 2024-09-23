package main

import (
	"github.com/labstack/echo/v4"
	"log"
	"os"
	"stock/db"
	"stock/routes"
)

func main() {

	db.Init() 
	
	e := echo.New()

	routes.RegisterRoutes(e) 
	routes.SetupRoutes(e)


	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" 
	}
	log.Fatal(e.Start(":" + port))
}
