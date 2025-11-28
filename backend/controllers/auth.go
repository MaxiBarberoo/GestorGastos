package controllers

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"gestor-gastos/models"
)

func (h *Handler) Register(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, "Los datos enviados no son válidos", err)
		return
	}

	var exists bool
	if err := h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", req.Email).Scan(&exists); err != nil {
		respondError(c, http.StatusInternalServerError, "No se pudo verificar si el email ya existe", err)
		return
	}
	if exists {
		respondError(c, http.StatusBadRequest, "El correo ingresado ya está registrado", nil)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "No se pudo crear el usuario", err)
		return
	}

	var u models.User
	err = h.DB.QueryRow(
		`INSERT INTO users (name, email, password_hash) VALUES ($1, $2, $3)
			RETURNING id, name, email, created_at`,
		req.Name, req.Email, string(hash),
	).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "No se pudo guardar el usuario", err)
		return
	}

	token, err := h.generateToken(u.ID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "No se pudo crear la sesión después del registro", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user": u, "token": token})
}

func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, "Los datos de acceso no son válidos", err)
		return
	}

	var u models.User
	var passwordHash string
	err := h.DB.QueryRow(
		`SELECT id, name, email, password_hash, created_at FROM users WHERE email=$1`,
		req.Email,
	).Scan(&u.ID, &u.Name, &u.Email, &passwordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondError(c, http.StatusUnauthorized, "Credenciales inválidas", nil)
			return
		}
		respondError(c, http.StatusInternalServerError, "No se pudo buscar el usuario", err)
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)) != nil {
		respondError(c, http.StatusUnauthorized, "Credenciales inválidas", nil)
		return
	}

	token, err := h.generateToken(u.ID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "No se pudo crear la sesión", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": u, "token": token})
}

func (h *Handler) Me(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		respondError(c, http.StatusUnauthorized, "No se encontró la sesión del usuario", nil)
		return
	}

	var u models.User
	err := h.DB.QueryRow(
		`SELECT id, name, email, created_at FROM users WHERE id=$1`,
		userID,
	).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "No se pudo obtener el perfil del usuario", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": u})
}

func (h *Handler) generateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.JWTSecret)
}
