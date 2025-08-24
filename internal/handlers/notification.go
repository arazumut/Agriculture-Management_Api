package handlers

import (
	"database/sql"
	"net/http"

	"agri-management-api/internal/models"
	"agri-management-api/internal/utils"

	"github.com/gin-gonic/gin"
)

// NotificationHandler bildirim işlemlerini yönetir
type NotificationHandler struct {
	db *sql.DB
}

// NewNotificationHandler yeni notification handler oluşturur
func NewNotificationHandler(db *sql.DB) *NotificationHandler {
	return &NotificationHandler{db: db}
}

// GetNotifications bildirim listesi
// @Summary Bildirim listesi
// @Description Kullanıcının bildirimlerini listeler
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Sayfa numarası"
// @Param limit query int false "Sayfa başına kayıt"
// @Param type query string false "Bildirim türü"
// @Param read query bool false "Okunmuş durumu"
// @Success 200 {object} models.APIResponse{data=map[string]interface{}}
// @Failure 401 {object} models.APIResponse
// @Router /notifications [get]
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	page, limit := utils.ParsePagination(c)
	notificationType := c.DefaultQuery("type", "all")
	read := c.DefaultQuery("read", "all")

	// Sorgu oluştur
	whereClause := "WHERE user_id = ?"
	args := []interface{}{userID}

	if notificationType != "all" {
		whereClause += " AND type = ?"
		args = append(args, notificationType)
	}

	if read == "true" {
		whereClause += " AND is_read = true"
	} else if read == "false" {
		whereClause += " AND is_read = false"
	}

	// Toplam kayıt sayısını al
	var total int
	err = h.db.QueryRow("SELECT COUNT(*) FROM notifications "+whereClause, args...).Scan(&total)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Toplam kayıt sayısı alınamadı", err.Error())
		return
	}

	// Okunmamış bildirim sayısı
	var unreadCount int
	err = h.db.QueryRow("SELECT COUNT(*) FROM notifications WHERE user_id = ? AND is_read = false", userID).Scan(&unreadCount)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Okunmamış bildirim sayısı alınamadı", err.Error())
		return
	}

	// Sayfalama hesapla
	pagination := utils.CalculatePagination(page, limit, total)

	// Bildirimleri getir
	offset := (page - 1) * limit
	query := `
		SELECT id, user_id, title, message, type, priority, is_read, created_at
		FROM notifications ` + whereClause + `
		ORDER BY created_at DESC LIMIT ? OFFSET ?
	`
	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Bildirimler alınamadı", err.Error())
		return
	}
	defer rows.Close()

	var notifications []models.NotificationExtended
	for rows.Next() {
		var notification models.NotificationExtended

		err := rows.Scan(
			&notification.ID, &notification.UserID, &notification.Title, &notification.Message,
			&notification.Type, &notification.Priority, &notification.IsRead, &notification.CreatedAt,
		)
		if err != nil {
			continue
		}

		// İlişkili varlık ve aksiyonlar için basit değerler (gerçek implementasyonda DB'den gelecek)
		notification.RelatedEntity = &models.RelatedEntity{
			Type: "general",
			ID:   "",
			Name: "",
		}

		notification.Actions = []models.Action{
			{Label: "Görüntüle", Action: "view", URL: "/"},
		}

		notifications = append(notifications, notification)
	}

	response := map[string]interface{}{
		"notifications": notifications,
		"pagination":    pagination,
		"unreadCount":   unreadCount,
	}

	utils.SuccessResponse(c, response, "Bildirimler başarıyla getirildi")
}

// MarkAsRead bildirim okundu işaretleme
// @Summary Bildirim okundu işaretleme
// @Description Belirli bir bildirimi okundu olarak işaretler
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Bildirim ID"
// @Success 200 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /notifications/{id}/read [patch]
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	notificationID := c.Param("id")
	if utils.IsEmptyString(notificationID) {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_ID", "Bildirim ID gerekli", nil)
		return
	}

	// Bildirimi okundu olarak işaretle
	result, err := h.db.Exec(`
		UPDATE notifications 
		SET is_read = true 
		WHERE id = ? AND user_id = ?
	`, notificationID, userID)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_ERROR", "Bildirim güncellenemedi", err.Error())
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "NOTIFICATION_NOT_FOUND", "Bildirim bulunamadı", nil)
		return
	}

	utils.SuccessResponse(c, nil, "Bildirim okundu olarak işaretlendi")
}

// MarkAllAsRead tüm bildirimleri okundu işaretleme
// @Summary Tüm bildirimleri okundu işaretleme
// @Description Kullanıcının tüm bildirimlerini okundu olarak işaretler
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /notifications/mark-all-read [patch]
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	// Tüm bildirimleri okundu olarak işaretle
	_, err = h.db.Exec(`
		UPDATE notifications 
		SET is_read = true 
		WHERE user_id = ? AND is_read = false
	`, userID)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_ERROR", "Bildirimler güncellenemedi", err.Error())
		return
	}

	utils.SuccessResponse(c, nil, "Tüm bildirimler okundu olarak işaretlendi")
}

// DeleteNotification bildirim silme
// @Summary Bildirim silme
// @Description Belirli bir bildirimi siler
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Bildirim ID"
// @Success 200 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /notifications/{id} [delete]
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	notificationID := c.Param("id")
	if utils.IsEmptyString(notificationID) {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_ID", "Bildirim ID gerekli", nil)
		return
	}

	// Bildirimi sil
	result, err := h.db.Exec("DELETE FROM notifications WHERE id = ? AND user_id = ?", notificationID, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DELETE_ERROR", "Bildirim silinemedi", err.Error())
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "NOTIFICATION_NOT_FOUND", "Bildirim bulunamadı", nil)
		return
	}

	utils.SuccessResponse(c, nil, "Bildirim başarıyla silindi")
}

// GetNotificationSettings bildirim ayarları
// @Summary Bildirim ayarları
// @Description Kullanıcının bildirim ayarlarını getirir
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=map[string]interface{}}
// @Failure 401 {object} models.APIResponse
// @Router /notifications/settings [get]
func (h *NotificationHandler) GetNotificationSettings(c *gin.Context) {
	_, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	// Bildirim ayarlarını getir (basit implementasyon - gerçek uygulamada DB'den gelecek)
	settings := map[string]interface{}{
		"pushNotifications":  true,
		"emailNotifications": true,
		"smsNotifications":   false,
		"notificationTypes": map[string]bool{
			"reminders": true,
			"alerts":    true,
			"updates":   true,
			"marketing": false,
		},
		"quietHours": map[string]interface{}{
			"enabled":   true,
			"startTime": "22:00",
			"endTime":   "08:00",
		},
	}

	utils.SuccessResponse(c, settings, "Bildirim ayarları başarıyla getirildi")
}

// UpdateNotificationSettings bildirim ayarları güncelleme
// @Summary Bildirim ayarları güncelleme
// @Description Kullanıcının bildirim ayarlarını günceller
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "Bildirim ayarları"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /notifications/settings [put]
func (h *NotificationHandler) UpdateNotificationSettings(c *gin.Context) {
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

	// Bildirim ayarlarını güncelle (basit implementasyon)
	// Gerçek uygulamada bu ayarlar veritabanına kaydedilecek
	utils.SuccessResponse(c, nil, "Bildirim ayarları başarıyla güncellendi")
}

// CreateNotification yeni bildirim oluşturma (dahili kullanım için)
func (h *NotificationHandler) CreateNotification(userID, title, message, notificationType, priority string) error {
	notificationID := utils.GenerateID()

	_, err := h.db.Exec(`
		INSERT INTO notifications (id, user_id, title, message, type, priority, is_read, created_at)
		VALUES (?, ?, ?, ?, ?, ?, false, CURRENT_TIMESTAMP)
	`, notificationID, userID, title, message, notificationType, priority)

	return err
}

// SendWelcomeNotification hoş geldin bildirimi gönder
func (h *NotificationHandler) SendWelcomeNotification(userID string) error {
	return h.CreateNotification(
		userID,
		"Hoş Geldiniz!",
		"Tarım Yönetim Sistemi'ne hoş geldiniz. Başlamak için dashboard'unuzu ziyaret edin.",
		"info",
		"medium",
	)
}

// SendReminderNotification hatırlatıcı bildirimi gönder
func (h *NotificationHandler) SendReminderNotification(userID, title, message string) error {
	return h.CreateNotification(
		userID,
		title,
		message,
		"reminder",
		"high",
	)
}

// SendAlertNotification uyarı bildirimi gönder
func (h *NotificationHandler) SendAlertNotification(userID, title, message string) error {
	return h.CreateNotification(
		userID,
		title,
		message,
		"alert",
		"high",
	)
}
