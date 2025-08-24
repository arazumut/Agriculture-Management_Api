package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"agri-management-api/internal/models"
	"agri-management-api/internal/utils"

	"github.com/gin-gonic/gin"
)

// ReportsHandler rapor işlemlerini yönetir
type ReportsHandler struct {
	db *sql.DB
}

// NewReportsHandler yeni reports handler oluşturur
func NewReportsHandler(db *sql.DB) *ReportsHandler {
	return &ReportsHandler{db: db}
}

// GetReports rapor listesi
// @Summary Rapor listesi
// @Description Kullanıcının raporlarını listeler
// @Tags Reports
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param type query string false "Rapor türü"
// @Param period query string false "Periyot"
// @Success 200 {object} models.APIResponse{data=[]map[string]interface{}}
// @Failure 401 {object} models.APIResponse
// @Router /reports [get]
func (h *ReportsHandler) GetReports(c *gin.Context) {
	_, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	reportType := c.DefaultQuery("type", "all")
	period := c.DefaultQuery("period", "all")

	// Mock rapor listesi (gerçek uygulamada DB'den gelecek)
	reports := []map[string]interface{}{
		{
			"id":            utils.GenerateID(),
			"title":         "Aylık Finansal Rapor",
			"type":          "financial",
			"description":   "Geçen ay için gelir, gider ve kar analizi",
			"generatedDate": time.Now().AddDate(0, 0, -1).Format("2006-01-02T15:04:05Z"),
			"period":        "2024-01",
			"format":        "pdf",
			"downloadUrl":   "/api/v1/reports/download/report-001.pdf",
		},
		{
			"id":            utils.GenerateID(),
			"title":         "Üretim Performans Raporu",
			"type":          "production",
			"description":   "Çeyreklik üretim performansı ve verimlilik analizi",
			"generatedDate": time.Now().AddDate(0, 0, -7).Format("2006-01-02T15:04:05Z"),
			"period":        "Q1-2024",
			"format":        "excel",
			"downloadUrl":   "/api/v1/reports/download/report-002.xlsx",
		},
		{
			"id":            utils.GenerateID(),
			"title":         "Hayvancılık Sağlık Raporu",
			"type":          "livestock",
			"description":   "Hayvan sağlığı ve aşılama durumu raporu",
			"generatedDate": time.Now().AddDate(0, 0, -14).Format("2006-01-02T15:04:05Z"),
			"period":        "2024-01",
			"format":        "pdf",
			"downloadUrl":   "/api/v1/reports/download/report-003.pdf",
		},
		{
			"id":            utils.GenerateID(),
			"title":         "Arazi Kullanım Raporu",
			"type":          "land",
			"description":   "Arazi kullanımı ve verimlilik analizi",
			"generatedDate": time.Now().AddDate(0, 0, -21).Format("2006-01-02T15:04:05Z"),
			"period":        "2023",
			"format":        "csv",
			"downloadUrl":   "/api/v1/reports/download/report-004.csv",
		},
	}

	// Filtreleme
	var filteredReports []map[string]interface{}
	for _, report := range reports {
		if reportType != "all" && report["type"] != reportType {
			continue
		}
		if period != "all" && report["period"] != period {
			continue
		}
		filteredReports = append(filteredReports, report)
	}

	utils.SuccessResponse(c, filteredReports, "Raporlar başarıyla getirildi")
}

// GenerateReport rapor oluşturma
// @Summary Rapor oluşturma
// @Description Yeni rapor oluşturur
// @Tags Reports
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "Rapor parametreleri"
// @Success 201 {object} models.APIResponse{data=map[string]interface{}}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /reports/generate [post]
func (h *ReportsHandler) GenerateReport(c *gin.Context) {
	_, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	var req struct {
		Type          string   `json:"type"`
		Period        string   `json:"period"`
		StartDate     string   `json:"startDate"`
		EndDate       string   `json:"endDate"`
		Format        string   `json:"format"`
		IncludeCharts bool     `json:"includeCharts"`
		Categories    []string `json:"categories"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	// Gerekli alanları kontrol et
	if utils.IsEmptyString(req.Type) || utils.IsEmptyString(req.Format) {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_FIELDS", "Rapor türü ve formatı gerekli", nil)
		return
	}

	// Rapor oluşturma işlemi simülasyonu
	reportID := utils.GenerateID()

	// Gerçek uygulamada burada:
	// 1. Seçili verileri DB'den çek
	// 2. Raporu oluştur (PDF, Excel, CSV)
	// 3. Dosyayı storage'a kaydet
	// 4. Download URL'i oluştur

	report := map[string]interface{}{
		"id":            reportID,
		"title":         h.getReportTitle(req.Type, req.Period),
		"type":          req.Type,
		"description":   h.getReportDescription(req.Type),
		"generatedDate": time.Now().Format("2006-01-02T15:04:05Z"),
		"period":        req.Period,
		"format":        req.Format,
		"status":        "completed",
		"downloadUrl":   "/api/v1/reports/" + reportID + "/download",
		"parameters": map[string]interface{}{
			"startDate":     req.StartDate,
			"endDate":       req.EndDate,
			"includeCharts": req.IncludeCharts,
			"categories":    req.Categories,
		},
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    report,
		Message: "Rapor başarıyla oluşturuldu",
	})
}

// DownloadReport rapor indirme
// @Summary Rapor indirme
// @Description Belirli bir raporu indirir
// @Tags Reports
// @Accept json
// @Produce application/octet-stream
// @Security BearerAuth
// @Param id path string true "Rapor ID"
// @Success 200 {file} binary
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /reports/{id}/download [get]
func (h *ReportsHandler) DownloadReport(c *gin.Context) {
	_, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	reportID := c.Param("id")
	if utils.IsEmptyString(reportID) {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_ID", "Rapor ID gerekli", nil)
		return
	}

	// Gerçek uygulamada dosya storage'dan alınacak
	// Şimdilik mock response
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=rapor-"+reportID+".pdf")
	c.Data(http.StatusOK, "application/pdf", []byte("Mock PDF content"))
}

// GetPerformanceMetrics performans metrikleri
// @Summary Performans metrikleri
// @Description Performans metriklerini getirir
// @Tags Reports
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param period query string false "Periyot"
// @Success 200 {object} models.APIResponse{data=map[string]interface{}}
// @Failure 401 {object} models.APIResponse
// @Router /reports/performance-metrics [get]
func (h *ReportsHandler) GetPerformanceMetrics(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	_ = c.DefaultQuery("period", "month")

	// Performans metriklerini hesapla
	efficiency := h.calculateEfficiency(userID)
	productivity := h.calculateProductivity(userID)
	profitability := h.calculateProfitability(userID)
	sustainability := h.calculateSustainability(userID)

	metrics := map[string]interface{}{
		"efficiency":     efficiency,
		"productivity":   productivity,
		"profitability":  profitability,
		"sustainability": sustainability,
		"trends": []map[string]interface{}{
			{
				"metric": "efficiency",
				"value":  efficiency,
				"change": 5.2,
				"trend":  "up",
			},
			{
				"metric": "productivity",
				"value":  productivity,
				"change": -2.1,
				"trend":  "down",
			},
			{
				"metric": "profitability",
				"value":  profitability,
				"change": 8.7,
				"trend":  "up",
			},
			{
				"metric": "sustainability",
				"value":  sustainability,
				"change": 3.4,
				"trend":  "up",
			},
		},
	}

	utils.SuccessResponse(c, metrics, "Performans metrikleri başarıyla getirildi")
}

// GetComparisonAnalysis karşılaştırma analizi
// @Summary Karşılaştırma analizi
// @Description İki periyot arasında karşılaştırma analizi yapar
// @Tags Reports
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param period1 query string true "İlk periyot"
// @Param period2 query string true "İkinci periyot"
// @Param metrics query string false "Karşılaştırılacak metrikler (virgülle ayrılmış)"
// @Success 200 {object} models.APIResponse{data=map[string]interface{}}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /reports/comparison [get]
func (h *ReportsHandler) GetComparisonAnalysis(c *gin.Context) {
	_, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	period1 := c.Query("period1")
	period2 := c.Query("period2")
	_ = c.DefaultQuery("metrics", "income,expense,profit,production")

	if period1 == "" || period2 == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_PERIODS", "İki periyot da gerekli", nil)
		return
	}

	// Karşılaştırma analizi (mock data)
	comparison := map[string]interface{}{
		"period1": period1,
		"period2": period2,
		"metrics": map[string]interface{}{
			"income": map[string]interface{}{
				"period1": 125000,
				"period2": 142000,
				"change":  13.6,
				"trend":   "up",
			},
			"expense": map[string]interface{}{
				"period1": 89000,
				"period2": 95000,
				"change":  6.7,
				"trend":   "up",
			},
			"profit": map[string]interface{}{
				"period1": 36000,
				"period2": 47000,
				"change":  30.6,
				"trend":   "up",
			},
			"production": map[string]interface{}{
				"period1": 2500,
				"period2": 2750,
				"change":  10.0,
				"trend":   "up",
			},
		},
		"summary": map[string]interface{}{
			"overallTrend":   "positive",
			"keyImprovement": "Kar artışı %30.6",
			"areaForFocus":   "Gider kontrolü",
		},
	}

	utils.SuccessResponse(c, comparison, "Karşılaştırma analizi başarıyla getirildi")
}

// Helper functions

func (h *ReportsHandler) getReportTitle(reportType, period string) string {
	titles := map[string]string{
		"financial":  "Finansal Rapor",
		"production": "Üretim Raporu",
		"livestock":  "Hayvancılık Raporu",
		"land":       "Arazi Raporu",
	}

	if title, exists := titles[reportType]; exists {
		return title + " - " + period
	}
	return "Genel Rapor - " + period
}

func (h *ReportsHandler) getReportDescription(reportType string) string {
	descriptions := map[string]string{
		"financial":  "Gelir, gider ve karlılık analizi",
		"production": "Üretim performansı ve verimlilik analizi",
		"livestock":  "Hayvan sağlığı ve üretkenlik raporu",
		"land":       "Arazi kullanımı ve verimlilik analizi",
	}

	if desc, exists := descriptions[reportType]; exists {
		return desc
	}
	return "Genel analiz raporu"
}

func (h *ReportsHandler) calculateEfficiency(userID string) float64 {
	// Verimlilik hesaplama algoritması
	// Gerçek uygulamada karmaşık hesaplamalar yapılacak
	return 85.5
}

func (h *ReportsHandler) calculateProductivity(userID string) float64 {
	// Üretkenlik hesaplama algoritması
	return 92.3
}

func (h *ReportsHandler) calculateProfitability(userID string) float64 {
	// Karlılık hesaplama algoritması
	return 78.9
}

func (h *ReportsHandler) calculateSustainability(userID string) float64 {
	// Sürdürülebilirlik hesaplama algoritması
	return 81.2
}
