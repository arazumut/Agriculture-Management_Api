package utils

import (
	"database/sql"
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"agri-management-api/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	mathRand = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// GenerateID benzersiz ID oluşturur
func GenerateID() string {
	return uuid.New().String()
}

// HashPassword şifreyi hash'ler
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPassword hash'lenmiş şifreyi kontrol eder
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateRandomString rastgele string oluşturur
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[mathRand.Intn(len(charset))]
	}
	return string(b)
}

// ParsePagination sayfalama parametrelerini parse eder
func ParsePagination(c *gin.Context) (page, limit int) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, _ = strconv.Atoi(pageStr)
	limit, _ = strconv.Atoi(limitStr)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	return page, limit
}

// CalculatePagination sayfalama bilgilerini hesaplar
func CalculatePagination(page, limit, total int) models.Pagination {
	totalPages := (total + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}

	return models.Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}

// SuccessResponse başarılı API yanıtı oluşturur
func SuccessResponse(c *gin.Context, data interface{}, message string) {
	requestID, _ := c.Get("request_id")

	response := models.APIResponse{
		Success: true,
		Data:    data,
		Message: message,
		Meta: &models.APIMeta{
			Timestamp: time.Now().Format(time.RFC3339),
			Version:   "1.0",
			RequestID: requestID.(string),
		},
	}

	c.JSON(http.StatusOK, response)
}

// ErrorResponse hata API yanıtı oluşturur
func ErrorResponse(c *gin.Context, statusCode int, code, message string, details interface{}) {
	requestID, _ := c.Get("request_id")

	response := models.APIResponse{
		Success: false,
		Error: &models.APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
		Meta: &models.APIMeta{
			Timestamp: time.Now().Format(time.RFC3339),
			Version:   "1.0",
			RequestID: requestID.(string),
		},
	}

	c.JSON(statusCode, response)
}

// GetUserID context'ten kullanıcı ID'sini alır
func GetUserID(c *gin.Context) (string, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", errors.New("user_id not found in context")
	}
	return userID.(string), nil
}

// GetUserEmail context'ten kullanıcı email'ini alır
func GetUserEmail(c *gin.Context) (string, error) {
	userEmail, exists := c.Get("user_email")
	if !exists {
		return "", errors.New("user_email not found in context")
	}
	return userEmail.(string), nil
}

// GetUserRole context'ten kullanıcı rolünü alır
func GetUserRole(c *gin.Context) (string, error) {
	userRole, exists := c.Get("user_role")
	if !exists {
		return "", errors.New("user_role not found in context")
	}
	return userRole.(string), nil
}

// ParseTime string'i time.Time'a çevirir
func ParseTime(timeStr string) (*time.Time, error) {
	if timeStr == "" {
		return nil, nil
	}

	// Farklı formatları dene
	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return &t, nil
		}
	}

	return nil, errors.New("invalid time format")
}

// FormatTime time.Time'ı string'e çevirir
func FormatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02T15:04:05Z")
}

// NullStringToPtr sql.NullString'i string pointer'a çevirir
func NullStringToPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

// NullFloat64ToPtr sql.NullFloat64'i float64 pointer'a çevirir
func NullFloat64ToPtr(nf sql.NullFloat64) *float64 {
	if nf.Valid {
		return &nf.Float64
	}
	return nil
}

// NullTimeToPtr sql.NullTime'i time.Time pointer'a çevirir
func NullTimeToPtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

// StringToNullString string'i sql.NullString'e çevirir
func StringToNullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  s != "",
	}
}

// Float64ToNullFloat64 float64'i sql.NullFloat64'e çevirir
func Float64ToNullFloat64(f float64) sql.NullFloat64 {
	return sql.NullFloat64{
		Float64: f,
		Valid:   true,
	}
}

// TimeToNullTime time.Time'i sql.NullTime'e çevirir
func TimeToNullTime(t time.Time) sql.NullTime {
	return sql.NullTime{
		Time:  t,
		Valid: !t.IsZero(),
	}
}

// ValidateEmail email formatını kontrol eder
func ValidateEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// SanitizeString string'i temizler
func SanitizeString(s string) string {
	return strings.TrimSpace(s)
}

// IsEmptyString string'in boş olup olmadığını kontrol eder
func IsEmptyString(s string) bool {
	return strings.TrimSpace(s) == ""
}

// ToJSON interface'i JSON string'e çevirir
func ToJSON(v interface{}) (string, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// FromJSON JSON string'i interface'e çevirir
func FromJSON(data string, v interface{}) error {
	return json.Unmarshal([]byte(data), v)
}
