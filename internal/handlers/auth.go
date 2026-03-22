package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hnfnfl/family-tree/internal/middleware"
	"github.com/hnfnfl/family-tree/internal/repository"
)

type AuthHandler struct {
	userRepo      *repository.UserRepository
	jwtSecret     string
	jwtExpireHour int
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role"`
	PersonID string `json:"person_id,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type AuthResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         User      `json:"user"`
}

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	PersonID string `json:"person_id,omitempty"`
}

func NewAuthHandler(userRepo *repository.UserRepository, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		userRepo:      userRepo,
		jwtSecret:     jwtSecret,
		jwtExpireHour: 24,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default role is VIEWER
	role := req.Role
	if role == "" {
		role = "VIEWER"
	}

	// Validate role
	validRoles := map[string]bool{"VIEWER": true, "EDITOR": true, "ADMIN": true}
	if !validRoles[role] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role. Must be VIEWER, EDITOR, or ADMIN"})
		return
	}

	user := &repository.User{
		Email: req.Email,
		Role:  role,
	}

	if req.PersonID != "" {
		user.PersonID = &req.PersonID
	}

	// Create user
	createdUser, err := h.userRepo.Create(c.Request.Context(), user, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate JWT token
	expiresAt := time.Now().Add(time.Hour * time.Duration(h.jwtExpireHour))
	token, err := middleware.GenerateToken(
		createdUser.ID,
		createdUser.Email,
		createdUser.Role,
		req.PersonID,
		h.jwtSecret,
		h.jwtExpireHour,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		Token:        token,
		RefreshToken: *createdUser.RefreshToken,
		ExpiresAt:    expiresAt,
		User: User{
			ID:       createdUser.ID,
			Email:    createdUser.Email,
			Role:     createdUser.Role,
			PersonID: req.PersonID,
		},
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user by email
	user, err := h.userRepo.GetByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Verify password
	if !h.userRepo.VerifyPassword(user, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Generate JWT token
	expiresAt := time.Now().Add(time.Hour * time.Duration(h.jwtExpireHour))
	personID := ""
	if user.PersonID != nil {
		personID = *user.PersonID
	}

	token, err := middleware.GenerateToken(
		user.ID,
		user.Email,
		user.Role,
		personID,
		h.jwtSecret,
		h.jwtExpireHour,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token:        token,
		RefreshToken: *user.RefreshToken,
		ExpiresAt:    expiresAt,
		User: User{
			ID:       user.ID,
			Email:    user.Email,
			Role:     user.Role,
			PersonID: personID,
		},
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement refresh token validation and rotation
	// For now, return not implemented
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Refresh token not implemented yet"})
}

func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, _ := c.Get("user_id")
	email, _ := c.Get("email")
	role, _ := c.Get("role")
	personID, _ := c.Get("person_id")

	c.JSON(http.StatusOK, gin.H{
		"user_id":    userID,
		"email":      email,
		"role":       role,
		"person_id":  personID,
	})
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	type UpdateProfileRequest struct {
		PersonID string `json:"person_id,omitempty"`
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var personID *string
	if req.PersonID != "" {
		personID = &req.PersonID
	}

	updatedUser, err := h.userRepo.UpdateProfile(c.Request.Context(), userID.(string), personID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, User{
		ID:       updatedUser.ID,
		Email:    updatedUser.Email,
		Role:     updatedUser.Role,
		PersonID: req.PersonID,
	})
}
