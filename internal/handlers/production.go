package handlers

import (
	"database/sql"
	"net/http"

	"agri-management-api/internal/models"
	"agri-management-api/internal/utils"

	"github.com/gin-gonic/gin"
)

// ProductionHandler üretim işlemlerini yönetir
type ProductionHandler struct {
	db *sql.DB
}

// NewProductionHandler yeni production handler oluşturur
func NewProductionHandler(db *sql.DB) *ProductionHandler {
	return &ProductionHandler{db: db}
}

// GetProductions üretim listesi
// @Summary Üretim listesi
// @Description Kullanıcının üretimlerini listeler
// @Tags Production
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Sayfa numarası"
// @Param limit query int false "Sayfa başına kayıt"
// @Param category query string false "Ürün kategorisi"
// @Param status query string false "Üretim durumu"
// @Success 200 {object} models.APIResponse{data=map[string]interface{}}
// @Failure 401 {object} models.APIResponse
// @Router /production [get]
func (h *ProductionHandler) GetProductions(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	page, limit := utils.ParsePagination(c)
	category := c.DefaultQuery("category", "all")
	status := c.DefaultQuery("status", "all")

	// Toplam kayıt sayısını al
	var total int
	whereClause := "WHERE user_id = ?"
	args := []interface{}{userID}

	if category != "all" {
		whereClause += " AND category = ?"
		args = append(args, category)
	}

	if status != "all" {
		whereClause += " AND status = ?"
		args = append(args, status)
	}

	err = h.db.QueryRow("SELECT COUNT(*) FROM production "+whereClause, args...).Scan(&total)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Toplam kayıt sayısı alınamadı", err.Error())
		return
	}

	// Sayfalama hesapla
	pagination := utils.CalculatePagination(page, limit, total)

	// Üretimleri getir
	offset := (page - 1) * limit
	query := `
		SELECT id, user_id, land_id, name, category, amount, unit, harvest_date,
		       quality, storage_location, status, price, notes, created_at, updated_at
		FROM production ` + whereClause + `
		ORDER BY created_at DESC LIMIT ? OFFSET ?
	`
	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Üretimler alınamadı", err.Error())
		return
	}
	defer rows.Close()

	var productions []models.Production
	for rows.Next() {
		var production models.Production
		var harvestDate sql.NullTime
		var price sql.NullFloat64

		err := rows.Scan(
			&production.ID, &production.UserID, &production.LandID, &production.Name,
			&production.Category, &production.Amount, &production.Unit, &harvestDate,
			&production.Quality, &production.StorageLocation, &production.Status,
			&price, &production.Notes, &production.CreatedAt, &production.UpdatedAt,
		)
		if err != nil {
			continue
		}

		production.HarvestDate = utils.NullTimeToPtr(harvestDate)
		production.Price = utils.NullFloat64ToPtr(price)

		productions = append(productions, production)
	}

	response := map[string]interface{}{
		"productions": productions,
		"pagination":  pagination,
	}

	utils.SuccessResponse(c, response, "Üretimler başarıyla getirildi")
}

// CreateProduction yeni üretim oluşturma
// @Summary Yeni üretim oluşturma
// @Description Yeni üretim kaydı oluşturur
// @Tags Production
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.Production true "Üretim bilgileri"
// @Success 201 {object} models.APIResponse{data=models.Production}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /production [post]
func (h *ProductionHandler) CreateProduction(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	var req models.Production
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	// Gerekli alanları kontrol et
	if utils.IsEmptyString(req.Name) || utils.IsEmptyString(req.Category) || req.Amount <= 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_FIELDS", "Gerekli alanlar eksik", nil)
		return
	}

	productionID := utils.GenerateID()

	// Üretimi oluştur
	_, err = h.db.Exec(`
		INSERT INTO production (id, user_id, land_id, name, category, amount, unit, harvest_date,
		                       quality, storage_location, status, price, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'active', ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, productionID, userID, req.LandID, req.Name, req.Category, req.Amount, req.Unit,
		req.HarvestDate, req.Quality, req.StorageLocation, req.Price, req.Notes)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Üretim oluşturulamadı", err.Error())
		return
	}

	// Oluşturulan üretimi getir
	var production models.Production
	var harvestDate sql.NullTime
	var price sql.NullFloat64

	err = h.db.QueryRow(`
		SELECT id, user_id, land_id, name, category, amount, unit, harvest_date,
		       quality, storage_location, status, price, notes, created_at, updated_at
		FROM production WHERE id = ?
	`, productionID).Scan(
		&production.ID, &production.UserID, &production.LandID, &production.Name,
		&production.Category, &production.Amount, &production.Unit, &harvestDate,
		&production.Quality, &production.StorageLocation, &production.Status,
		&price, &production.Notes, &production.CreatedAt, &production.UpdatedAt,
	)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "FETCH_ERROR", "Oluşturulan üretim getirilemedi", err.Error())
		return
	}

	production.HarvestDate = utils.NullTimeToPtr(harvestDate)
	production.Price = utils.NullFloat64ToPtr(price)

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    production,
		Message: "Üretim başarıyla oluşturuldu",
	})
}

// GetProduction üretim detayları
// @Summary Üretim detayları
// @Description Belirli bir üretimin detaylarını getirir
// @Tags Production
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Üretim ID"
// @Success 200 {object} models.APIResponse{data=models.Production}
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /production/{id} [get]
func (h *ProductionHandler) GetProduction(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	productionID := c.Param("id")
	if utils.IsEmptyString(productionID) {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_ID", "Üretim ID gerekli", nil)
		return
	}

	var production models.Production
	var harvestDate sql.NullTime
	var price sql.NullFloat64

	err = h.db.QueryRow(`
		SELECT id, user_id, land_id, name, category, amount, unit, harvest_date,
		       quality, storage_location, status, price, notes, created_at, updated_at
		FROM production WHERE id = ? AND user_id = ?
	`, productionID, userID).Scan(
		&production.ID, &production.UserID, &production.LandID, &production.Name,
		&production.Category, &production.Amount, &production.Unit, &harvestDate,
		&production.Quality, &production.StorageLocation, &production.Status,
		&price, &production.Notes, &production.CreatedAt, &production.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.ErrorResponse(c, http.StatusNotFound, "PRODUCTION_NOT_FOUND", "Üretim bulunamadı", nil)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Üretim getirilemedi", err.Error())
		}
		return
	}

	production.HarvestDate = utils.NullTimeToPtr(harvestDate)
	production.Price = utils.NullFloat64ToPtr(price)

	utils.SuccessResponse(c, production, "Üretim detayları başarıyla getirildi")
}

// UpdateProduction üretim güncelleme
// @Summary Üretim güncelleme
// @Description Mevcut üretim bilgilerini günceller
// @Tags Production
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Üretim ID"
// @Param request body models.Production true "Güncellenecek üretim bilgileri"
// @Success 200 {object} models.APIResponse{data=models.Production}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /production/{id} [put]
func (h *ProductionHandler) UpdateProduction(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	productionID := c.Param("id")
	if utils.IsEmptyString(productionID) {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_ID", "Üretim ID gerekli", nil)
		return
	}

	var req models.Production
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	// Üretimi güncelle
	_, err = h.db.Exec(`
		UPDATE production 
		SET name = ?, category = ?, amount = ?, unit = ?, harvest_date = ?, quality = ?,
		    storage_location = ?, status = ?, price = ?, notes = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ?
	`, req.Name, req.Category, req.Amount, req.Unit, req.HarvestDate, req.Quality,
		req.StorageLocation, req.Status, req.Price, req.Notes, productionID, userID)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_ERROR", "Üretim güncellenemedi", err.Error())
		return
	}

	// Güncellenmiş üretimi getir
	h.GetProduction(c)
}

// DeleteProduction üretim silme
// @Summary Üretim silme
// @Description Belirli bir üretimi siler
// @Tags Production
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Üretim ID"
// @Success 200 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /production/{id} [delete]
func (h *ProductionHandler) DeleteProduction(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	productionID := c.Param("id")
	if utils.IsEmptyString(productionID) {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_ID", "Üretim ID gerekli", nil)
		return
	}

	// Üretimi sil
	result, err := h.db.Exec("DELETE FROM production WHERE id = ? AND user_id = ?", productionID, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DELETE_ERROR", "Üretim silinemedi", err.Error())
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "PRODUCTION_NOT_FOUND", "Üretim bulunamadı", nil)
		return
	}

	utils.SuccessResponse(c, nil, "Üretim başarıyla silindi")
}

// GetProductionStatistics üretim istatistikleri
// @Summary Üretim istatistikleri
// @Description Üretim istatistiklerini getirir
// @Tags Production
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=map[string]interface{}}
// @Failure 401 {object} models.APIResponse
// @Router /production/statistics [get]
func (h *ProductionHandler) GetProductionStatistics(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	// Aktif ürün sayısı
	var activeProducts int
	err = h.db.QueryRow("SELECT COUNT(*) FROM production WHERE user_id = ? AND status = 'active'", userID).Scan(&activeProducts)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Aktif ürün sayısı alınamadı", err.Error())
		return
	}

	// Toplam üretim
	var totalProduction float64
	err = h.db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM production WHERE user_id = ?", userID).Scan(&totalProduction)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Toplam üretim alınamadı", err.Error())
		return
	}

	// Ortalama verimlilik
	var averageProductivity float64
	err = h.db.QueryRow("SELECT COALESCE(AVG(amount), 0) FROM production WHERE user_id = ?", userID).Scan(&averageProductivity)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Ortalama verimlilik alınamadı", err.Error())
		return
	}

	// Kalite dağılımı
	var aPlus, a, b, cQuality int
	h.db.QueryRow("SELECT COUNT(*) FROM production WHERE user_id = ? AND quality = 'A+'", userID).Scan(&aPlus)
	h.db.QueryRow("SELECT COUNT(*) FROM production WHERE user_id = ? AND quality = 'A'", userID).Scan(&a)
	h.db.QueryRow("SELECT COUNT(*) FROM production WHERE user_id = ? AND quality = 'B'", userID).Scan(&b)
	h.db.QueryRow("SELECT COUNT(*) FROM production WHERE user_id = ? AND quality = 'C'", userID).Scan(&cQuality)

	// Kategori bazında dağılım
	rows, err := h.db.Query(`
		SELECT category, COUNT(*) as count, COALESCE(SUM(amount), 0) as amount
		FROM production WHERE user_id = ?
		GROUP BY category
	`, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Kategori dağılımı alınamadı", err.Error())
		return
	}
	defer rows.Close()

	var categoryBreakdown []map[string]interface{}
	totalCount := 0
	for rows.Next() {
		var category string
		var count int
		var amount float64

		err := rows.Scan(&category, &count, &amount)
		if err != nil {
			continue
		}

		totalCount += count
		categoryBreakdown = append(categoryBreakdown, map[string]interface{}{
			"name":   category,
			"count":  count,
			"amount": amount,
		})
	}

	// Yüzdeleri hesapla
	for i := range categoryBreakdown {
		if totalCount > 0 {
			count := categoryBreakdown[i]["count"].(int)
			categoryBreakdown[i]["percentage"] = float64(count) / float64(totalCount) * 100
		} else {
			categoryBreakdown[i]["percentage"] = 0
		}
	}

	statistics := map[string]interface{}{
		"activeProducts":      activeProducts,
		"totalProduction":     totalProduction,
		"averageProductivity": averageProductivity,
		"qualityDistribution": map[string]int{
			"A+": aPlus,
			"A":  a,
			"B":  b,
			"C":  cQuality,
		},
		"categoryBreakdown": categoryBreakdown,
	}

	utils.SuccessResponse(c, statistics, "Üretim istatistikleri başarıyla getirildi")
}

// GetProductionCategories üretim kategorileri
// @Summary Üretim kategorileri
// @Description Üretim kategorilerini getirir
// @Tags Production
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=[]models.CategoryData}
// @Failure 401 {object} models.APIResponse
// @Router /production/categories [get]
func (h *ProductionHandler) GetProductionCategories(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	// Kategori verilerini getir
	rows, err := h.db.Query(`
		SELECT category, COUNT(*) as count
		FROM production WHERE user_id = ?
		GROUP BY category
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
		case "vegetables":
			category.Icon = "🥬"
			category.Color = "#4CAF50"
		case "fruits":
			category.Icon = "🍎"
			category.Color = "#FF5722"
		case "grains":
			category.Icon = "🌾"
			category.Color = "#FF9800"
		case "dairy":
			category.Icon = "🥛"
			category.Color = "#2196F3"
		case "meat":
			category.Icon = "🥩"
			category.Color = "#795548"
		default:
			category.Icon = "🌱"
			category.Color = "#607D8B"
		}

		categories = append(categories, category)
	}

	utils.SuccessResponse(c, categories, "Üretim kategorileri başarıyla getirildi")
}
