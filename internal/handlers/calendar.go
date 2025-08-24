package handlers

import (
	"database/sql"
	"net/http"

	"agri-management-api/internal/models"
	"agri-management-api/internal/utils"

	"github.com/gin-gonic/gin"
)

// CalendarHandler takvim işlemlerini yönetir
type CalendarHandler struct {
	db *sql.DB
}

// NewCalendarHandler yeni calendar handler oluşturur
func NewCalendarHandler(db *sql.DB) *CalendarHandler {
	return &CalendarHandler{db: db}
}

// GetEvents etkinlik listesi
// @Summary Etkinlik listesi
// @Description Takvim etkinliklerini listeler
// @Tags Calendar
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param startDate query string false "Başlangıç tarihi"
// @Param endDate query string false "Bitiş tarihi"
// @Param type query string false "Etkinlik türü"
// @Param status query string false "Etkinlik durumu"
// @Success 200 {object} models.APIResponse{data=[]models.Event}
// @Failure 401 {object} models.APIResponse
// @Router /calendar/events [get]
func (h *CalendarHandler) GetEvents(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	startDate := c.DefaultQuery("startDate", "")
	endDate := c.DefaultQuery("endDate", "")
	eventType := c.DefaultQuery("type", "all")
	status := c.DefaultQuery("status", "all")

	// Sorgu oluştur
	whereClause := "WHERE user_id = ?"
	args := []interface{}{userID}

	if startDate != "" {
		whereClause += " AND start_date >= ?"
		args = append(args, startDate)
	}

	if endDate != "" {
		whereClause += " AND end_date <= ?"
		args = append(args, endDate)
	}

	if eventType != "all" {
		whereClause += " AND type = ?"
		args = append(args, eventType)
	}

	if status != "all" {
		whereClause += " AND status = ?"
		args = append(args, status)
	}

	// Etkinlikleri getir
	rows, err := h.db.Query(`
		SELECT id, user_id, title, description, type, start_date, end_date, is_all_day,
		       status, priority, location, created_at, updated_at
		FROM events `+whereClause+`
		ORDER BY start_date ASC
	`, args...)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Etkinlikler alınamadı", err.Error())
		return
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		var startDate, endDate sql.NullTime

		err := rows.Scan(
			&event.ID, &event.UserID, &event.Title, &event.Description, &event.Type,
			&startDate, &endDate, &event.IsAllDay, &event.Status, &event.Priority,
			&event.Location, &event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			continue
		}

		event.StartDate = utils.NullTimeToPtr(startDate)
		event.EndDate = utils.NullTimeToPtr(endDate)

		// İlişkili varlık bilgilerini getir (basit implementasyon)
		event.RelatedEntity = &models.RelatedEntity{
			Type: "general",
			ID:   "",
			Name: "",
		}

		// Hatırlatıcıları getir (basit implementasyon)
		event.Reminders = []models.Reminder{
			{Time: 30, Method: "notification"},
		}

		events = append(events, event)
	}

	utils.SuccessResponse(c, events, "Etkinlikler başarıyla getirildi")
}

// CreateEvent yeni etkinlik ekleme
// @Summary Yeni etkinlik ekleme
// @Description Yeni takvim etkinliği oluşturur
// @Tags Calendar
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.Event true "Etkinlik bilgileri"
// @Success 201 {object} models.APIResponse{data=models.Event}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /calendar/events [post]
func (h *CalendarHandler) CreateEvent(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	var req models.Event
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	// Gerekli alanları kontrol et
	if utils.IsEmptyString(req.Title) || utils.IsEmptyString(req.Type) {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_FIELDS", "Gerekli alanlar eksik", nil)
		return
	}

	eventID := utils.GenerateID()

	// Etkinliği oluştur
	_, err = h.db.Exec(`
		INSERT INTO events (id, user_id, title, description, type, start_date, end_date,
		                   is_all_day, status, priority, location, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'pending', ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, eventID, userID, req.Title, req.Description, req.Type, req.StartDate, req.EndDate,
		req.IsAllDay, req.Priority, req.Location)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Etkinlik oluşturulamadı", err.Error())
		return
	}

	// Oluşturulan etkinliği getir
	var event models.Event
	var startDate, endDate sql.NullTime

	err = h.db.QueryRow(`
		SELECT id, user_id, title, description, type, start_date, end_date, is_all_day,
		       status, priority, location, created_at, updated_at
		FROM events WHERE id = ?
	`, eventID).Scan(
		&event.ID, &event.UserID, &event.Title, &event.Description, &event.Type,
		&startDate, &endDate, &event.IsAllDay, &event.Status, &event.Priority,
		&event.Location, &event.CreatedAt, &event.UpdatedAt,
	)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "FETCH_ERROR", "Oluşturulan etkinlik getirilemedi", err.Error())
		return
	}

	event.StartDate = utils.NullTimeToPtr(startDate)
	event.EndDate = utils.NullTimeToPtr(endDate)

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    event,
		Message: "Etkinlik başarıyla oluşturuldu",
	})
}

// GetEvent etkinlik detayları
// @Summary Etkinlik detayları
// @Description Belirli bir etkinliğin detaylarını getirir
// @Tags Calendar
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Etkinlik ID"
// @Success 200 {object} models.APIResponse{data=models.Event}
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /calendar/events/{id} [get]
func (h *CalendarHandler) GetEvent(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	eventID := c.Param("id")
	if utils.IsEmptyString(eventID) {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_ID", "Etkinlik ID gerekli", nil)
		return
	}

	var event models.Event
	var startDate, endDate sql.NullTime

	err = h.db.QueryRow(`
		SELECT id, user_id, title, description, type, start_date, end_date, is_all_day,
		       status, priority, location, created_at, updated_at
		FROM events WHERE id = ? AND user_id = ?
	`, eventID, userID).Scan(
		&event.ID, &event.UserID, &event.Title, &event.Description, &event.Type,
		&startDate, &endDate, &event.IsAllDay, &event.Status, &event.Priority,
		&event.Location, &event.CreatedAt, &event.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.ErrorResponse(c, http.StatusNotFound, "EVENT_NOT_FOUND", "Etkinlik bulunamadı", nil)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Etkinlik getirilemedi", err.Error())
		}
		return
	}

	event.StartDate = utils.NullTimeToPtr(startDate)
	event.EndDate = utils.NullTimeToPtr(endDate)

	utils.SuccessResponse(c, event, "Etkinlik detayları başarıyla getirildi")
}

// UpdateEvent etkinlik güncelleme
// @Summary Etkinlik güncelleme
// @Description Mevcut etkinlik bilgilerini günceller
// @Tags Calendar
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Etkinlik ID"
// @Param request body models.Event true "Güncellenecek etkinlik bilgileri"
// @Success 200 {object} models.APIResponse{data=models.Event}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /calendar/events/{id} [put]
func (h *CalendarHandler) UpdateEvent(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	eventID := c.Param("id")
	if utils.IsEmptyString(eventID) {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_ID", "Etkinlik ID gerekli", nil)
		return
	}

	var req models.Event
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	// Etkinliği güncelle
	_, err = h.db.Exec(`
		UPDATE events 
		SET title = ?, description = ?, type = ?, start_date = ?, end_date = ?,
		    is_all_day = ?, status = ?, priority = ?, location = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ?
	`, req.Title, req.Description, req.Type, req.StartDate, req.EndDate,
		req.IsAllDay, req.Status, req.Priority, req.Location, eventID, userID)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_ERROR", "Etkinlik güncellenemedi", err.Error())
		return
	}

	// Güncellenmiş etkinliği getir
	h.GetEvent(c)
}

// DeleteEvent etkinlik silme
// @Summary Etkinlik silme
// @Description Belirli bir etkinliği siler
// @Tags Calendar
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Etkinlik ID"
// @Success 200 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /calendar/events/{id} [delete]
func (h *CalendarHandler) DeleteEvent(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	eventID := c.Param("id")
	if utils.IsEmptyString(eventID) {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_ID", "Etkinlik ID gerekli", nil)
		return
	}

	// Etkinliği sil
	result, err := h.db.Exec("DELETE FROM events WHERE id = ? AND user_id = ?", eventID, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DELETE_ERROR", "Etkinlik silinemedi", err.Error())
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "EVENT_NOT_FOUND", "Etkinlik bulunamadı", nil)
		return
	}

	utils.SuccessResponse(c, nil, "Etkinlik başarıyla silindi")
}

// UpdateEventStatus etkinlik durumu güncelleme
// @Summary Etkinlik durumu güncelleme
// @Description Etkinlik durumunu günceller
// @Tags Calendar
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Etkinlik ID"
// @Param request body map[string]string true "Durum bilgileri"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /calendar/events/{id}/status [patch]
func (h *CalendarHandler) UpdateEventStatus(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	eventID := c.Param("id")
	if utils.IsEmptyString(eventID) {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_ID", "Etkinlik ID gerekli", nil)
		return
	}

	var req struct {
		Status string `json:"status"`
		Notes  string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	// Etkinlik durumunu güncelle
	_, err = h.db.Exec(`
		UPDATE events 
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ?
	`, req.Status, eventID, userID)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_ERROR", "Etkinlik durumu güncellenemedi", err.Error())
		return
	}

	utils.SuccessResponse(c, nil, "Etkinlik durumu başarıyla güncellendi")
}

// GetCalendarStatistics takvim istatistikleri
// @Summary Takvim istatistikleri
// @Description Takvim istatistiklerini getirir
// @Tags Calendar
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param period query string false "Periyot"
// @Success 200 {object} models.APIResponse{data=map[string]interface{}}
// @Failure 401 {object} models.APIResponse
// @Router /calendar/statistics [get]
func (h *CalendarHandler) GetCalendarStatistics(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	// Toplam etkinlik sayısı
	var totalEvents int
	err = h.db.QueryRow("SELECT COUNT(*) FROM events WHERE user_id = ?", userID).Scan(&totalEvents)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Toplam etkinlik sayısı alınamadı", err.Error())
		return
	}

	// Tamamlanan etkinlikler
	var completedEvents int
	err = h.db.QueryRow("SELECT COUNT(*) FROM events WHERE user_id = ? AND status = 'completed'", userID).Scan(&completedEvents)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Tamamlanan etkinlik sayısı alınamadı", err.Error())
		return
	}

	// Bekleyen etkinlikler
	var pendingEvents int
	err = h.db.QueryRow("SELECT COUNT(*) FROM events WHERE user_id = ? AND status = 'pending'", userID).Scan(&pendingEvents)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Bekleyen etkinlik sayısı alınamadı", err.Error())
		return
	}

	// Bugünün etkinlikleri
	var todayEvents int
	err = h.db.QueryRow("SELECT COUNT(*) FROM events WHERE user_id = ? AND DATE(start_date) = CURDATE()", userID).Scan(&todayEvents)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Bugünün etkinlik sayısı alınamadı", err.Error())
		return
	}

	// Yaklaşan etkinlikler (gelecek 7 gün)
	var upcomingEvents int
	err = h.db.QueryRow(`
		SELECT COUNT(*) FROM events 
		WHERE user_id = ? AND start_date > NOW() AND start_date <= DATE_ADD(NOW(), INTERVAL 7 DAY)
	`, userID).Scan(&upcomingEvents)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Yaklaşan etkinlik sayısı alınamadı", err.Error())
		return
	}

	// Tür bazında etkinlik sayıları
	rows, err := h.db.Query(`
		SELECT type, COUNT(*) as count
		FROM events WHERE user_id = ?
		GROUP BY type
	`, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Tür bazında etkinlik sayıları alınamadı", err.Error())
		return
	}
	defer rows.Close()

	var eventsByType []map[string]interface{}
	for rows.Next() {
		var eventType string
		var count int

		err := rows.Scan(&eventType, &count)
		if err != nil {
			continue
		}

		eventsByType = append(eventsByType, map[string]interface{}{
			"type":  eventType,
			"count": count,
		})
	}

	statistics := map[string]interface{}{
		"totalEvents":     totalEvents,
		"completedEvents": completedEvents,
		"pendingEvents":   pendingEvents,
		"todayEvents":     todayEvents,
		"upcomingEvents":  upcomingEvents,
		"eventsByType":    eventsByType,
	}

	utils.SuccessResponse(c, statistics, "Takvim istatistikleri başarıyla getirildi")
}
