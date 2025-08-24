package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"agri-management-api/internal/models"
	"agri-management-api/internal/utils"

	"github.com/gin-gonic/gin"
)

// FinanceHandler finans işlemlerini yönetir
type FinanceHandler struct {
	db *sql.DB
}

// NewFinanceHandler yeni finance handler oluşturur
func NewFinanceHandler(db *sql.DB) *FinanceHandler {
	return &FinanceHandler{db: db}
}

// GetFinanceSummary finansal özet
// @Summary Finansal özet
// @Description Finansal özet verileri getirir
// @Tags Finance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param period query string false "Periyot"
// @Success 200 {object} models.APIResponse{data=map[string]interface{}}
// @Failure 401 {object} models.APIResponse
// @Router /finance/summary [get]
func (h *FinanceHandler) GetFinanceSummary(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	period := c.DefaultQuery("period", "month")

	// Periyoda göre tarih aralığı belirle
	var startDate, endDate string
	now := time.Now()

	switch period {
	case "month":
		startDate = now.Format("2006-01") + "-01"
		endDate = now.Format("2006-01-02")
	case "quarter":
		quarter := (now.Month()-1)/3 + 1
		startDate = now.Format("2006") + "-" + time.Month((quarter-1)*3 + 1).String()[:2] + "-01"
		endDate = now.Format("2006-01-02")
	case "year":
		startDate = now.Format("2006") + "-01-01"
		endDate = now.Format("2006-01-02")
	default:
		startDate = now.Format("2006-01") + "-01"
		endDate = now.Format("2006-01-02")
	}

	// Toplam gelir
	var totalIncome float64
	err = h.db.QueryRow(`
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions 
		WHERE user_id = ? AND type = 'income' AND date >= ? AND date <= ?
	`, userID, startDate, endDate).Scan(&totalIncome)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Toplam gelir alınamadı", err.Error())
		return
	}

	// Toplam gider
	var totalExpense float64
	err = h.db.QueryRow(`
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions 
		WHERE user_id = ? AND type = 'expense' AND date >= ? AND date <= ?
	`, userID, startDate, endDate).Scan(&totalExpense)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Toplam gider alınamadı", err.Error())
		return
	}

	// Net kar
	netProfit := totalIncome - totalExpense

	// Bekleyen ödemeler
	var pendingPayments float64
	err = h.db.QueryRow(`
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions 
		WHERE user_id = ? AND status = 'pending'
	`, userID).Scan(&pendingPayments)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Bekleyen ödemeler alınamadı", err.Error())
		return
	}

	// Trend hesaplamaları (basit implementasyon)
	summary := map[string]interface{}{
		"totalIncome":     totalIncome,
		"totalExpense":    totalExpense,
		"netProfit":       netProfit,
		"pendingPayments": pendingPayments,
		"trends": map[string]float64{
			"income":  5.2,  // Mock data
			"expense": -3.1, // Mock data
			"profit":  8.5,  // Mock data
		},
	}

	utils.SuccessResponse(c, summary, "Finansal özet başarıyla getirildi")
}

// GetTransactions işlem listesi
// @Summary İşlem listesi
// @Description Finansal işlemlerin listesini getirir
// @Tags Finance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Sayfa numarası"
// @Param limit query int false "Sayfa başına kayıt"
// @Param type query string false "İşlem türü"
// @Param category query string false "Kategori"
// @Param startDate query string false "Başlangıç tarihi"
// @Param endDate query string false "Bitiş tarihi"
// @Success 200 {object} models.APIResponse{data=map[string]interface{}}
// @Failure 401 {object} models.APIResponse
// @Router /finance/transactions [get]
func (h *FinanceHandler) GetTransactions(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	page, limit := utils.ParsePagination(c)
	transactionType := c.DefaultQuery("type", "all")
	category := c.DefaultQuery("category", "all")
	startDate := c.DefaultQuery("startDate", "")
	endDate := c.DefaultQuery("endDate", "")

	// Sorgu oluştur
	whereClause := "WHERE user_id = ?"
	args := []interface{}{userID}

	if transactionType != "all" {
		whereClause += " AND type = ?"
		args = append(args, transactionType)
	}

	if category != "all" {
		whereClause += " AND category = ?"
		args = append(args, category)
	}

	if startDate != "" {
		whereClause += " AND date >= ?"
		args = append(args, startDate)
	}

	if endDate != "" {
		whereClause += " AND date <= ?"
		args = append(args, endDate)
	}

	// Toplam kayıt sayısını al
	var total int
	err = h.db.QueryRow("SELECT COUNT(*) FROM transactions "+whereClause, args...).Scan(&total)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Toplam kayıt sayısı alınamadı", err.Error())
		return
	}

	// Sayfalama hesapla
	pagination := utils.CalculatePagination(page, limit, total)

	// İşlemleri getir
	offset := (page - 1) * limit
	query := `
		SELECT id, user_id, type, category, description, amount, currency, date,
		       status, payment_method, receipt, notes, created_at, updated_at
		FROM transactions ` + whereClause + `
		ORDER BY date DESC LIMIT ? OFFSET ?
	`
	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "İşlemler alınamadı", err.Error())
		return
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var transaction models.Transaction

		err := rows.Scan(
			&transaction.ID, &transaction.UserID, &transaction.Type, &transaction.Category,
			&transaction.Description, &transaction.Amount, &transaction.Currency, &transaction.Date,
			&transaction.Status, &transaction.PaymentMethod, &transaction.Receipt, &transaction.Notes,
			&transaction.CreatedAt, &transaction.UpdatedAt,
		)
		if err != nil {
			continue
		}

		transactions = append(transactions, transaction)
	}

	response := map[string]interface{}{
		"transactions": transactions,
		"pagination":   pagination,
	}

	utils.SuccessResponse(c, response, "İşlemler başarıyla getirildi")
}

// CreateTransaction yeni işlem ekleme
// @Summary Yeni işlem ekleme
// @Description Yeni finansal işlem oluşturur
// @Tags Finance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.Transaction true "İşlem bilgileri"
// @Success 201 {object} models.APIResponse{data=models.Transaction}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /finance/transactions [post]
func (h *FinanceHandler) CreateTransaction(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	var req models.Transaction
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	// Gerekli alanları kontrol et
	if utils.IsEmptyString(req.Type) || utils.IsEmptyString(req.Category) || req.Amount <= 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_FIELDS", "Gerekli alanlar eksik", nil)
		return
	}

	transactionID := utils.GenerateID()

	// İşlemi oluştur
	_, err = h.db.Exec(`
		INSERT INTO transactions (id, user_id, type, category, description, amount, currency,
		                         date, status, payment_method, receipt, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'completed', ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, transactionID, userID, req.Type, req.Category, req.Description, req.Amount, req.Currency,
		req.Date, req.PaymentMethod, req.Receipt, req.Notes)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "İşlem oluşturulamadı", err.Error())
		return
	}

	// Oluşturulan işlemi getir
	var transaction models.Transaction
	err = h.db.QueryRow(`
		SELECT id, user_id, type, category, description, amount, currency, date,
		       status, payment_method, receipt, notes, created_at, updated_at
		FROM transactions WHERE id = ?
	`, transactionID).Scan(
		&transaction.ID, &transaction.UserID, &transaction.Type, &transaction.Category,
		&transaction.Description, &transaction.Amount, &transaction.Currency, &transaction.Date,
		&transaction.Status, &transaction.PaymentMethod, &transaction.Receipt, &transaction.Notes,
		&transaction.CreatedAt, &transaction.UpdatedAt,
	)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "FETCH_ERROR", "Oluşturulan işlem getirilemedi", err.Error())
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    transaction,
		Message: "İşlem başarıyla oluşturuldu",
	})
}

// GetTransaction işlem detayları
// @Summary İşlem detayları
// @Description Belirli bir işlemin detaylarını getirir
// @Tags Finance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "İşlem ID"
// @Success 200 {object} models.APIResponse{data=models.Transaction}
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /finance/transactions/{id} [get]
func (h *FinanceHandler) GetTransaction(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	transactionID := c.Param("id")
	if utils.IsEmptyString(transactionID) {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_ID", "İşlem ID gerekli", nil)
		return
	}

	var transaction models.Transaction
	err = h.db.QueryRow(`
		SELECT id, user_id, type, category, description, amount, currency, date,
		       status, payment_method, receipt, notes, created_at, updated_at
		FROM transactions WHERE id = ? AND user_id = ?
	`, transactionID, userID).Scan(
		&transaction.ID, &transaction.UserID, &transaction.Type, &transaction.Category,
		&transaction.Description, &transaction.Amount, &transaction.Currency, &transaction.Date,
		&transaction.Status, &transaction.PaymentMethod, &transaction.Receipt, &transaction.Notes,
		&transaction.CreatedAt, &transaction.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.ErrorResponse(c, http.StatusNotFound, "TRANSACTION_NOT_FOUND", "İşlem bulunamadı", nil)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "İşlem getirilemedi", err.Error())
		}
		return
	}

	utils.SuccessResponse(c, transaction, "İşlem detayları başarıyla getirildi")
}

// UpdateTransaction işlem güncelleme
// @Summary İşlem güncelleme
// @Description Mevcut işlem bilgilerini günceller
// @Tags Finance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "İşlem ID"
// @Param request body models.Transaction true "Güncellenecek işlem bilgileri"
// @Success 200 {object} models.APIResponse{data=models.Transaction}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /finance/transactions/{id} [put]
func (h *FinanceHandler) UpdateTransaction(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	transactionID := c.Param("id")
	if utils.IsEmptyString(transactionID) {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_ID", "İşlem ID gerekli", nil)
		return
	}

	var req models.Transaction
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	// İşlemi güncelle
	_, err = h.db.Exec(`
		UPDATE transactions 
		SET type = ?, category = ?, description = ?, amount = ?, currency = ?, date = ?,
		    status = ?, payment_method = ?, receipt = ?, notes = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ?
	`, req.Type, req.Category, req.Description, req.Amount, req.Currency, req.Date,
		req.Status, req.PaymentMethod, req.Receipt, req.Notes, transactionID, userID)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_ERROR", "İşlem güncellenemedi", err.Error())
		return
	}

	// Güncellenmiş işlemi getir
	h.GetTransaction(c)
}

// DeleteTransaction işlem silme
// @Summary İşlem silme
// @Description Belirli bir işlemi siler
// @Tags Finance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "İşlem ID"
// @Success 200 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /finance/transactions/{id} [delete]
func (h *FinanceHandler) DeleteTransaction(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	transactionID := c.Param("id")
	if utils.IsEmptyString(transactionID) {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_ID", "İşlem ID gerekli", nil)
		return
	}

	// İşlemi sil
	result, err := h.db.Exec("DELETE FROM transactions WHERE id = ? AND user_id = ?", transactionID, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DELETE_ERROR", "İşlem silinemedi", err.Error())
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "TRANSACTION_NOT_FOUND", "İşlem bulunamadı", nil)
		return
	}

	utils.SuccessResponse(c, nil, "İşlem başarıyla silindi")
}

// GetCategories kategori listesi
// @Summary Kategori listesi
// @Description Finansal işlem kategorilerini getirir
// @Tags Finance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=map[string][]string}
// @Failure 401 {object} models.APIResponse
// @Router /finance/categories [get]
func (h *FinanceHandler) GetCategories(c *gin.Context) {
	_, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	// Gelir kategorileri
	incomeCategories := []string{
		"Ürün Satışı",
		"Hayvan Satışı",
		"Süt Satışı",
		"Hizmet Geliri",
		"Diğer Gelirler",
	}

	// Gider kategorileri
	expenseCategories := []string{
		"Yem",
		"Gübre",
		"Tohum",
		"İlaç",
		"Akaryakıt",
		"Elektrik",
		"Su",
		"İşçilik",
		"Veteriner",
		"Bakım-Onarım",
		"Sigorta",
		"Vergi",
		"Diğer Giderler",
	}

	categories := map[string][]string{
		"income":  incomeCategories,
		"expense": expenseCategories,
	}

	utils.SuccessResponse(c, categories, "Kategoriler başarıyla getirildi")
}

// GetFinanceAnalysis gelir-gider analizi
// @Summary Gelir-gider analizi
// @Description Finansal analiz verileri getirir
// @Tags Finance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param period query string false "Periyot"
// @Param startDate query string false "Başlangıç tarihi"
// @Param endDate query string false "Bitiş tarihi"
// @Success 200 {object} models.APIResponse{data=map[string]interface{}}
// @Failure 401 {object} models.APIResponse
// @Router /finance/analysis [get]
func (h *FinanceHandler) GetFinanceAnalysis(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	period := c.DefaultQuery("period", "month")
	startDate := c.DefaultQuery("startDate", "")
	endDate := c.DefaultQuery("endDate", "")

	// Tarih aralığını belirle
	if startDate == "" || endDate == "" {
		now := time.Now()
		switch period {
		case "month":
			startDate = now.AddDate(0, -6, 0).Format("2006-01-02")
			endDate = now.Format("2006-01-02")
		case "quarter":
			startDate = now.AddDate(0, -12, 0).Format("2006-01-02")
			endDate = now.Format("2006-01-02")
		case "year":
			startDate = now.AddDate(-3, 0, 0).Format("2006-01-02")
			endDate = now.Format("2006-01-02")
		}
	}

	// Aylık analiz
	rows, err := h.db.Query(`
		SELECT strftime('%Y-%m', date) as month,
		       SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END) as income,
		       SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END) as expense
		FROM transactions 
		WHERE user_id = ? AND date >= ? AND date <= ?
		GROUP BY strftime('%Y-%m', date)
		ORDER BY month
	`, userID, startDate, endDate)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Aylık analiz alınamadı", err.Error())
		return
	}
	defer rows.Close()

	var monthly []map[string]interface{}
	for rows.Next() {
		var month string
		var income, expense float64

		err := rows.Scan(&month, &income, &expense)
		if err != nil {
			continue
		}

		monthly = append(monthly, map[string]interface{}{
			"month":   month,
			"income":  income,
			"expense": expense,
			"profit":  income - expense,
		})
	}

	// Kategori bazında analiz
	rows, err = h.db.Query(`
		SELECT category, SUM(amount) as amount
		FROM transactions 
		WHERE user_id = ? AND date >= ? AND date <= ?
		GROUP BY category
		ORDER BY amount DESC
	`, userID, startDate, endDate)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Kategori analizi alınamadı", err.Error())
		return
	}
	defer rows.Close()

	var byCategory []map[string]interface{}
	var totalAmount float64
	for rows.Next() {
		var category string
		var amount float64

		err := rows.Scan(&category, &amount)
		if err != nil {
			continue
		}

		totalAmount += amount
		byCategory = append(byCategory, map[string]interface{}{
			"category": category,
			"amount":   amount,
		})
	}

	// Yüzdeleri hesapla
	for i := range byCategory {
		if totalAmount > 0 {
			amount := byCategory[i]["amount"].(float64)
			byCategory[i]["percentage"] = amount / totalAmount * 100
		} else {
			byCategory[i]["percentage"] = 0
		}
	}

	analysis := map[string]interface{}{
		"monthly":    monthly,
		"byCategory": byCategory,
	}

	utils.SuccessResponse(c, analysis, "Finansal analiz başarıyla getirildi")
}
