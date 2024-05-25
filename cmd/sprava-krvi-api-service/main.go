package main

import (
	"log"
	"os"
	"strings"

	"github.com/Marek-FIIT/sprava-krvi-webapi/api"
	"github.com/Marek-FIIT/sprava-krvi-webapi/internal/sprava_krvi"
	"github.com/gin-gonic/gin"

	"context"
	"time"

	"github.com/Marek-FIIT/sprava-krvi-webapi/internal/db_service"
	"github.com/gin-contrib/cors"
)

func main() {
	log.Printf("Server started")
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}
	environment := os.Getenv("API_ENVIRONMENT")
	if !strings.EqualFold(environment, "production") { // case insensitive comparison
		gin.SetMode(gin.DebugMode)
	}
	engine := gin.New()
	engine.Use(gin.Recovery())

	corsMiddleware := cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "PUT", "POST", "DELETE", "PATCH"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{""},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	})
	engine.Use(corsMiddleware)

	// setup context update  middleware
	dbServiceDonors := db_service.NewMongoService[sprava_krvi.Donor](db_service.MongoServiceConfig{Collection: "donor"})
	defer dbServiceDonors.Disconnect(context.Background())
	engine.Use(func(ctx *gin.Context) {
		ctx.Set("db_service_donors", dbServiceDonors)
		ctx.Next()
	})

	dbServiceUnits := db_service.NewMongoService[sprava_krvi.Unit](db_service.MongoServiceConfig{Collection: "unit"})
	defer dbServiceUnits.Disconnect(context.Background())
	engine.Use(func(ctx *gin.Context) {
		ctx.Set("db_service_units", dbServiceUnits)
		ctx.Next()
	})

	// request routings
	sprava_krvi.AddRoutes(engine)
	engine.GET("/openapi", api.HandleOpenApi)
	engine.Run(":" + port)
}
