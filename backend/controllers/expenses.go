package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"gestor-gastos/models"
)

func (h *Handler) ListExpenses(c *gin.Context) {
	userID := c.GetInt64("userID")
	fromParam := c.Query("from")
	toParam := c.Query("to")

	query := `SELECT id, user_id, name, tag, amount, expense_date FROM expenses WHERE user_id=$1`
	args := []interface{}{userID}

	if fromParam != "" {
		if _, err := time.Parse("2006-01-02", fromParam); err != nil {
			respondValidationError(c, "El parámetro 'from' debe usar el formato YYYY-MM-DD", err)
			return
		}
		query += fmt.Sprintf(" AND expense_date >= $%d", len(args)+1)
		args = append(args, fromParam)
	}

	if toParam != "" {
		if _, err := time.Parse("2006-01-02", toParam); err != nil {
			respondValidationError(c, "El parámetro 'to' debe usar el formato YYYY-MM-DD", err)
			return
		}
		query += fmt.Sprintf(" AND expense_date <= $%d", len(args)+1)
		args = append(args, toParam)
	}

	query += " ORDER BY expense_date DESC, id DESC"

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "No se pudo obtener la lista de gastos", err)
		return
	}
	defer rows.Close()

	var expenses []models.Expense
	for rows.Next() {
		var exp models.Expense
		var date time.Time
		if err := rows.Scan(&exp.ID, &exp.UserID, &exp.Name, &exp.Tag, &exp.Amount, &date); err != nil {
			respondError(c, http.StatusInternalServerError, "No se pudo leer la lista de gastos", err)
			return
		}
		exp.Date = date.Format("2006-01-02")
		expenses = append(expenses, exp)
	}

	c.JSON(http.StatusOK, gin.H{"expenses": expenses})
}

func (h *Handler) CreateExpense(c *gin.Context) {
	userID := c.GetInt64("userID")
	var req struct {
		Name   string  `json:"name" binding:"required"`
		Tag    string  `json:"tag" binding:"required"`
		Amount float64 `json:"amount" binding:"required"`
		Date   string  `json:"date" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, "Los datos del gasto no son válidos", err)
		return
	}

	expenseDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		respondValidationError(c, "La fecha del gasto no tiene el formato correcto", err)
		return
	}

	var exp models.Expense
	err = h.DB.QueryRow(
		`INSERT INTO expenses (user_id, name, tag, amount, expense_date)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, user_id, name, tag, amount, expense_date`,
		userID, req.Name, req.Tag, req.Amount, expenseDate,
	).Scan(&exp.ID, &exp.UserID, &exp.Name, &exp.Tag, &exp.Amount, &expenseDate)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "No se pudo guardar el gasto", err)
		return
	}
	exp.Date = expenseDate.Format("2006-01-02")

	c.JSON(http.StatusCreated, gin.H{
		"expense": exp,
	})
}

func (h *Handler) DeleteExpense(c *gin.Context) {
	userID := c.GetInt64("userID")
	idParam := c.Param("id")
	expenseID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		respondValidationError(c, "El identificador del gasto no es válido", err)
		return
	}

	tx, err := h.DB.BeginTx(c, nil)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "No se pudo iniciar la operación de eliminación", err)
		return
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		`UPDATE monthly_expenses
		 SET last_applied_at=NULL, last_applied_expense_id=NULL
		 WHERE user_id=$1 AND last_applied_expense_id=$2`,
		userID, expenseID,
	); err != nil {
		respondError(c, http.StatusInternalServerError, "No se pudo actualizar el gasto recurrente asociado", err)
		return
	}

	result, err := tx.Exec(`DELETE FROM expenses WHERE id=$1 AND user_id=$2`, expenseID, userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "No se pudo eliminar el gasto", err)
		return
	}

	rows, err := result.RowsAffected()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "No se pudo confirmar la eliminación del gasto", err)
		return
	}
	if rows == 0 {
		respondError(c, http.StatusNotFound, "No se encontró el gasto solicitado", nil)
		return
	}

	if err := tx.Commit(); err != nil {
		respondError(c, http.StatusInternalServerError, "No se pudo confirmar la eliminación del gasto", err)
		return
	}

	c.Status(http.StatusNoContent)
}
