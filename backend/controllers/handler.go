package controllers

import "database/sql"

// Handler wires dependencies into every controller.
type Handler struct {
	DB        *sql.DB
	JWTSecret []byte
}

func NewHandler(db *sql.DB, jwtSecret string) *Handler {
	return &Handler{
		DB:        db,
		JWTSecret: []byte(jwtSecret),
	}
}
