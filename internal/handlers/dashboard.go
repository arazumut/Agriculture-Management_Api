package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"agri-management-api/internal/models"
	"agri-management-api/internal/utils"

	"github.com/gin-gonic/gin"
)

// DashboardHandler dashboard işlemlerini yönetir
type DashboardHandler struct {
	db *sql.DB
}

// NewDashboardHandler yeni dashboard handler oluşturur
func NewDashboardHandler(db *sql.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

// GetSummary dashboard özet verileri
// @Summary Dashboard özet
// @Description Dashboard için özet istatistikleri getirir
// @Tags Dashboard
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=models.DashboardSummary}
// @Failure 401 {object} models.APIResponse
// @Router /dashboard/summary [get]
func (h *DashboardHandler) GetSummary(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	// Hayvan sayısı
	var animalCount int
	err = h.db.QueryRow("SELECT COUNT(*) FROM livestock WHERE user_id = ?", userID).Scan(&animalCount)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Hayvan sayısı alınamadı", err.Error())
		return
	}

	// Arazi bilgileri
	var landCount int
	var totalArea float64
	var avgProductivity float64
	err = h.db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(area), 0), COALESCE(AVG(productivity), 0)
		FROM lands WHERE user_id = ? AND status = 'active'
	`, userID).Scan(&landCount, &totalArea, &avgProductivity)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Arazi bilgileri alınamadı", err.Error())
		return
	}

	// Aylık gelir
	var monthlyIncome float64
	currentMonth := time.Now().Format("2006-01")
	err = h.db.QueryRow(`
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions 
		WHERE user_id = ? AND type = 'income' AND strftime('%Y-%m', date) = ?
	`, userID, currentMonth).Scan(&monthlyIncome)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Aylık gelir alınamadı", err.Error())
		return
	}

	// Aylık gider
	var monthlyExpense float64
	err = h.db.QueryRow(`
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions 
		WHERE user_id = ? AND type = 'expense' AND strftime('%Y-%m', date) = ?
	`, userID, currentMonth).Scan(&monthlyExpense)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Aylık gider alınamadı", err.Error())
		return
	}

	// Aktif ürün sayısı
	var activeProductCount int
	var productCategoryCount int
	err = h.db.QueryRow(`
		SELECT COUNT(*), COUNT(DISTINCT category)
		FROM production 
		WHERE user_id = ? AND status = 'active'
	`, userID).Scan(&activeProductCount, &productCategoryCount)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Ürün bilgileri alınamadı", err.Error())
		return
	}

	// Trend hesaplama (basit implementasyon)
	lastMonth := time.Now().AddDate(0, -1, 0).Format("2006-01")
	var lastMonthIncome float64
	var lastMonthExpense float64
	
	h.db.QueryRow(`
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions 
		WHERE user_id = ? AND type = 'income' AND strftime('%Y-%m', date) = ?
	`, userID, lastMonth).Scan(&lastMonthIncome)
	
	h.db.QueryRow(`
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions 
		WHERE user_id = ? AND type = 'expense' AND strftime('%Y-%m', date) = ?
	`, userID, lastMonth).Scan(&lastMonthExpense)

	incomeTrend := "+0"
	expenseTrend := "+0"
	
	if lastMonthIncome > 0 {
		change := ((monthlyIncome - lastMonthIncome) / lastMonthIncome) * 100
		if change > 0 {
			incomeTrend = "+" + strconv.FormatFloat(change, 'f', 1, 64) + "%"
		} else {
			incomeTrend = strconv.FormatFloat(change, 'f', 1, 64) + "%"
		}
	}
	
	if lastMonthExpense > 0 {
		change := ((monthlyExpense - lastMonthExpense) / lastMonthExpense) * 100
		if change > 0 {
			expenseTrend = "+" + strconv.FormatFloat(change, 'f', 1, 64) + "%"
		} else {
			expenseTrend = strconv.FormatFloat(change, 'f', 1, 64) + "%"
		}
	}

	summary := models.DashboardSummary{
		TotalAnimals: models.AnimalSummary{
			Count:      animalCount,
			Trend:      "+0",
			Percentage: 0,
		},
		TotalLands: models.LandSummary{
			Area:        totalArea,
			Count:       landCount,
			Productivity: avgProductivity,
		},
		MonthlyIncome: models.FinanceSummary{
			Amount:   monthlyIncome,
			Currency: "TRY",
			Trend:    incomeTrend,
		},
		MonthlyExpense: models.FinanceSummary{
			Amount:   monthlyExpense,
			Currency: "TRY",
			Trend:    expenseTrend,
		},
		ActiveProducts: models.ProductSummary{
			Count:      activeProductCount,
			Categories: productCategoryCount,
		},
	}

	utils.SuccessResponse(c, summary, "Dashboard özeti başarıyla getirildi")
}

// GetRecentActivities son aktiviteler
// @Summary Son aktiviteler
// @Description Son aktiviteleri listeler
// @Tags Dashboard
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit (default: 10)"
// @Success 200 {object} models.APIResponse{data=[]map[string]interface{}}
// @Failure 401 {object} models.APIResponse
// @Router /dashboard/recent-activities [get]
func (h *DashboardHandler) GetRecentActivities(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		limit = 10
	}

	// Son aktiviteleri birleştir (hayvan, arazi, üretim, finans)
	activities := []map[string]interface{}{}

	// Hayvan aktiviteleri
	rows, err := h.db.Query(`
		SELECT 'health_check' as type, 'Sağlık kontrolü' as title, 
		       'Hayvan sağlık kontrolü yapıldı' as description, created_at as date,
		       'livestock' as category, '🐄' as icon
		FROM health_records hr
		JOIN livestock l ON hr.livestock_id = l.id
		WHERE l.user_id = ?
		ORDER BY hr.created_at DESC LIMIT ?
	`, userID, limit/4)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var activity map[string]interface{}
			rows.Scan(&activity)
			activities = append(activities, activity)
		}
	}

	// Arazi aktiviteleri
	rows, err = h.db.Query(`
		SELECT 'irrigation' as type, 'Sulama' as title,
		       'Arazi sulama işlemi yapıldı' as description, created_at as date,
		       'land' as category, '🌱' as icon
		FROM land_activities la
		JOIN lands l ON la.land_id = l.id
		WHERE l.user_id = ?
		ORDER BY la.created_at DESC LIMIT ?
	`, userID, limit/4)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var activity map[string]interface{}
			rows.Scan(&activity)
			activities = append(activities, activity)
		}
	}

	// Üretim aktiviteleri
	rows, err = h.db.Query(`
		SELECT 'harvest' as type, 'Hasat' as title,
		       'Ürün hasadı yapıldı' as description, created_at as date,
		       'production' as category, '🌾' as icon
		FROM production
		WHERE user_id = ?
		ORDER BY created_at DESC LIMIT ?
	`, userID, limit/4)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var activity map[string]interface{}
			rows.Scan(&activity)
			activities = append(activities, activity)
		}
	}

	// Finans aktiviteleri
	rows, err = h.db.Query(`
		SELECT type, category as title,
		       description, date as date,
		       'finance' as category, '💰' as icon
		FROM transactions
		WHERE user_id = ?
		ORDER BY date DESC LIMIT ?
	`, userID, limit/4)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var activity map[string]interface{}
			rows.Scan(&activity)
			activities = append(activities, activity)
		}
	}

	// Aktivite sayısını sınırla
	if len(activities) > limit {
		activities = activities[:limit]
	}

	utils.SuccessResponse(c, activities, "Son aktiviteler başarıyla getirildi")
}

// GetIncomeExpenseChart gelir-gider grafik verileri
// @Summary Gelir-gider grafik
// @Description Aylık gelir-gider grafik verilerini getirir
// @Tags Dashboard
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param period query string false "Period (month/quarter/year)" Enums(month, quarter, year)
// @Success 200 {object} models.APIResponse{data=map[string]interface{}}
// @Failure 401 {object} models.APIResponse
// @Router /dashboard/charts/income-expense [get]
func (h *DashboardHandler) GetIncomeExpenseChart(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	_ = c.DefaultQuery("period", "month")
	
	var labels []string
	var income []float64
	var expense []float64
	var profit []float64

	// Son 12 ay verisi
	for i := 11; i >= 0; i-- {
		date := time.Now().AddDate(0, -i, 0)
		monthStr := date.Format("2006-01")
		labels = append(labels, date.Format("Jan 2006"))

		var monthIncome, monthExpense float64
		
		h.db.QueryRow(`
			SELECT COALESCE(SUM(amount), 0)
			FROM transactions 
			WHERE user_id = ? AND type = 'income' AND strftime('%Y-%m', date) = ?
		`, userID, monthStr).Scan(&monthIncome)
		
		h.db.QueryRow(`
			SELECT COALESCE(SUM(amount), 0)
			FROM transactions 
			WHERE user_id = ? AND type = 'expense' AND strftime('%Y-%m', date) = ?
		`, userID, monthStr).Scan(&monthExpense)

		income = append(income, monthIncome)
		expense = append(expense, monthExpense)
		profit = append(profit, monthIncome-monthExpense)
	}

	chartData := map[string]interface{}{
		"labels":  labels,
		"income":  income,
		"expense": expense,
		"profit":  profit,
	}

	utils.SuccessResponse(c, chartData, "Gelir-gider grafik verileri başarıyla getirildi")
}

// GetProductionChart üretim grafik verileri
// @Summary Üretim grafik
// @Description Üretim kategorileri grafik verilerini getirir
// @Tags Dashboard
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=map[string]interface{}}
// @Failure 401 {object} models.APIResponse
// @Router /dashboard/charts/production [get]
func (h *DashboardHandler) GetProductionChart(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	rows, err := h.db.Query(`
		SELECT category, COUNT(*) as count
		FROM production 
		WHERE user_id = ? AND status = 'active'
		GROUP BY category
		ORDER BY count DESC
	`, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Üretim verileri alınamadı", err.Error())
		return
	}
	defer rows.Close()

	var categories []string
	var values []int
	var colors []string

	colorPalette := []string{"#FF6384", "#36A2EB", "#FFCE56", "#4BC0C0", "#9966FF", "#FF9F40"}

	i := 0
	for rows.Next() {
		var category string
		var count int
		rows.Scan(&category, &count)
		
		categories = append(categories, category)
		values = append(values, count)
		colors = append(colors, colorPalette[i%len(colorPalette)])
		i++
	}

	chartData := map[string]interface{}{
		"categories": categories,
		"values":     values,
		"colors":     colors,
	}

	utils.SuccessResponse(c, chartData, "Üretim grafik verileri başarıyla getirildi")
}
