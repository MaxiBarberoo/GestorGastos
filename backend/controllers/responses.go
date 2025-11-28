package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// respondError envía un error consistente al cliente con detalles opcionales.
func respondError(c *gin.Context, status int, message string, err error) {
	body := gin.H{"error": message}
	if err != nil {
		body["details"] = err.Error()
	}
	c.JSON(status, body)
}

// respondValidationError es un helper específico para errores 400.
func respondValidationError(c *gin.Context, message string, err error) {
	respondError(c, http.StatusBadRequest, message, err)
}
