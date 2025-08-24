package handlers

import (
	"database/sql"
	"net/http"

	"agri-management-api/internal/models"
	"agri-management-api/internal/utils"
	"agri-management-api/pkg/auth"

	"github.com/gin-gonic/gin"
)

// AuthHandler kimlik doğrulama işlemlerini yönetir
type AuthHandler struct {
	db         *sql.DB
	jwtManager *auth.JWTManager
}

// NewAuthHandler yeni auth handler oluşturur
func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{
		db:         db,
		jwtManager: auth.NewJWTManager(),
	}
}

// Register kullanıcı kaydı
// @Summary Kullanıcı kaydı
// @Description Yeni kullanıcı kaydı oluşturur
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Kayıt bilgileri"
// @Success 201 {object} models.APIResponse{data=models.AuthResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 409 {object} models.APIResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	// Şifre kontrolü
	if req.Password != req.ConfirmPassword {
		utils.ErrorResponse(c, http.StatusBadRequest, "PASSWORD_MISMATCH", "Şifreler eşleşmiyor", nil)
		return
	}

	// Email kontrolü
	var existingUser models.User
	err := h.db.QueryRow("SELECT id FROM users WHERE email = ?", req.Email).Scan(&existingUser.ID)
	if err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "EMAIL_EXISTS", "Bu email adresi zaten kullanımda", nil)
		return
	}

	// Şifreyi hash'le
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "HASH_ERROR", "Şifre hash'lenemedi", err.Error())
		return
	}

	// Kullanıcıyı oluştur
	userID := utils.GenerateID()
	_, err = h.db.Exec(`
		INSERT INTO users (id, name, email, password, farm_name, location, role, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, 'farmer', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, userID, req.Name, req.Email, hashedPassword, req.FarmName, req.Location)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DB_ERROR", "Kullanıcı oluşturulamadı", err.Error())
		return
	}

	// Token oluştur
	token, err := h.jwtManager.GenerateToken(userID, req.Email, "farmer")
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "TOKEN_ERROR", "Token oluşturulamadı", err.Error())
		return
	}

	// Refresh token oluştur
	refreshToken, err := h.jwtManager.GenerateToken(userID, req.Email, "farmer")
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "REFRESH_TOKEN_ERROR", "Refresh token oluşturulamadı", err.Error())
		return
	}

	user := models.User{
		ID:         userID,
		Name:       req.Name,
		Email:      req.Email,
		FarmName:   req.FarmName,
		Location:   req.Location,
		Role:       "farmer",
		IsVerified: false,
	}

	response := models.AuthResponse{
		User:         user,
		Token:        token,
		RefreshToken: refreshToken,
	}

	utils.SuccessResponse(c, response, "Kullanıcı başarıyla oluşturuldu")
}

// Login kullanıcı girişi
// @Summary Kullanıcı girişi
// @Description Kullanıcı girişi yapar ve token döner
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Giriş bilgileri"
// @Success 200 {object} models.APIResponse{data=models.AuthResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	// Kullanıcıyı bul
	var user models.User
	err := h.db.QueryRow(`
		SELECT id, name, email, password, avatar, role, farm_name, location, is_verified, created_at, updated_at
		FROM users WHERE email = ?
	`, req.Email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password, &user.Avatar,
		&user.Role, &user.FarmName, &user.Location, &user.IsVerified,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Geçersiz email veya şifre", nil)
		return
	}

	// Şifreyi kontrol et
	if !utils.CheckPassword(req.Password, user.Password) {
		utils.ErrorResponse(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Geçersiz email veya şifre", nil)
		return
	}

	// Token oluştur
	token, err := h.jwtManager.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "TOKEN_ERROR", "Token oluşturulamadı", err.Error())
		return
	}

	// Refresh token oluştur
	refreshToken, err := h.jwtManager.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "REFRESH_TOKEN_ERROR", "Refresh token oluşturulamadı", err.Error())
		return
	}

	response := models.AuthResponse{
		User:         user,
		Token:        token,
		RefreshToken: refreshToken,
	}

	utils.SuccessResponse(c, response, "Giriş başarılı")
}

// Refresh token yenileme
// @Summary Token yenileme
// @Description Refresh token ile yeni access token oluşturur
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Refresh token"
// @Success 200 {object} models.APIResponse{data=map[string]string}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	refreshToken, exists := req["refreshToken"]
	if !exists {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_TOKEN", "Refresh token gerekli", nil)
		return
	}

	// Token'ı yenile
	newToken, err := h.jwtManager.RefreshToken(refreshToken)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "INVALID_TOKEN", "Geçersiz refresh token", err.Error())
		return
	}

	response := map[string]string{
		"token": newToken,
	}

	utils.SuccessResponse(c, response, "Token başarıyla yenilendi")
}

// GetProfile kullanıcı profili
// @Summary Kullanıcı profili
// @Description Mevcut kullanıcının profil bilgilerini getirir
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=models.User}
// @Failure 401 {object} models.APIResponse
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	var user models.User
	err = h.db.QueryRow(`
		SELECT id, name, email, avatar, role, farm_name, location, is_verified, created_at, updated_at
		FROM users WHERE id = ?
	`, userID).Scan(
		&user.ID, &user.Name, &user.Email, &user.Avatar, &user.Role,
		&user.FarmName, &user.Location, &user.IsVerified, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "USER_NOT_FOUND", "Kullanıcı bulunamadı", err.Error())
		return
	}

	utils.SuccessResponse(c, user, "Profil bilgileri başarıyla getirildi")
}

// UpdateProfile profil güncelleme
// @Summary Profil güncelleme
// @Description Kullanıcı profil bilgilerini günceller
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.User true "Güncellenecek profil bilgileri"
// @Success 200 {object} models.APIResponse{data=models.User}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /auth/profile [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	var req models.User
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	// Profili güncelle
	_, err = h.db.Exec(`
		UPDATE users 
		SET name = ?, farm_name = ?, location = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, req.Name, req.FarmName, req.Location, userID)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_ERROR", "Profil güncellenemedi", err.Error())
		return
	}

	// Güncellenmiş profili getir
	var user models.User
	err = h.db.QueryRow(`
		SELECT id, name, email, avatar, role, farm_name, location, is_verified, created_at, updated_at
		FROM users WHERE id = ?
	`, userID).Scan(
		&user.ID, &user.Name, &user.Email, &user.Avatar, &user.Role,
		&user.FarmName, &user.Location, &user.IsVerified, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "FETCH_ERROR", "Güncellenmiş profil getirilemedi", err.Error())
		return
	}

	utils.SuccessResponse(c, user, "Profil başarıyla güncellendi")
}

// ChangePassword şifre değiştirme
// @Summary Şifre değiştirme
// @Description Kullanıcı şifresini değiştirir
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]string true "Şifre bilgileri"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /auth/change-password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Kullanıcı kimliği doğrulanamadı", nil)
		return
	}

	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	currentPassword, exists := req["currentPassword"]
	if !exists {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_CURRENT_PASSWORD", "Mevcut şifre gerekli", nil)
		return
	}

	newPassword, exists := req["newPassword"]
	if !exists {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_NEW_PASSWORD", "Yeni şifre gerekli", nil)
		return
	}

	// Mevcut şifreyi kontrol et
	var hashedPassword string
	err = h.db.QueryRow("SELECT password FROM users WHERE id = ?", userID).Scan(&hashedPassword)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "USER_NOT_FOUND", "Kullanıcı bulunamadı", err.Error())
		return
	}

	if !utils.CheckPassword(currentPassword, hashedPassword) {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_CURRENT_PASSWORD", "Mevcut şifre yanlış", nil)
		return
	}

	// Yeni şifreyi hash'le
	newHashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "HASH_ERROR", "Yeni şifre hash'lenemedi", err.Error())
		return
	}

	// Şifreyi güncelle
	_, err = h.db.Exec(`
		UPDATE users 
		SET password = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, newHashedPassword, userID)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_ERROR", "Şifre güncellenemedi", err.Error())
		return
	}

	utils.SuccessResponse(c, nil, "Şifre başarıyla değiştirildi")
}

// Logout çıkış yapma
// @Summary Çıkış yapma
// @Description Kullanıcı çıkışı yapar
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// JWT tabanlı sistemde client tarafında token'ı silmek yeterli
	// Burada ek güvenlik önlemleri alınabilir (blacklist, vs.)
	utils.SuccessResponse(c, nil, "Başarıyla çıkış yapıldı")
}
