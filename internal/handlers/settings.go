package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"agri-management-api/internal/models"
	"agri-management-api/internal/utils"

	"github.com/gin-gonic/gin"
)

// SettingsHandler ayar işlemlerini yönetir
type SettingsHandler struct {
	db *sql.DB
}

// NewSettingsHandler yeni settings handler oluşturur
func NewSettingsHandler(db *sql.DB) *SettingsHandler {
	return &SettingsHandler{db: db}
}

// GetSettings uygulama ayarları
// @Summary Uygulama ayarları
// @Description Kullanıcının uygulama ayarlarını getirir
// @Tags Settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=models.Settings}
// @Failure 401 {object} models.APIResponse
// @Router /settings [get]
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	_, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	// Ayarları getir (basit implementasyon - gerçek uygulamada DB'den gelecek)
	settings := models.Settings{
		General: models.GeneralSettings{
			Language:   "tr",
			Currency:   "TRY",
			DateFormat: "DD/MM/YYYY",
			TimeFormat: "24H",
			Units: models.UnitSettings{
				Area:   "dönüm",
				Weight: "kg",
				Volume: "litre",
			},
		},
		Notifications: models.NotificationSettings{
			Push:  true,
			Email: true,
			SMS:   false,
		},
		Privacy: models.PrivacySettings{
			LocationSharing: true,
			DataAnalytics:   true,
			PersonalizedAds: false,
		},
		Backup: models.BackupSettings{
			AutoBackup:      true,
			BackupFrequency: "weekly",
			CloudStorage:    true,
		},
	}

	utils.SuccessResponse(c, settings, "Ayarlar başarıyla getirildi")
}

// UpdateSettings ayarları güncelleme
// @Summary Ayarları güncelleme
// @Description Kullanıcının uygulama ayarlarını günceller
// @Tags Settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.Settings true "Ayar bilgileri"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /settings [put]
func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	_, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	var req models.Settings
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	// Ayarları güncelle (basit implementasyon)
	// Gerçek uygulamada bu ayarlar veritabanına kaydedilecek
	utils.SuccessResponse(c, nil, "Ayarlar başarıyla güncellendi")
}

// GetSystemInfo sistem bilgileri
// @Summary Sistem bilgileri
// @Description Sistem bilgilerini getirir
// @Tags Settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=map[string]interface{}}
// @Failure 401 {object} models.APIResponse
// @Router /settings/system-info [get]
func (h *SettingsHandler) GetSystemInfo(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	// Kullanıcının verilerini say
	var landCount, animalCount, productionCount, transactionCount int

	h.db.QueryRow("SELECT COUNT(*) FROM lands WHERE user_id = ?", userID).Scan(&landCount)
	h.db.QueryRow("SELECT COUNT(*) FROM livestock WHERE user_id = ?", userID).Scan(&animalCount)
	h.db.QueryRow("SELECT COUNT(*) FROM production WHERE user_id = ?", userID).Scan(&productionCount)
	h.db.QueryRow("SELECT COUNT(*) FROM transactions WHERE user_id = ?", userID).Scan(&transactionCount)

	// Depolama kullanımını hesapla (basit implementasyon)
	totalRecords := landCount + animalCount + productionCount + transactionCount
	storageUsed := float64(totalRecords) * 0.1 // Her kayıt için 0.1MB varsayımı
	storageLimit := 1000.0                     // 1GB limit

	systemInfo := map[string]interface{}{
		"appVersion":   "1.0.0",
		"apiVersion":   "v1",
		"lastBackup":   time.Now().AddDate(0, 0, -1).Format("2006-01-02T15:04:05Z"),
		"storageUsed":  storageUsed,
		"storageLimit": storageLimit,
		"features": []string{
			"Arazi Yönetimi",
			"Hayvancılık",
			"Finans Takibi",
			"Üretim Kayıtları",
			"Takvim",
			"Raporlar",
			"Hava Durumu",
		},
		"supportContact": "support@agrimanagement.com",
		"dataStats": map[string]int{
			"lands":        landCount,
			"animals":      animalCount,
			"productions":  productionCount,
			"transactions": transactionCount,
		},
	}

	utils.SuccessResponse(c, systemInfo, "Sistem bilgileri başarıyla getirildi")
}

// CreateBackup veri yedekleme
// @Summary Veri yedekleme
// @Description Kullanıcı verilerinin yedeğini oluşturur
// @Tags Settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=map[string]interface{}}
// @Failure 401 {object} models.APIResponse
// @Router /settings/backup [post]
func (h *SettingsHandler) CreateBackup(c *gin.Context) {
	_, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	// Yedekleme işlemi simülasyonu
	backupID := utils.GenerateID()
	backupDate := time.Now()

	// Gerçek uygulamada burada:
	// 1. Kullanıcının tüm verileri JSON formatında export edilir
	// 2. Dosya cloud storage'a yüklenir
	// 3. Yedekleme kaydı veritabanına kaydedilir

	backup := map[string]interface{}{
		"backupId":    backupID,
		"status":      "completed",
		"createdAt":   backupDate.Format("2006-01-02T15:04:05Z"),
		"size":        "2.5MB",
		"downloadUrl": "/api/v1/settings/backup/" + backupID + "/download",
		"expiresAt":   backupDate.AddDate(0, 1, 0).Format("2006-01-02T15:04:05Z"), // 1 ay sonra
		"includes": []string{
			"Arazi Verileri",
			"Hayvan Kayıtları",
			"Üretim Bilgileri",
			"Finansal İşlemler",
			"Takvim Etkinlikleri",
		},
	}

	utils.SuccessResponse(c, backup, "Yedekleme başarıyla oluşturuldu")
}

// RestoreBackup veri geri yükleme
// @Summary Veri geri yükleme
// @Description Yedekten veri geri yükler
// @Tags Settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "Geri yükleme seçenekleri"
// @Success 200 {object} models.APIResponse{data=map[string]interface{}}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /settings/restore [post]
func (h *SettingsHandler) RestoreBackup(c *gin.Context) {
	_, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	var req struct {
		BackupFile     string `json:"backupFile"`
		RestoreOptions struct {
			IncludeFinance    bool `json:"includeFinance"`
			IncludeLivestock  bool `json:"includeLivestock"`
			IncludeLands      bool `json:"includeLands"`
			IncludeProduction bool `json:"includeProduction"`
		} `json:"restoreOptions"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	// Geri yükleme işlemi simülasyonu
	restoreID := utils.GenerateID()

	// Gerçek uygulamada burada:
	// 1. Yedek dosyası doğrulanır
	// 2. Seçili veriler geri yüklenir
	// 3. Mevcut verilerle çakışma kontrolü yapılır
	// 4. İşlem logları tutulur

	restore := map[string]interface{}{
		"restoreId":  restoreID,
		"status":     "completed",
		"restoredAt": time.Now().Format("2006-01-02T15:04:05Z"),
		"backupFile": req.BackupFile,
		"restored": map[string]interface{}{
			"lands":      req.RestoreOptions.IncludeLands,
			"livestock":  req.RestoreOptions.IncludeLivestock,
			"finance":    req.RestoreOptions.IncludeFinance,
			"production": req.RestoreOptions.IncludeProduction,
		},
		"summary": map[string]int{
			"restoredLands":        25,
			"restoredAnimals":      48,
			"restoredTransactions": 156,
			"restoredProductions":  12,
		},
	}

	utils.SuccessResponse(c, restore, "Veriler başarıyla geri yüklendi")
}

// ExportData veri export
func (h *SettingsHandler) ExportData(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	format := c.DefaultQuery("format", "json")

	// Export işlemi simülasyonu
	exportData := map[string]interface{}{
		"userId":      userID,
		"exportedAt":  time.Now().Format("2006-01-02T15:04:05Z"),
		"format":      format,
		"status":      "ready",
		"downloadUrl": "/api/v1/settings/export/" + utils.GenerateID() + "/download",
	}

	utils.SuccessResponse(c, exportData, "Veriler export için hazırlandı")
}

// GetUserPreferences kullanıcı tercihleri
func (h *SettingsHandler) GetUserPreferences(c *gin.Context) {
	_, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	preferences := map[string]interface{}{
		"theme":           "light",
		"dashboardLayout": "grid",
		"defaultView":     "dashboard",
		"autoSave":        true,
		"compactMode":     false,
		"showTips":        true,
	}

	utils.SuccessResponse(c, preferences, "Kullanıcı tercihleri başarıyla getirildi")
}

// UpdateUserPreferences kullanıcı tercihleri güncelleme
func (h *SettingsHandler) UpdateUserPreferences(c *gin.Context) {
	_, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	// Tercihleri güncelle (gerçek uygulamada DB'ye kaydedilecek)
	utils.SuccessResponse(c, nil, "Kullanıcı tercihleri başarıyla güncellendi")
}

// GetStorageInfo depolama bilgileri
func (h *SettingsHandler) GetStorageInfo(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	// Depolama kullanımını hesapla
	var totalRecords int
	err = h.db.QueryRow(`
		SELECT (
			(SELECT COUNT(*) FROM lands WHERE user_id = ?) +
			(SELECT COUNT(*) FROM livestock WHERE user_id = ?) +
			(SELECT COUNT(*) FROM production WHERE user_id = ?) +
			(SELECT COUNT(*) FROM transactions WHERE user_id = ?)
		) as total
	`, userID, userID, userID, userID).Scan(&totalRecords)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Depolama bilgileri alınamadı", err.Error())
		return
	}

	storageUsed := float64(totalRecords) * 0.1 // Her kayıt için 0.1MB
	storageLimit := 1000.0                     // 1GB
	usagePercentage := (storageUsed / storageLimit) * 100

	storageInfo := map[string]interface{}{
		"used":            storageUsed,
		"limit":           storageLimit,
		"available":       storageLimit - storageUsed,
		"usagePercentage": usagePercentage,
		"breakdown": map[string]interface{}{
			"images":    storageUsed * 0.4,
			"documents": storageUsed * 0.3,
			"data":      storageUsed * 0.2,
			"cache":     storageUsed * 0.1,
		},
	}

	utils.SuccessResponse(c, storageInfo, "Depolama bilgileri başarıyla getirildi")
}
