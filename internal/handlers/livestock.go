package handlers

import (
	"database/sql"
	"net/http"

	"agri-management-api/internal/models"
	"agri-management-api/internal/utils"

	"github.com/gin-gonic/gin"
)

// LivestockHandler hayvan işlemlerini yönetir
type LivestockHandler struct {
	db *sql.DB
}

// NewLivestockHandler yeni livestock handler oluşturur
func NewLivestockHandler(db *sql.DB) *LivestockHandler {
	return &LivestockHandler{db: db}
}

// GetLivestock hayvan listesi
// @Summary Hayvan listesi
// @Description Kullanıcının hayvanlarını listeler
// @Tags Livestock
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Sayfa numarası"
// @Param limit query int false "Sayfa başına kayıt"
// @Param type query string false "Hayvan türü"
// @Param status query string false "Sağlık durumu"
// @Success 200 {object} models.APIResponse{data=map[string]interface{}}
// @Failure 401 {object} models.APIResponse
// @Router /livestock [get]
func (h *LivestockHandler) GetLivestock(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	page, limit := utils.ParsePagination(c)
	animalType := c.DefaultQuery("type", "all")
	status := c.DefaultQuery("status", "all")

	// Toplam kayıt sayısını al
	var total int
	whereClause := "WHERE user_id = ?"
	args := []interface{}{userID}

	if animalType != "all" {
		whereClause += " AND type = ?"
		args = append(args, animalType)
	}

	if status != "all" {
		whereClause += " AND health_status = ?"
		args = append(args, status)
	}

	err = h.db.QueryRow("SELECT COUNT(*) FROM livestock "+whereClause, args...).Scan(&total)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Toplam kayıt sayısı alınamadı", err.Error())
		return
	}

	// Sayfalama hesapla
	pagination := utils.CalculatePagination(page, limit, total)

	// Hayvanları getir
	offset := (page - 1) * limit
	query := `
		SELECT id, user_id, tag_number, type, breed, gender, birth_date, weight,
		       health_status, location, mother, father, notes, created_at, updated_at
		FROM livestock ` + whereClause + `
		ORDER BY created_at DESC LIMIT ? OFFSET ?
	`
	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Hayvanlar alınamadı", err.Error())
		return
	}
	defer rows.Close()

	var animals []models.Livestock
	for rows.Next() {
		var animal models.Livestock
		var birthDate sql.NullTime
		var weight sql.NullFloat64

		err := rows.Scan(
			&animal.ID, &animal.UserID, &animal.TagNumber, &animal.Type, &animal.Breed,
			&animal.Gender, &birthDate, &weight, &animal.HealthStatus, &animal.Location,
			&animal.Mother, &animal.Father, &animal.Notes, &animal.CreatedAt, &animal.UpdatedAt,
		)
		if err != nil {
			continue
		}

		animal.BirthDate = utils.NullTimeToPtr(birthDate)
		animal.Weight = utils.NullFloat64ToPtr(weight)

		animals = append(animals, animal)
	}

	response := map[string]interface{}{
		"animals":    animals,
		"pagination": pagination,
	}

	utils.SuccessResponse(c, response, "Hayvanlar başarıyla getirildi")
}

// CreateLivestock yeni hayvan oluşturma
// @Summary Yeni hayvan oluşturma
// @Description Yeni hayvan kaydı oluşturur
// @Tags Livestock
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.Livestock true "Hayvan bilgileri"
// @Success 201 {object} models.APIResponse{data=models.Livestock}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /livestock [post]
func (h *LivestockHandler) CreateLivestock(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	var req models.Livestock
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	// Gerekli alanları kontrol et
	if utils.IsEmptyString(req.TagNumber) || utils.IsEmptyString(req.Type) || utils.IsEmptyString(req.Breed) {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_FIELDS", "Gerekli alanlar eksik", nil)
		return
	}

	// Tag number benzersiz mi kontrol et
	var exists bool
	err = h.db.QueryRow("SELECT 1 FROM livestock WHERE tag_number = ? AND user_id = ?", req.TagNumber, userID).Scan(&exists)
	if err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "TAG_EXISTS", "Bu etiket numarası zaten kullanımda", nil)
		return
	}

	animalID := utils.GenerateID()

	// Hayvanı oluştur
	_, err = h.db.Exec(`
		INSERT INTO livestock (id, user_id, tag_number, type, breed, gender, birth_date,
		                      weight, health_status, location, mother, father, notes,
		                      created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, animalID, userID, req.TagNumber, req.Type, req.Breed, req.Gender, req.BirthDate,
		req.Weight, req.HealthStatus, req.Location, req.Mother, req.Father, req.Notes)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Hayvan oluşturulamadı", err.Error())
		return
	}

	// Oluşturulan hayvanı getir
	var animal models.Livestock
	var birthDate sql.NullTime
	var weight sql.NullFloat64

	err = h.db.QueryRow(`
		SELECT id, user_id, tag_number, type, breed, gender, birth_date, weight,
		       health_status, location, mother, father, notes, created_at, updated_at
		FROM livestock WHERE id = ?
	`, animalID).Scan(
		&animal.ID, &animal.UserID, &animal.TagNumber, &animal.Type, &animal.Breed,
		&animal.Gender, &birthDate, &weight, &animal.HealthStatus, &animal.Location,
		&animal.Mother, &animal.Father, &animal.Notes, &animal.CreatedAt, &animal.UpdatedAt,
	)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "FETCH_ERROR", "Oluşturulan hayvan getirilemedi", err.Error())
		return
	}

	animal.BirthDate = utils.NullTimeToPtr(birthDate)
	animal.Weight = utils.NullFloat64ToPtr(weight)

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    animal,
		Message: "Hayvan başarıyla oluşturuldu",
	})
}

// GetLivestock hayvan detayları
// @Summary Hayvan detayları
// @Description Belirli bir hayvanın detaylarını getirir
// @Tags Livestock
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Hayvan ID"
// @Success 200 {object} models.APIResponse{data=models.Livestock}
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /livestock/{id} [get]
func (h *LivestockHandler) GetLivestockByID(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	animalID := c.Param("id")
	if utils.IsEmptyString(animalID) {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_ID", "Hayvan ID gerekli", nil)
		return
	}

	var animal models.Livestock
	var birthDate sql.NullTime
	var weight sql.NullFloat64

	err = h.db.QueryRow(`
		SELECT id, user_id, tag_number, type, breed, gender, birth_date, weight,
		       health_status, location, mother, father, notes, created_at, updated_at
		FROM livestock WHERE id = ? AND user_id = ?
	`, animalID, userID).Scan(
		&animal.ID, &animal.UserID, &animal.TagNumber, &animal.Type, &animal.Breed,
		&animal.Gender, &birthDate, &weight, &animal.HealthStatus, &animal.Location,
		&animal.Mother, &animal.Father, &animal.Notes, &animal.CreatedAt, &animal.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.ErrorResponse(c, http.StatusNotFound, "ANIMAL_NOT_FOUND", "Hayvan bulunamadı", nil)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Hayvan getirilemedi", err.Error())
		}
		return
	}

	animal.BirthDate = utils.NullTimeToPtr(birthDate)
	animal.Weight = utils.NullFloat64ToPtr(weight)

	utils.SuccessResponse(c, animal, "Hayvan detayları başarıyla getirildi")
}

// UpdateLivestock hayvan güncelleme
// @Summary Hayvan güncelleme
// @Description Mevcut hayvan bilgilerini günceller
// @Tags Livestock
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Hayvan ID"
// @Param request body models.Livestock true "Güncellenecek hayvan bilgileri"
// @Success 200 {object} models.APIResponse{data=models.Livestock}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /livestock/{id} [put]
func (h *LivestockHandler) UpdateLivestock(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	animalID := c.Param("id")
	if utils.IsEmptyString(animalID) {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_ID", "Hayvan ID gerekli", nil)
		return
	}

	var req models.Livestock
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	// Hayvanı güncelle
	_, err = h.db.Exec(`
		UPDATE livestock 
		SET tag_number = ?, type = ?, breed = ?, gender = ?, birth_date = ?, weight = ?,
		    health_status = ?, location = ?, mother = ?, father = ?, notes = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ?
	`, req.TagNumber, req.Type, req.Breed, req.Gender, req.BirthDate, req.Weight,
		req.HealthStatus, req.Location, req.Mother, req.Father, req.Notes, animalID, userID)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_ERROR", "Hayvan güncellenemedi", err.Error())
		return
	}

	// Güncellenmiş hayvanı getir
	h.GetLivestockByID(c)
}

// DeleteLivestock hayvan silme
// @Summary Hayvan silme
// @Description Belirli bir hayvanı siler
// @Tags Livestock
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Hayvan ID"
// @Success 200 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /livestock/{id} [delete]
func (h *LivestockHandler) DeleteLivestock(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	animalID := c.Param("id")
	if utils.IsEmptyString(animalID) {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_ID", "Hayvan ID gerekli", nil)
		return
	}

	// Hayvanı sil
	result, err := h.db.Exec("DELETE FROM livestock WHERE id = ? AND user_id = ?", animalID, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DELETE_ERROR", "Hayvan silinemedi", err.Error())
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "ANIMAL_NOT_FOUND", "Hayvan bulunamadı", nil)
		return
	}

	utils.SuccessResponse(c, nil, "Hayvan başarıyla silindi")
}

// GetLivestockStatistics hayvancılık istatistikleri
// @Summary Hayvancılık istatistikleri
// @Description Hayvancılık istatistiklerini getirir
// @Tags Livestock
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=map[string]interface{}}
// @Failure 401 {object} models.APIResponse
// @Router /livestock/statistics [get]
func (h *LivestockHandler) GetLivestockStatistics(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	// Toplam hayvan sayısı
	var totalAnimals int
	err = h.db.QueryRow("SELECT COUNT(*) FROM livestock WHERE user_id = ?", userID).Scan(&totalAnimals)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Toplam hayvan sayısı alınamadı", err.Error())
		return
	}

	// Tür bazında hayvan sayıları
	var cattle, sheep, goat, chicken int
	h.db.QueryRow("SELECT COUNT(*) FROM livestock WHERE user_id = ? AND type = 'cattle'", userID).Scan(&cattle)
	h.db.QueryRow("SELECT COUNT(*) FROM livestock WHERE user_id = ? AND type = 'sheep'", userID).Scan(&sheep)
	h.db.QueryRow("SELECT COUNT(*) FROM livestock WHERE user_id = ? AND type = 'goat'", userID).Scan(&goat)
	h.db.QueryRow("SELECT COUNT(*) FROM livestock WHERE user_id = ? AND type = 'chicken'", userID).Scan(&chicken)

	// Sağlık durumu istatistikleri
	var healthy, sick, pregnant, vaccinationNeeded int
	h.db.QueryRow("SELECT COUNT(*) FROM livestock WHERE user_id = ? AND health_status = 'healthy'", userID).Scan(&healthy)
	h.db.QueryRow("SELECT COUNT(*) FROM livestock WHERE user_id = ? AND health_status = 'sick'", userID).Scan(&sick)
	h.db.QueryRow("SELECT COUNT(*) FROM livestock WHERE user_id = ? AND health_status = 'pregnant'", userID).Scan(&pregnant)
	h.db.QueryRow("SELECT COUNT(*) FROM livestock WHERE user_id = ? AND health_status = 'vaccination_needed'", userID).Scan(&vaccinationNeeded)

	// Günlük süt üretimi (basit hesaplama)
	var dailyMilkProduction float64
	err = h.db.QueryRow(`
		SELECT COALESCE(SUM(amount), 0)
		FROM milk_production 
		WHERE user_id = ? AND DATE(date) = CURDATE()
	`, userID).Scan(&dailyMilkProduction)

	// Aşılama oranı
	var vaccinationRate float64
	if totalAnimals > 0 {
		vaccinationRate = float64(healthy) / float64(totalAnimals) * 100
	}

	statistics := map[string]interface{}{
		"totalAnimals": totalAnimals,
		"animalsByType": map[string]int{
			"cattle":  cattle,
			"sheep":   sheep,
			"goat":    goat,
			"chicken": chicken,
		},
		"healthStatistics": map[string]int{
			"healthy":            healthy,
			"sick":               sick,
			"pregnant":           pregnant,
			"vaccination_needed": vaccinationNeeded,
		},
		"dailyMilkProduction": dailyMilkProduction,
		"vaccinationRate":     vaccinationRate,
	}

	utils.SuccessResponse(c, statistics, "Hayvancılık istatistikleri başarıyla getirildi")
}

// GetLivestockCategories hayvan kategorileri
// @Summary Hayvan kategorileri
// @Description Hayvan kategorilerini getirir
// @Tags Livestock
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=[]models.CategoryData}
// @Failure 401 {object} models.APIResponse
// @Router /livestock/categories [get]
func (h *LivestockHandler) GetLivestockCategories(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	// Kategori verilerini getir
	rows, err := h.db.Query(`
		SELECT type, COUNT(*) as count
		FROM livestock WHERE user_id = ?
		GROUP BY type
	`, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Kategori verileri alınamadı", err.Error())
		return
	}
	defer rows.Close()

	var categories []models.CategoryData
	for rows.Next() {
		var category models.CategoryData
		err := rows.Scan(&category.Name, &category.Count)
		if err != nil {
			continue
		}

		// Her kategori için icon ve renk ata
		switch category.Name {
		case "cattle":
			category.Icon = "🐄"
			category.Color = "#4CAF50"
		case "sheep":
			category.Icon = "🐑"
			category.Color = "#2196F3"
		case "goat":
			category.Icon = "🐐"
			category.Color = "#FF9800"
		case "chicken":
			category.Icon = "🐔"
			category.Color = "#9C27B0"
		default:
			category.Icon = "🐾"
			category.Color = "#607D8B"
		}

		categories = append(categories, category)
	}

	utils.SuccessResponse(c, categories, "Hayvan kategorileri başarıyla getirildi")
}

// GetHealthRecords sağlık kayıtları
// @Summary Sağlık kayıtları
// @Description Belirli bir hayvanın sağlık kayıtlarını listeler
// @Tags Livestock
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Hayvan ID"
// @Success 200 {object} models.APIResponse{data=[]models.HealthRecord}
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /livestock/{id}/health-records [get]
func (h *LivestockHandler) GetHealthRecords(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	animalID := c.Param("id")
	if utils.IsEmptyString(animalID) {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_ID", "Hayvan ID gerekli", nil)
		return
	}

	// Hayvan kullanıcıya ait mi kontrol et
	var exists bool
	err = h.db.QueryRow("SELECT 1 FROM livestock WHERE id = ? AND user_id = ?", animalID, userID).Scan(&exists)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "ANIMAL_NOT_FOUND", "Hayvan bulunamadı", nil)
		return
	}

	// Sağlık kayıtlarını getir
	rows, err := h.db.Query(`
		SELECT id, animal_id, type, description, date, veterinarian, cost, notes, next_checkup
		FROM health_records WHERE animal_id = ?
		ORDER BY date DESC
	`, animalID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Sağlık kayıtları alınamadı", err.Error())
		return
	}
	defer rows.Close()

	var records []models.HealthRecord
	for rows.Next() {
		var record models.HealthRecord
		var date, nextCheckup sql.NullTime
		var cost sql.NullFloat64

		err := rows.Scan(
			&record.ID, &record.AnimalID, &record.Type, &record.Description,
			&date, &record.Veterinarian, &cost, &record.Notes, &nextCheckup,
		)
		if err != nil {
			continue
		}

		record.Date = utils.NullTimeToPtr(date)
		record.Cost = utils.NullFloat64ToPtr(cost)
		record.NextCheckup = utils.NullTimeToPtr(nextCheckup)

		records = append(records, record)
	}

	utils.SuccessResponse(c, records, "Sağlık kayıtları başarıyla getirildi")
}

// CreateHealthRecord sağlık kaydı oluşturma
// @Summary Sağlık kaydı oluşturma
// @Description Yeni sağlık kaydı oluşturur
// @Tags Livestock
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Hayvan ID"
// @Param request body models.HealthRecord true "Sağlık kaydı bilgileri"
// @Success 201 {object} models.APIResponse{data=models.HealthRecord}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /livestock/{id}/health-records [post]
func (h *LivestockHandler) CreateHealthRecord(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	animalID := c.Param("id")
	if utils.IsEmptyString(animalID) {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_ID", "Hayvan ID gerekli", nil)
		return
	}

	var req models.HealthRecord
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	// Hayvan kullanıcıya ait mi kontrol et
	var exists bool
	err = h.db.QueryRow("SELECT 1 FROM livestock WHERE id = ? AND user_id = ?", animalID, userID).Scan(&exists)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "ANIMAL_NOT_FOUND", "Hayvan bulunamadı", nil)
		return
	}

	// Sağlık kaydını oluştur
	recordID := utils.GenerateID()
	_, err = h.db.Exec(`
		INSERT INTO health_records (id, animal_id, type, description, date, veterinarian,
		                           cost, notes, next_checkup, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, recordID, animalID, req.Type, req.Description, req.Date, req.Veterinarian,
		req.Cost, req.Notes, req.NextCheckup)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Sağlık kaydı oluşturulamadı", err.Error())
		return
	}

	// Oluşturulan kaydı getir
	var record models.HealthRecord
	var date, nextCheckup sql.NullTime
	var cost sql.NullFloat64

	err = h.db.QueryRow(`
		SELECT id, animal_id, type, description, date, veterinarian, cost, notes, next_checkup, created_at
		FROM health_records WHERE id = ?
	`, recordID).Scan(
		&record.ID, &record.AnimalID, &record.Type, &record.Description,
		&date, &record.Veterinarian, &cost, &record.Notes, &nextCheckup, &record.CreatedAt,
	)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "FETCH_ERROR", "Oluşturulan kayıt getirilemedi", err.Error())
		return
	}

	record.Date = utils.NullTimeToPtr(date)
	record.Cost = utils.NullFloat64ToPtr(cost)
	record.NextCheckup = utils.NullTimeToPtr(nextCheckup)

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    record,
		Message: "Sağlık kaydı başarıyla oluşturuldu",
	})
}

// GetMilkProduction süt üretim kayıtları
// @Summary Süt üretim kayıtları
// @Description Süt üretim kayıtlarını getirir
// @Tags Livestock
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param startDate query string false "Başlangıç tarihi"
// @Param endDate query string false "Bitiş tarihi"
// @Param animalId query string false "Hayvan ID"
// @Success 200 {object} models.APIResponse{data=[]models.MilkProduction}
// @Failure 401 {object} models.APIResponse
// @Router /livestock/milk-production [get]
func (h *LivestockHandler) GetMilkProduction(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	startDate := c.DefaultQuery("startDate", "")
	endDate := c.DefaultQuery("endDate", "")
	animalID := c.DefaultQuery("animalId", "")

	// Sorgu oluştur
	whereClause := "WHERE user_id = ?"
	args := []interface{}{userID}

	if animalID != "" {
		whereClause += " AND animal_id = ?"
		args = append(args, animalID)
	}

	if startDate != "" {
		whereClause += " AND date >= ?"
		args = append(args, startDate)
	}

	if endDate != "" {
		whereClause += " AND date <= ?"
		args = append(args, endDate)
	}

	// Süt üretim kayıtlarını getir
	rows, err := h.db.Query(`
		SELECT id, animal_id, date, amount, quality, notes, created_at
		FROM milk_production `+whereClause+`
		ORDER BY date DESC
	`, args...)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Süt üretim kayıtları alınamadı", err.Error())
		return
	}
	defer rows.Close()

	var productions []models.MilkProductionRecord
	for rows.Next() {
		var production models.MilkProductionRecord
		var date sql.NullTime

		err := rows.Scan(
			&production.ID, &production.AnimalID, &date, &production.Amount,
			&production.Quality, &production.Notes, &production.CreatedAt,
		)
		if err != nil {
			continue
		}

		production.Date = utils.NullTimeToPtr(date)
		productions = append(productions, production)
	}

	utils.SuccessResponse(c, productions, "Süt üretim kayıtları başarıyla getirildi")
}

// CreateMilkProduction süt üretim kaydı oluşturma
// @Summary Süt üretim kaydı oluşturma
// @Description Yeni süt üretim kaydı oluşturur
// @Tags Livestock
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.MilkProduction true "Süt üretim bilgileri"
// @Success 201 {object} models.APIResponse{data=models.MilkProduction}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /livestock/milk-production [post]
func (h *LivestockHandler) CreateMilkProduction(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	var req models.MilkProductionRecord
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	// Gerekli alanları kontrol et
	if req.AnimalID == "" || req.Amount <= 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_FIELDS", "Gerekli alanlar eksik", nil)
		return
	}

	// Hayvan kullanıcıya ait mi kontrol et
	var exists bool
	err = h.db.QueryRow("SELECT 1 FROM livestock WHERE id = ? AND user_id = ?", req.AnimalID, userID).Scan(&exists)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "ANIMAL_NOT_FOUND", "Hayvan bulunamadı", nil)
		return
	}

	// Süt üretim kaydını oluştur
	productionID := utils.GenerateID()
	_, err = h.db.Exec(`
		INSERT INTO milk_production (id, user_id, animal_id, date, amount, quality, notes, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, productionID, userID, req.AnimalID, req.Date, req.Amount, req.Quality, req.Notes)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Süt üretim kaydı oluşturulamadı", err.Error())
		return
	}

	// Oluşturulan kaydı getir
	var production models.MilkProductionRecord
	var date sql.NullTime

	err = h.db.QueryRow(`
		SELECT id, animal_id, date, amount, quality, notes, created_at
		FROM milk_production WHERE id = ?
	`, productionID).Scan(
		&production.ID, &production.AnimalID, &date, &production.Amount,
		&production.Quality, &production.Notes, &production.CreatedAt,
	)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "FETCH_ERROR", "Oluşturulan kayıt getirilemedi", err.Error())
		return
	}

	production.Date = utils.NullTimeToPtr(date)

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    production,
		Message: "Süt üretim kaydı başarıyla oluşturuldu",
	})
}
