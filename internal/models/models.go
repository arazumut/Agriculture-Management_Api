package models

import (
	"time"
)

// User kullanıcı modeli
type User struct {
	ID         string    `json:"id" db:"id"`
	Name       string    `json:"name" db:"name"`
	Email      string    `json:"email" db:"email"`
	Password   string    `json:"-" db:"password"`
	Avatar     string    `json:"avatar" db:"avatar"`
	Role       string    `json:"role" db:"role"`
	FarmName   string    `json:"farmName" db:"farm_name"`
	Location   string    `json:"location" db:"location"`
	IsVerified bool      `json:"isVerified" db:"is_verified"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time `json:"updatedAt" db:"updated_at"`
}

// Land arazi modeli
type Land struct {
	ID             string     `json:"id" db:"id"`
	UserID         string     `json:"userId" db:"user_id"`
	Name           string     `json:"name" db:"name"`
	Area           float64    `json:"area" db:"area"`
	Unit           string     `json:"unit" db:"unit"`
	Crop           string     `json:"crop" db:"crop"`
	Status         string     `json:"status" db:"status"`
	LastActivity   *time.Time `json:"lastActivity" db:"last_activity"`
	Productivity   float64    `json:"productivity" db:"productivity"`
	Location       Location   `json:"location" db:"-"`
	SoilType       string     `json:"soilType" db:"soil_type"`
	IrrigationType string     `json:"irrigationType" db:"irrigation_type"`
	CreatedAt      time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time  `json:"updatedAt" db:"updated_at"`
}

// Location konum modeli
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address"`
}

// Livestock hayvan modeli
type Livestock struct {
	ID           string     `json:"id" db:"id"`
	UserID       string     `json:"userId" db:"user_id"`
	TagNumber    string     `json:"tagNumber" db:"tag_number"`
	Type         string     `json:"type" db:"type"`
	Breed        string     `json:"breed" db:"breed"`
	Gender       string     `json:"gender" db:"gender"`
	BirthDate    *time.Time `json:"birthDate" db:"birth_date"`
	Weight       *float64   `json:"weight" db:"weight"`
	HealthStatus string     `json:"healthStatus" db:"health_status"`
	Location     string     `json:"location" db:"location"`
	Mother       string     `json:"mother" db:"mother"`
	Father       string     `json:"father" db:"father"`
	Notes        string     `json:"notes" db:"notes"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time  `json:"updatedAt" db:"updated_at"`
}

// Production üretim modeli
type Production struct {
	ID              string     `json:"id" db:"id"`
	UserID          string     `json:"userId" db:"user_id"`
	LandID          *string    `json:"landId" db:"land_id"`
	Name            string     `json:"name" db:"name"`
	Category        string     `json:"category" db:"category"`
	Amount          float64    `json:"amount" db:"amount"`
	Unit            string     `json:"unit" db:"unit"`
	HarvestDate     *time.Time `json:"harvestDate" db:"harvest_date"`
	Quality         string     `json:"quality" db:"quality"`
	StorageLocation string     `json:"storageLocation" db:"storage_location"`
	Status          string     `json:"status" db:"status"`
	Price           *float64   `json:"price" db:"price"`
	Notes           string     `json:"notes" db:"notes"`
	CreatedAt       time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time  `json:"updatedAt" db:"updated_at"`
}

// Transaction finansal işlem modeli
type Transaction struct {
	ID            string    `json:"id" db:"id"`
	UserID        string    `json:"userId" db:"user_id"`
	Type          string    `json:"type" db:"type"`
	Category      string    `json:"category" db:"category"`
	Description   string    `json:"description" db:"description"`
	Amount        float64   `json:"amount" db:"amount"`
	Currency      string    `json:"currency" db:"currency"`
	Date          time.Time `json:"date" db:"date"`
	Status        string    `json:"status" db:"status"`
	PaymentMethod string    `json:"paymentMethod" db:"payment_method"`
	Receipt       string    `json:"receipt" db:"receipt"`
	Notes         string    `json:"notes" db:"notes"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time `json:"updatedAt" db:"updated_at"`
}

// Event etkinlik modeli
type Event struct {
	ID                string     `json:"id" db:"id"`
	UserID            string     `json:"userId" db:"user_id"`
	Title             string     `json:"title" db:"title"`
	Description       string     `json:"description" db:"description"`
	Type              string     `json:"type" db:"type"`
	StartDate         time.Time  `json:"startDate" db:"start_date"`
	EndDate           *time.Time `json:"endDate" db:"end_date"`
	IsAllDay          bool       `json:"isAllDay" db:"is_all_day"`
	Status            string     `json:"status" db:"status"`
	Priority          string     `json:"priority" db:"priority"`
	Location          string     `json:"location" db:"location"`
	RelatedEntityType *string    `json:"relatedEntityType" db:"related_entity_type"`
	RelatedEntityID   *string    `json:"relatedEntityId" db:"related_entity_id"`
	CreatedAt         time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt         time.Time  `json:"updatedAt" db:"updated_at"`
}

// Notification bildirim modeli
type Notification struct {
	ID                string    `json:"id" db:"id"`
	UserID            string    `json:"userId" db:"user_id"`
	Title             string    `json:"title" db:"title"`
	Message           string    `json:"message" db:"message"`
	Type              string    `json:"type" db:"type"`
	Priority          string    `json:"priority" db:"priority"`
	IsRead            bool      `json:"isRead" db:"is_read"`
	RelatedEntityType *string   `json:"relatedEntityType" db:"related_entity_type"`
	RelatedEntityID   *string   `json:"relatedEntityId" db:"related_entity_id"`
	CreatedAt         time.Time `json:"createdAt" db:"created_at"`
}

// HealthRecord sağlık kaydı modeli
type HealthRecord struct {
	ID           string     `json:"id" db:"id"`
	LivestockID  string     `json:"livestockId" db:"livestock_id"`
	Type         string     `json:"type" db:"type"`
	Description  string     `json:"description" db:"description"`
	Date         time.Time  `json:"date" db:"date"`
	Veterinarian string     `json:"veterinarian" db:"veterinarian"`
	Cost         *float64   `json:"cost" db:"cost"`
	Notes        string     `json:"notes" db:"notes"`
	NextCheckup  *time.Time `json:"nextCheckup" db:"next_checkup"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
}

// MilkProduction süt üretim modeli
type MilkProduction struct {
	ID          string    `json:"id" db:"id"`
	LivestockID string    `json:"livestockId" db:"livestock_id"`
	Date        time.Time `json:"date" db:"date"`
	Amount      float64   `json:"amount" db:"amount"`
	Quality     string    `json:"quality" db:"quality"`
	Notes       string    `json:"notes" db:"notes"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
}

// LandActivity arazi aktivitesi modeli
type LandActivity struct {
	ID            string     `json:"id" db:"id"`
	LandID        string     `json:"landId" db:"land_id"`
	Type          string     `json:"type" db:"type"`
	Description   string     `json:"description" db:"description"`
	ScheduledDate *time.Time `json:"scheduledDate" db:"scheduled_date"`
	ActualDate    *time.Time `json:"actualDate" db:"actual_date"`
	Notes         string     `json:"notes" db:"notes"`
	Cost          *float64   `json:"cost" db:"cost"`
	Result        string     `json:"result" db:"result"`
	CreatedAt     time.Time  `json:"createdAt" db:"created_at"`
}

// Request/Response modelleri

// LoginRequest giriş isteği
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest kayıt isteği
type RegisterRequest struct {
	Name            string `json:"name" binding:"required"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
	FarmName        string `json:"farmName" binding:"required"`
	Location        string `json:"location" binding:"required"`
}

// AuthResponse kimlik doğrulama yanıtı
type AuthResponse struct {
	User         User   `json:"user"`
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

// DashboardSummary dashboard özet verileri
type DashboardSummary struct {
	TotalAnimals   AnimalSummary  `json:"totalAnimals"`
	TotalLands     LandSummary    `json:"totalLands"`
	MonthlyIncome  FinanceSummary `json:"monthlyIncome"`
	MonthlyExpense FinanceSummary `json:"monthlyExpense"`
	ActiveProducts ProductSummary `json:"activeProducts"`
}

// AnimalSummary hayvan özeti
type AnimalSummary struct {
	Count      int     `json:"count"`
	Trend      string  `json:"trend"`
	Percentage float64 `json:"percentage"`
}

// LandSummary arazi özeti
type LandSummary struct {
	Area         float64 `json:"area"`
	Count        int     `json:"count"`
	Productivity float64 `json:"productivity"`
}

// FinanceSummary finans özeti
type FinanceSummary struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Trend    string  `json:"trend"`
}

// ProductSummary ürün özeti
type ProductSummary struct {
	Count      int `json:"count"`
	Categories int `json:"categories"`
}

// Pagination sayfalama
type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// APIResponse genel API yanıt formatı
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    *APIMeta    `json:"meta,omitempty"`
}

// APIError API hata formatı
type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// APIMeta API meta bilgileri
type APIMeta struct {
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
	RequestID string `json:"requestId"`
}
