package handlers

import (
	"database/sql"
	"net/http"

	"agri-management-api/internal/models"
	"agri-management-api/internal/utils"

	"github.com/gin-gonic/gin"
)

// LandHandler arazi işlemlerini yönetir
type LandHandler struct {
	db *sql.DB
}

// NewLandHandler yeni land handler oluşturur
func NewLandHandler(db *sql.DB) *LandHandler {
	return &LandHandler{db: db}
}

// GetLands arazi listesi
// @Summary Arazi listesi
// @Description Kullanıcının ar// GetLandActivities arazi aktiviteleri
// @Summary Arazi aktiviteleri
// @Descri// CreateLandActivity arazi aktivitesi oluşturma
// @Summary Arazi aktivitesi oluşturma
// @Description Arazi için yeni aktivite oluşturur
// @Tags Lands
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Arazi ID"
// @Param request body models.LandActivityRecord true "Aktivite bilgileri"
// @Success 201 {object} models.APIResponse{data=models.LandActivityRecord}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /lands/{id}/activities [post]in aktiviteleri listeler
// @Tags Lands
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Arazi ID"
// @Success 200 {object} models.APIResponse{data=[]models.LandActivityRecord}
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /lands/{id}/activities [get]steler
// @Tags Lands
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Sayfa numarası"
// @Param limit query int false "Sayfa başına kayıt"
// @Param status query string false "Arazi durumu"
// @Success 200 {object} models.APIResponse{data=map[string]interface{}}
// @Failure 401 {object} models.APIResponse
// @Router /lands [get]
func (h *LandHandler) GetLands(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	page, limit := utils.ParsePagination(c)
	status := c.DefaultQuery("status", "all")

	// Toplam kayıt sayısını al
	var total int
	whereClause := "WHERE user_id = ?"
	args := []interface{}{userID}

	if status != "all" {
		whereClause += " AND status = ?"
		args = append(args, status)
	}

	err = h.db.QueryRow("SELECT COUNT(*) FROM lands "+whereClause, args...).Scan(&total)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Toplam kayıt sayısı alınamadı", err.Error())
		return
	}

	// Sayfalama hesapla
	pagination := utils.CalculatePagination(page, limit, total)

	// Arazileri getir
	offset := (page - 1) * limit
	query := `
		SELECT id, user_id, name, area, unit, crop, status, last_activity, 
		       productivity, latitude, longitude, address, soil_type, irrigation_type,
		       created_at, updated_at
		FROM lands ` + whereClause + `
		ORDER BY created_at DESC LIMIT ? OFFSET ?
	`
	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Araziler alınamadı", err.Error())
		return
	}
	defer rows.Close()

	var lands []models.Land
	for rows.Next() {
		var land models.Land
		var lastActivity sql.NullTime
		var latitude, longitude sql.NullFloat64
		var address string

		err := rows.Scan(
			&land.ID, &land.UserID, &land.Name, &land.Area, &land.Unit, &land.Crop,
			&land.Status, &lastActivity, &land.Productivity, &latitude, &longitude,
			&address, &land.SoilType, &land.IrrigationType, &land.CreatedAt, &land.UpdatedAt,
		)
		if err != nil {
			continue
		}

		land.LastActivity = utils.NullTimeToPtr(lastActivity)
		if latitude.Valid && longitude.Valid {
			land.Location = models.Location{
				Latitude:  latitude.Float64,
				Longitude: longitude.Float64,
				Address:   "",
			}
		}

		lands = append(lands, land)
	}

	response := map[string]interface{}{
		"lands":      lands,
		"pagination": pagination,
	}

	utils.SuccessResponse(c, response, "Araziler başarıyla getirildi")
}

// CreateLand yeni arazi oluşturma
// @Summary Yeni arazi oluşturma
// @Description Yeni arazi kaydı oluşturur
// @Tags Lands
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.Land true "Arazi bilgileri"
// @Success 201 {object} models.APIResponse{data=models.Land}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /lands [post]
func (h *LandHandler) CreateLand(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	var req models.Land
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	// Gerekli alanları kontrol et
	if utils.IsEmptyString(req.Name) || req.Area <= 0 || utils.IsEmptyString(req.Unit) {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_FIELDS", "Gerekli alanlar eksik", nil)
		return
	}

	landID := utils.GenerateID()

	// Araziyi oluştur
	_, err = h.db.Exec(`
		INSERT INTO lands (id, user_id, name, area, unit, crop, status, productivity,
		                  latitude, longitude, address, soil_type, irrigation_type,
		                  created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, 'active', 0, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, landID, userID, req.Name, req.Area, req.Unit, req.Crop,
		req.Location.Latitude, req.Location.Longitude, req.Location.Address,
		req.SoilType, req.IrrigationType)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Arazi oluşturulamadı", err.Error())
		return
	}

	// Oluşturulan araziyi getir
	var land models.Land
	var latitude, longitude sql.NullFloat64
	var address string
	err = h.db.QueryRow(`
		SELECT id, user_id, name, area, unit, crop, status, last_activity, 
		       productivity, latitude, longitude, address, soil_type, irrigation_type,
		       created_at, updated_at
		FROM lands WHERE id = ?
	`, landID).Scan(
		&land.ID, &land.UserID, &land.Name, &land.Area, &land.Unit, &land.Crop,
		&land.Status, &land.LastActivity, &land.Productivity, &latitude, &longitude,
		&address, &land.SoilType, &land.IrrigationType, &land.CreatedAt, &land.UpdatedAt,
	)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "FETCH_ERROR", "Oluşturulan arazi getirilemedi", err.Error())
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    land,
		Message: "Arazi başarıyla oluşturuldu",
	})
}

// GetLand arazi detayları
// @Summary Arazi detayları
// @Description Belirli bir arazinin detaylarını getirir
// @Tags Lands
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Arazi ID"
// @Success 200 {object} models.APIResponse{data=models.Land}
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /lands/{id} [get]
func (h *LandHandler) GetLand(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	landID := c.Param("id")
	if utils.IsEmptyString(landID) {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_ID", "Arazi ID gerekli", nil)
		return
	}

	var land models.Land
	var lastActivity sql.NullTime
	var latitude, longitude sql.NullFloat64
	var address string

	err = h.db.QueryRow(`
		SELECT id, user_id, name, area, unit, crop, status, last_activity, 
		       productivity, latitude, longitude, address, soil_type, irrigation_type,
		       created_at, updated_at
		FROM lands WHERE id = ? AND user_id = ?
	`, landID, userID).Scan(
		&land.ID, &land.UserID, &land.Name, &land.Area, &land.Unit, &land.Crop,
		&land.Status, &lastActivity, &land.Productivity, &latitude, &longitude,
		&address, &land.SoilType, &land.IrrigationType, &land.CreatedAt, &land.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.ErrorResponse(c, http.StatusNotFound, "LAND_NOT_FOUND", "Arazi bulunamadı", nil)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Arazi getirilemedi", err.Error())
		}
		return
	}

	land.LastActivity = utils.NullTimeToPtr(lastActivity)
	if latitude.Valid && longitude.Valid {
		land.Location = models.Location{
			Latitude:  latitude.Float64,
			Longitude: longitude.Float64,
			Address:   address,
		}
	} else {
		land.Location = models.Location{
			Address: address,
		}
	}

	utils.SuccessResponse(c, land, "Arazi detayları başarıyla getirildi")
}

// UpdateLand arazi güncelleme
// @Summary Arazi güncelleme
// @Description Mevcut arazi bilgilerini günceller
// @Tags Lands
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Arazi ID"
// @Param request body models.Land true "Güncellenecek arazi bilgileri"
// @Success 200 {object} models.APIResponse{data=models.Land}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /lands/{id} [put]
func (h *LandHandler) UpdateLand(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	landID := c.Param("id")
	if utils.IsEmptyString(landID) {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_ID", "Arazi ID gerekli", nil)
		return
	}

	var req models.Land
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	// Araziyi güncelle
	_, err = h.db.Exec(`
		UPDATE lands 
		SET name = ?, area = ?, unit = ?, crop = ?, status = ?, productivity = ?,
		    latitude = ?, longitude = ?, address = ?, soil_type = ?, irrigation_type = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ?
	`, req.Name, req.Area, req.Unit, req.Crop, req.Status, req.Productivity,
		req.Location.Latitude, req.Location.Longitude, req.Location.Address,
		req.SoilType, req.IrrigationType, landID, userID)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_ERROR", "Arazi güncellenemedi", err.Error())
		return
	}

	// Güncellenmiş araziyi getir
	h.GetLand(c)
}

// DeleteLand arazi silme
// @Summary Arazi silme
// @Description Belirli bir araziyi siler
// @Tags Lands
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Arazi ID"
// @Success 200 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /lands/{id} [delete]
func (h *LandHandler) DeleteLand(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	landID := c.Param("id")
	if utils.IsEmptyString(landID) {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_ID", "Arazi ID gerekli", nil)
		return
	}

	// Araziyi sil
	result, err := h.db.Exec("DELETE FROM lands WHERE id = ? AND user_id = ?", landID, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DELETE_ERROR", "Arazi silinemedi", err.Error())
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "LAND_NOT_FOUND", "Arazi bulunamadı", nil)
		return
	}

	utils.SuccessResponse(c, nil, "Arazi başarıyla silindi")
}

// GetLandStatistics arazi istatistikleri
// @Summary Arazi istatistikleri
// @Description Arazi istatistiklerini getirir
// @Tags Lands
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=map[string]interface{}}
// @Failure 401 {object} models.APIResponse
// @Router /lands/statistics [get]
func (h *LandHandler) GetLandStatistics(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	// İstatistikleri hesapla
	var totalArea float64
	var totalLands int
	var avgProductivity float64
	var activeCrops int

	err = h.db.QueryRow(`
		SELECT COALESCE(SUM(area), 0), COUNT(*), COALESCE(AVG(productivity), 0),
		       COUNT(DISTINCT CASE WHEN crop IS NOT NULL AND crop != '' THEN crop END)
		FROM lands WHERE user_id = ?
	`, userID).Scan(&totalArea, &totalLands, &avgProductivity, &activeCrops)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "İstatistikler alınamadı", err.Error())
		return
	}

	// Durum bazında arazi sayıları
	var activeLands, inactiveLands, maintenanceLands int

	h.db.QueryRow("SELECT COUNT(*) FROM lands WHERE user_id = ? AND status = 'active'", userID).Scan(&activeLands)
	h.db.QueryRow("SELECT COUNT(*) FROM lands WHERE user_id = ? AND status = 'inactive'", userID).Scan(&inactiveLands)
	h.db.QueryRow("SELECT COUNT(*) FROM lands WHERE user_id = ? AND status = 'maintenance'", userID).Scan(&maintenanceLands)

	statistics := map[string]interface{}{
		"totalArea":           totalArea,
		"totalLands":          totalLands,
		"averageProductivity": avgProductivity,
		"activeCrops":         activeCrops,
		"landsByStatus": map[string]int{
			"active":      activeLands,
			"inactive":    inactiveLands,
			"maintenance": maintenanceLands,
		},
	}

	utils.SuccessResponse(c, statistics, "Arazi istatistikleri başarıyla getirildi")
}

// GetProductivityAnalysis verimlilik analizi
// @Summary Verimlilik analizi
// @Description Arazi verimlilik analizini getirir
// @Tags Lands
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param period query string false "Analiz periyodu"
// @Success 200 {object} models.APIResponse{data=map[string]interface{}}
// @Failure 401 {object} models.APIResponse
// @Router /lands/productivity-analysis [get]
func (h *LandHandler) GetProductivityAnalysis(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	period := c.DefaultQuery("period", "month")

	// Verimlilik analizi (basit implementasyon)
	var avgProductivity float64
	var maxProductivity float64
	var minProductivity float64

	err = h.db.QueryRow(`
		SELECT COALESCE(AVG(productivity), 0), COALESCE(MAX(productivity), 0), COALESCE(MIN(productivity), 0)
		FROM lands WHERE user_id = ? AND productivity > 0
	`, userID).Scan(&avgProductivity, &maxProductivity, &minProductivity)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Verimlilik analizi alınamadı", err.Error())
		return
	}

	analysis := map[string]interface{}{
		"period":              period,
		"averageProductivity": avgProductivity,
		"maxProductivity":     maxProductivity,
		"minProductivity":     minProductivity,
		"totalLands":          0, // Bu değer daha sonra hesaplanabilir
	}

	utils.SuccessResponse(c, analysis, "Verimlilik analizi başarıyla getirildi")
}

// GetLandActivities arazi aktiviteleri
// @Summary Arazi aktiviteleri
// @Description Belirli bir arazinin aktivitelerini listeler
// @Tags Lands
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Arazi ID"
// @Success 200 {object} models.APIResponse{data=[]models.LandActivityRecord}
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /lands/{id}/activities [get]
func (h *LandHandler) GetLandActivities(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	landID := c.Param("id")
	if utils.IsEmptyString(landID) {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_ID", "Arazi ID gerekli", nil)
		return
	}

	// Arazi kullanıcıya ait mi kontrol et
	var exists bool
	err = h.db.QueryRow("SELECT 1 FROM lands WHERE id = ? AND user_id = ?", landID, userID).Scan(&exists)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "LAND_NOT_FOUND", "Arazi bulunamadı", nil)
		return
	}

	// Aktivite listesini getir
	rows, err := h.db.Query(`
		SELECT id, land_id, type, description, scheduled_date, actual_date,
		       notes, cost, result, created_at
		FROM land_activities WHERE land_id = ?
		ORDER BY created_at DESC
	`, landID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Aktivite listesi alınamadı", err.Error())
		return
	}
	defer rows.Close()

	var activities []models.LandActivityRecord
	for rows.Next() {
		var activity models.LandActivityRecord
		var scheduledDate, actualDate sql.NullTime
		var cost sql.NullFloat64

		err := rows.Scan(
			&activity.ID, &activity.LandID, &activity.Type, &activity.Description,
			&scheduledDate, &actualDate, &activity.Notes, &cost, &activity.Result, &activity.CreatedAt,
		)
		if err != nil {
			continue
		}

		activity.ScheduledDate = utils.NullTimeToPtr(scheduledDate)
		activity.ActualDate = utils.NullTimeToPtr(actualDate)
		activity.Cost = utils.NullFloat64ToPtr(cost)

		activities = append(activities, activity)
	}

	utils.SuccessResponse(c, activities, "Arazi aktiviteleri başarıyla getirildi")
}

// CreateLandActivity arazi aktivitesi oluşturma
// @Summary Arazi aktivitesi oluşturma
// @Description Yeni arazi aktivitesi kaydı oluşturur
// @Tags Lands
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Arazi ID"
// @Param request body models.LandActivityRecord true "Aktivite bilgileri"
// @Success 201 {object} models.APIResponse{data=models.LandActivityRecord}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /lands/{id}/activities [post]
func (h *LandHandler) CreateLandActivity(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	landID := c.Param("id")
	if utils.IsEmptyString(landID) {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_ID", "Arazi ID gerekli", nil)
		return
	}

	var req models.LandActivityRecord
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	// Arazi kullanıcıya ait mi kontrol et
	var exists bool
	err = h.db.QueryRow("SELECT 1 FROM lands WHERE id = ? AND user_id = ?", landID, userID).Scan(&exists)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "LAND_NOT_FOUND", "Arazi bulunamadı", nil)
		return
	}

	// Aktiviteyi oluştur
	activityID := utils.GenerateID()
	_, err = h.db.Exec(`
		INSERT INTO land_activities (id, land_id, type, description, scheduled_date,
		                           actual_date, notes, cost, result, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, activityID, landID, req.Type, req.Description, req.ScheduledDate,
		req.ActualDate, req.Notes, req.Cost, req.Result)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Aktivite oluşturulamadı", err.Error())
		return
	}

	// Oluşturulan aktiviteyi getir
	var activity models.LandActivityRecord
	var scheduledDate, actualDate sql.NullTime
	var cost sql.NullFloat64

	err = h.db.QueryRow(`
		SELECT id, land_id, type, description, scheduled_date, actual_date,
		       notes, cost, result, created_at
		FROM land_activities WHERE id = ?
	`, activityID).Scan(
		&activity.ID, &activity.LandID, &activity.Type, &activity.Description,
		&scheduledDate, &actualDate, &activity.Notes, &cost, &activity.Result, &activity.CreatedAt,
	)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "FETCH_ERROR", "Oluşturulan aktivite getirilemedi", err.Error())
		return
	}

	activity.ScheduledDate = utils.NullTimeToPtr(scheduledDate)
	activity.ActualDate = utils.NullTimeToPtr(actualDate)
	activity.Cost = utils.NullFloat64ToPtr(cost)

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    activity,
		Message: "Arazi aktivitesi başarıyla oluşturuldu",
	})
}
