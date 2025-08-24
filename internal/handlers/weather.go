package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"agri-management-api/internal/models"
	"agri-management-api/internal/utils"

	"github.com/gin-gonic/gin"
)

// WeatherHandler hava durumu işlemlerini yönetir
type WeatherHandler struct {
	db *sql.DB
}

// NewWeatherHandler yeni weather handler oluşturur
func NewWeatherHandler(db *sql.DB) *WeatherHandler {
	return &WeatherHandler{db: db}
}

// GetCurrentWeather güncel hava durumu
// @Summary Güncel hava durumu
// @Description Belirtilen koordinatlar için güncel hava durumu getirir
// @Tags Weather
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param lat query number true "Enlem"
// @Param lon query number true "Boylam"
// @Success 200 {object} models.APIResponse{data=models.Weather}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /weather/current [get]
func (h *WeatherHandler) GetCurrentWeather(c *gin.Context) {
	_, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	latStr := c.Query("lat")
	lonStr := c.Query("lon")

	if latStr == "" || lonStr == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_COORDINATES", "Enlem ve boylam gerekli", nil)
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_LATITUDE", "Geçersiz enlem değeri", nil)
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_LONGITUDE", "Geçersiz boylam değeri", nil)
		return
	}

	// Hava durumu verilerini al (OpenWeatherMap API simülasyonu)
	weather, err := h.fetchCurrentWeather(lat, lon)
	if err != nil {
		// API hatası durumunda mock data döndür
		weather = h.getMockCurrentWeather(lat, lon)
	}

	utils.SuccessResponse(c, weather, "Güncel hava durumu başarıyla getirildi")
}

// GetWeatherForecast hava durumu tahmini
// @Summary Hava durumu tahmini
// @Description Belirtilen koordinatlar için 7 günlük hava durumu tahmini getirir
// @Tags Weather
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param lat query number true "Enlem"
// @Param lon query number true "Boylam"
// @Param days query int false "Gün sayısı (varsayılan: 7)"
// @Success 200 {object} models.APIResponse{data=[]models.WeatherForecast}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /weather/forecast [get]
func (h *WeatherHandler) GetWeatherForecast(c *gin.Context) {
	_, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	latStr := c.Query("lat")
	lonStr := c.Query("lon")
	daysStr := c.DefaultQuery("days", "7")

	if latStr == "" || lonStr == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_COORDINATES", "Enlem ve boylam gerekli", nil)
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_LATITUDE", "Geçersiz enlem değeri", nil)
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_LONGITUDE", "Geçersiz boylam değeri", nil)
		return
	}

	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 7 {
		days = 7
	}

	// Hava durumu tahminini al
	forecast, err := h.fetchWeatherForecast(lat, lon, days)
	if err != nil {
		// API hatası durumunda mock data döndür
		forecast = h.getMockWeatherForecast(days)
	}

	utils.SuccessResponse(c, forecast, "Hava durumu tahmini başarıyla getirildi")
}

// GetAgriculturalAlerts tarımsal uyarılar
// @Summary Tarımsal uyarılar
// @Description Belirtilen koordinatlar için tarımsal uyarıları getirir
// @Tags Weather
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param lat query number true "Enlem"
// @Param lon query number true "Boylam"
// @Success 200 {object} models.APIResponse{data=[]models.AgriculturalAlert}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /weather/agricultural-alerts [get]
func (h *WeatherHandler) GetAgriculturalAlerts(c *gin.Context) {
	_, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	latStr := c.Query("lat")
	lonStr := c.Query("lon")

	if latStr == "" || lonStr == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_COORDINATES", "Enlem ve boylam gerekli", nil)
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_LATITUDE", "Geçersiz enlem değeri", nil)
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_LONGITUDE", "Geçersiz boylam değeri", nil)
		return
	}

	// Tarımsal uyarıları al
	alerts := h.getAgriculturalAlerts(lat, lon)

	utils.SuccessResponse(c, alerts, "Tarımsal uyarılar başarıyla getirildi")
}

// fetchCurrentWeather gerçek API'den güncel hava durumu alır
func (h *WeatherHandler) fetchCurrentWeather(lat, lon float64) (*models.Weather, error) {
	// OpenWeatherMap API key (gerçek uygulamada environment variable'dan alınacak)
	apiKey := "YOUR_API_KEY"
	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&appid=%s&units=metric&lang=tr", lat, lon, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResponse struct {
		Name string `json:"name"`
		Main struct {
			Temp     float64 `json:"temp"`
			Humidity float64 `json:"humidity"`
			Pressure float64 `json:"pressure"`
		} `json:"main"`
		Weather []struct {
			Main        string `json:"main"`
			Description string `json:"description"`
			Icon        string `json:"icon"`
		} `json:"weather"`
		Wind struct {
			Speed float64 `json:"speed"`
			Deg   float64 `json:"deg"`
		} `json:"wind"`
		Visibility int `json:"visibility"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}

	weather := &models.Weather{
		Location:      apiResponse.Name,
		Temperature:   apiResponse.Main.Temp,
		Humidity:      apiResponse.Main.Humidity,
		WindSpeed:     apiResponse.Wind.Speed,
		WindDirection: getWindDirection(apiResponse.Wind.Deg),
		Pressure:      apiResponse.Main.Pressure,
		Visibility:    float64(apiResponse.Visibility) / 1000, // m to km
		UVIndex:       5.0,                                    // Mock value
		Condition:     apiResponse.Weather[0].Description,
		Icon:          apiResponse.Weather[0].Icon,
		LastUpdated:   time.Now().Format("2006-01-02T15:04:05Z"),
	}

	return weather, nil
}

// fetchWeatherForecast gerçek API'den hava durumu tahmini alır
func (h *WeatherHandler) fetchWeatherForecast(lat, lon float64, days int) ([]models.WeatherForecast, error) {
	// Bu fonksiyon gerçek API çağrısı yapacak
	// Şimdilik mock data döndürüyoruz
	return h.getMockWeatherForecast(days), nil
}

// getMockCurrentWeather mock güncel hava durumu
func (h *WeatherHandler) getMockCurrentWeather(lat, lon float64) *models.Weather {
	return &models.Weather{
		Location:      "İstanbul",
		Temperature:   22.5,
		Humidity:      65.0,
		WindSpeed:     12.5,
		WindDirection: "KB",
		Pressure:      1015.0,
		Visibility:    10.0,
		UVIndex:       6.0,
		Condition:     "Parçalı bulutlu",
		Icon:          "02d",
		LastUpdated:   time.Now().Format("2006-01-02T15:04:05Z"),
	}
}

// getMockWeatherForecast mock hava durumu tahmini
func (h *WeatherHandler) getMockWeatherForecast(days int) []models.WeatherForecast {
	var forecast []models.WeatherForecast

	conditions := []string{"Güneşli", "Parçalı bulutlu", "Bulutlu", "Hafif yağmur", "Güneşli"}
	icons := []string{"01d", "02d", "03d", "10d", "01d"}

	for i := 0; i < days; i++ {
		date := time.Now().AddDate(0, 0, i+1)

		forecast = append(forecast, models.WeatherForecast{
			Date:       date.Format("2006-01-02"),
			MinTemp:    15.0 + float64(i%3),
			MaxTemp:    25.0 + float64(i%5),
			Condition:  conditions[i%len(conditions)],
			Icon:       icons[i%len(icons)],
			Humidity:   60.0 + float64(i%20),
			RainChance: float64((i * 15) % 80),
			WindSpeed:  8.0 + float64(i%10),
		})
	}

	return forecast
}

// getAgriculturalAlerts tarımsal uyarıları döndürür
func (h *WeatherHandler) getAgriculturalAlerts(lat, lon float64) []models.AgriculturalAlert {
	// Bu fonksiyon meteoroloji verilerine göre tarımsal uyarılar üretecek
	// Şimdilik örnek uyarılar döndürüyoruz

	alerts := []models.AgriculturalAlert{
		{
			Type:        "frost",
			Severity:    "medium",
			Title:       "Don Uyarısı",
			Description: "Bu gece sıcaklık 0°C'nin altına düşebilir. Hassas bitkileri koruyun.",
			StartDate:   time.Now().Format("2006-01-02T15:04:05Z"),
			EndDate:     time.Now().AddDate(0, 0, 1).Format("2006-01-02T15:04:05Z"),
			Recommendations: []string{
				"Hassas bitkileri örtü ile koruyun",
				"Sulama sistemlerini donmaya karşı koruyun",
				"Hayvanlar için sıcak barınak sağlayın",
			},
		},
		{
			Type:        "drought",
			Severity:    "low",
			Title:       "Kuraklık Takibi",
			Description: "Son 10 gündür yağış almadınız. Su kaynaklarınızı kontrol edin.",
			StartDate:   time.Now().AddDate(0, 0, -10).Format("2006-01-02T15:04:05Z"),
			EndDate:     time.Now().AddDate(0, 0, 3).Format("2006-01-02T15:04:05Z"),
			Recommendations: []string{
				"Su tasarrufu yapın",
				"Damla sulama sistemini aktif edin",
				"Toprak nemini kontrol edin",
			},
		},
	}

	return alerts
}

// getWindDirection rüzgar derecesini yön olarak çevirir
func getWindDirection(deg float64) string {
	directions := []string{"K", "KKD", "KD", "DKD", "D", "DGD", "GD", "GGD", "G", "GGB", "GB", "BGB", "B", "BBK", "BK", "KBK"}
	index := int((deg + 11.25) / 22.5)
	return directions[index%16]
}

// SaveWeatherData hava durumu verilerini cache'e kaydet
func (h *WeatherHandler) SaveWeatherData(lat, lon float64, weather *models.Weather) error {
	// Hava durumu verilerini veritabanına cache olarak kaydet
	_, err := h.db.Exec(`
		INSERT OR REPLACE INTO weather_cache (lat, lon, data, cached_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`, lat, lon, weather)

	return err
}

// GetCachedWeatherData cache'den hava durumu verilerini al
func (h *WeatherHandler) GetCachedWeatherData(lat, lon float64) (*models.Weather, error) {
	var weatherData string
	var cachedAt time.Time

	err := h.db.QueryRow(`
		SELECT data, cached_at 
		FROM weather_cache 
		WHERE lat = ? AND lon = ? AND cached_at > datetime('now', '-1 hour')
	`, lat, lon).Scan(&weatherData, &cachedAt)

	if err != nil {
		return nil, err
	}

	var weather models.Weather
	err = json.Unmarshal([]byte(weatherData), &weather)
	if err != nil {
		return nil, err
	}

	return &weather, nil
}
