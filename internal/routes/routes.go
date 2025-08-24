package routes

import (
	"database/sql"

	"agri-management-api/internal/handlers"
	"agri-management-api/internal/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes tüm route'ları ayarlar
func SetupRoutes(r *gin.Engine, db *sql.DB) {
	// Middleware'leri ekle
	r.Use(middleware.RequestID())

	// API v1 router
	v1 := r.Group("/api/v1")
	{
		// Auth routes (public)
		authHandler := handlers.NewAuthHandler(db)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)

			// Protected auth routes
			authProtected := auth.Group("")
			authProtected.Use(middleware.Auth())
			{
				authProtected.GET("/profile", authHandler.GetProfile)
				authProtected.PUT("/profile", authHandler.UpdateProfile)
				authProtected.PUT("/change-password", authHandler.ChangePassword)
				authProtected.POST("/logout", authHandler.Logout)
			}
		}

		// Dashboard routes (protected)
		dashboardHandler := handlers.NewDashboardHandler(db)
		dashboard := v1.Group("/dashboard")
		dashboard.Use(middleware.Auth())
		{
			dashboard.GET("/summary", dashboardHandler.GetSummary)
			dashboard.GET("/recent-activities", dashboardHandler.GetRecentActivities)

			charts := dashboard.Group("/charts")
			{
				charts.GET("/income-expense", dashboardHandler.GetIncomeExpenseChart)
				charts.GET("/production", dashboardHandler.GetProductionChart)
			}
		}

		// Land routes (protected)
		landHandler := handlers.NewLandHandler(db)
		lands := v1.Group("/lands")
		lands.Use(middleware.Auth())
		{
			lands.GET("", landHandler.GetLands)
			lands.POST("", landHandler.CreateLand)
			lands.GET("/:id", landHandler.GetLand)
			lands.PUT("/:id", landHandler.UpdateLand)
			lands.DELETE("/:id", landHandler.DeleteLand)
			lands.GET("/statistics", landHandler.GetLandStatistics)
			lands.GET("/productivity-analysis", landHandler.GetProductivityAnalysis)

			// Land activities
			lands.GET("/:id/activities", landHandler.GetLandActivities)
			lands.POST("/:id/activities", landHandler.CreateLandActivity)
		}

		// Livestock routes (protected)
		livestockHandler := handlers.NewLivestockHandler(db)
		livestock := v1.Group("/livestock")
		livestock.Use(middleware.Auth())
		{
			livestock.GET("", livestockHandler.GetLivestock)
			livestock.POST("", livestockHandler.CreateLivestock)
			livestock.GET("/:id", livestockHandler.GetLivestock)
			livestock.PUT("/:id", livestockHandler.UpdateLivestock)
			livestock.DELETE("/:id", livestockHandler.DeleteLivestock)
			livestock.GET("/statistics", livestockHandler.GetLivestockStatistics)
			livestock.GET("/categories", livestockHandler.GetLivestockCategories)

			// Health records
			livestock.GET("/:id/health-records", livestockHandler.GetHealthRecords)
			livestock.POST("/:id/health-records", livestockHandler.CreateHealthRecord)

			// Milk production
			livestock.GET("/milk-production", livestockHandler.GetMilkProduction)
			livestock.POST("/milk-production", livestockHandler.CreateMilkProduction)
		}

		// Production routes (protected)
		productionHandler := handlers.NewProductionHandler(db)
		production := v1.Group("/production")
		production.Use(middleware.Auth())
		{
			production.GET("", productionHandler.GetProductions)
			production.POST("", productionHandler.CreateProduction)
			production.GET("/:id", productionHandler.GetProduction)
			production.PUT("/:id", productionHandler.UpdateProduction)
			production.DELETE("/:id", productionHandler.DeleteProduction)
			production.GET("/statistics", productionHandler.GetProductionStatistics)
			production.GET("/categories", productionHandler.GetProductionCategories)
		}

		// Finance routes (protected)
		financeHandler := handlers.NewFinanceHandler(db)
		finance := v1.Group("/finance")
		finance.Use(middleware.Auth())
		{
			finance.GET("/summary", financeHandler.GetFinanceSummary)
			finance.GET("/transactions", financeHandler.GetTransactions)
			finance.POST("/transactions", financeHandler.CreateTransaction)
			finance.GET("/transactions/:id", financeHandler.GetTransaction)
			finance.PUT("/transactions/:id", financeHandler.UpdateTransaction)
			finance.DELETE("/transactions/:id", financeHandler.DeleteTransaction)
			finance.GET("/categories", financeHandler.GetCategories)
			finance.GET("/analysis", financeHandler.GetFinanceAnalysis)
		}

		// Calendar routes (protected)
		calendarHandler := handlers.NewCalendarHandler(db)
		calendar := v1.Group("/calendar")
		calendar.Use(middleware.Auth())
		{
			calendar.GET("/events", calendarHandler.GetEvents)
			calendar.POST("/events", calendarHandler.CreateEvent)
			calendar.GET("/events/:id", calendarHandler.GetEvent)
			calendar.PUT("/events/:id", calendarHandler.UpdateEvent)
			calendar.DELETE("/events/:id", calendarHandler.DeleteEvent)
			calendar.PATCH("/events/:id/status", calendarHandler.UpdateEventStatus)
			calendar.GET("/statistics", calendarHandler.GetCalendarStatistics)
		}

		// Notification routes (protected)
		notificationHandler := handlers.NewNotificationHandler(db)
		notifications := v1.Group("/notifications")
		notifications.Use(middleware.Auth())
		{
			notifications.GET("", notificationHandler.GetNotifications)
			notifications.PATCH("/:id/read", notificationHandler.MarkAsRead)
			notifications.PATCH("/mark-all-read", notificationHandler.MarkAllAsRead)
			notifications.DELETE("/:id", notificationHandler.DeleteNotification)
			notifications.GET("/settings", notificationHandler.GetNotificationSettings)
			notifications.PUT("/settings", notificationHandler.UpdateNotificationSettings)
		}

		// Settings routes (protected)
		settingsHandler := handlers.NewSettingsHandler(db)
		settings := v1.Group("/settings")
		settings.Use(middleware.Auth())
		{
			settings.GET("", settingsHandler.GetSettings)
			settings.PUT("", settingsHandler.UpdateSettings)
			settings.GET("/system-info", settingsHandler.GetSystemInfo)
			settings.POST("/backup", settingsHandler.CreateBackup)
			settings.POST("/restore", settingsHandler.RestoreBackup)
		}

		// Weather routes (protected)
		weatherHandler := handlers.NewWeatherHandler(db)
		weather := v1.Group("/weather")
		weather.Use(middleware.Auth())
		{
			weather.GET("/current", weatherHandler.GetCurrentWeather)
			weather.GET("/forecast", weatherHandler.GetWeatherForecast)
			weather.GET("/agricultural-alerts", weatherHandler.GetAgriculturalAlerts)
		}
	}

	// Swagger dokümantasyonu
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Tarım Yönetim Sistemi API çalışıyor",
			"version": "1.0.0",
		})
	})

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Tarım Yönetim Sistemi API'ye Hoş Geldiniz!",
			"version": "1.0.0",
			"docs":    "/swagger/index.html",
			"health":  "/health",
		})
	})
}
