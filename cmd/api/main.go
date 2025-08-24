// Package main Tarım Yönetim Sistemi API
// @title Tarım Yönetim Sistemi API
// @version 1.0
// @description Flutter mobil uygulaması için Tarım Yönetim Sistemi REST API
// @termsOfService http://swagger.io/terms/

// @contact.name API Destek
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT token ile kimlik doğrulama

package main

import (
	"log"
	"os"

	"agri-management-api/docs"
	"agri-management-api/internal/database"
	"agri-management-api/internal/middleware"
	"agri-management-api/internal/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// @title Tarım Yönetim Sistemi API
// @version 1.0
// @description Flutter mobil uygulaması için Tarım Yönetim Sistemi REST API
// @termsOfService http://swagger.io/terms/

// @contact.name API Destek
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT token ile kimlik doğrulama

func main() {
	// Environment değişkenlerini yükle
	if err := godotenv.Load("config.env"); err != nil {
		log.Println("config.env dosyası bulunamadı, varsayılan değerler kullanılıyor")
	}

	// Veritabanını başlat
	db, err := database.InitDB()
	if err != nil {
		log.Fatal("Veritabanı başlatılamadı:", err)
	}
	defer db.Close()

	// Gin router'ı oluştur
	gin.SetMode(gin.ReleaseMode)
	if os.Getenv("ENV") == "development" {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.Default()

	// Middleware'leri ekle
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())

	// Routes'ları ayarla
	routes.SetupRoutes(r, db)

	// Swagger dokümantasyonu
	docs.SwaggerInfo.Title = "Tarım Yönetim Sistemi API"
	docs.SwaggerInfo.Description = "Flutter mobil uygulaması için Tarım Yönetim Sistemi REST API"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Tarım Yönetim Sistemi API başlatılıyor... Port: %s", port)
	log.Printf("📚 Swagger dokümantasyonu: http://localhost:%s/swagger/index.html", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Sunucu başlatılamadı:", err)
	}
}
