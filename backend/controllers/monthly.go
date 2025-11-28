package controllers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"gestor-gastos/models"
)

func (h *Handler) ListMonthlyExpenses(c *gin.Context) {
	userID := c.GetInt64("userID")
	rows, err := h.DB.Query(
		`SELECT id, user_id, name, tag, amount, last_applied_at, last_applied_expense_id
		 FROM monthly_expenses
		 WHERE user_id=$1
		 ORDER BY id DESC`, userID,
	)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "No se pudo obtener la lista de gastos mensuales", err)
		return
	}
	defer rows.Close()

	var monthly []models.MonthlyExpense
	for rows.Next() {
		var item models.MonthlyExpense
		var lastApplied sql.NullTime
		var lastExpense sql.NullInt64
		if err := rows.Scan(&item.ID, &item.UserID, &item.Name, &item.Tag, &item.Amount, &lastApplied, &lastExpense); err != nil {
			respondError(c, http.StatusInternalServerError, "No se pudo leer la lista de gastos mensuales", err)
			return
		}
		if lastApplied.Valid {
			item.LastAppliedAt = &lastApplied.Time
		}
		if lastExpense.Valid {
			id := lastExpense.Int64
			item.LastExpenseID = &id
		}
		monthly = append(monthly, item)
	}

	c.JSON(http.StatusOK, gin.H{"monthlyExpenses": monthly})
}

func (h *Handler) CreateMonthlyExpense(c *gin.Context) {
	userID := c.GetInt64("userID")
	var req struct {
		Name   string  `json:"name" binding:"required"`
		Tag    string  `json:"tag" binding:"required"`
		Amount float64 `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, "Los datos del gasto recurrente no son válidos", err)
		return
	}

	var item models.MonthlyExpense
	err := h.DB.QueryRow(
		`INSERT INTO monthly_expenses (user_id, name, tag, amount)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, user_id, name, tag, amount, last_applied_at, last_applied_expense_id`,
		userID, req.Name, req.Tag, req.Amount,
	).Scan(&item.ID, &item.UserID, &item.Name, &item.Tag, &item.Amount, &item.LastAppliedAt, &item.LastExpenseID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "No se pudo guardar el gasto recurrente", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"monthlyExpense": item})
}

func (h *Handler) DeleteMonthlyExpense(c *gin.Context) {
	userID := c.GetInt64("userID")
	idParam := c.Param("id")
	itemID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		respondValidationError(c, "El identificador del gasto recurrente no es válido", err)
		return
	}

	result, err := h.DB.Exec(`DELETE FROM monthly_expenses WHERE id=$1 AND user_id=$2`, itemID, userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "No se pudo eliminar el gasto recurrente", err)
		return
	}

	rows, err := result.RowsAffected()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "No se pudo confirmar la eliminación del gasto recurrente", err)
		return
	}
	if rows == 0 {
		respondError(c, http.StatusNotFound, "No se encontró el gasto recurrente solicitado", nil)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) ApplyMonthlyExpense(c *gin.Context) {
	userID := c.GetInt64("userID")
	itemIDStr := c.Param("id")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		respondValidationError(c, "El identificador del gasto recurrente no es válido", err)
		return
	}

	var item models.MonthlyExpense
	var lastApplied sql.NullTime
	var lastExpense sql.NullInt64
	err = h.DB.QueryRow(
		`SELECT id, user_id, name, tag, amount, last_applied_at, last_applied_expense_id
		 FROM monthly_expenses
		 WHERE id=$1 AND user_id=$2`,
		itemID, userID,
	).Scan(&item.ID, &item.UserID, &item.Name, &item.Tag, &item.Amount, &lastApplied, &lastExpense)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(c, http.StatusNotFound, "No se encontró el gasto recurrente", nil)
			return
		}
		respondError(c, http.StatusInternalServerError, "No se pudo recuperar el gasto recurrente", err)
		return
	}

	now := time.Now()
	if lastApplied.Valid {
		item.LastAppliedAt = &lastApplied.Time
		if lastApplied.Time.Year() == now.Year() && lastApplied.Time.Month() == now.Month() {
			respondError(c, http.StatusBadRequest, "Este gasto recurrente ya se aplicó en el mes actual", nil)
			return
		}
	}
	if lastExpense.Valid {
		id := lastExpense.Int64
		item.LastExpenseID = &id
	}

	tx, err := h.DB.BeginTx(c, nil)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "No se pudo iniciar la operación de aplicación", err)
		return
	}
	defer tx.Rollback()

	var exp models.Expense
	expenseDate := now.Format("2006-01-02")
	err = tx.QueryRow(
		`INSERT INTO expenses (user_id, name, tag, amount, expense_date)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, user_id, name, tag, amount, expense_date`,
		userID, item.Name, item.Tag, item.Amount, expenseDate,
	).Scan(&exp.ID, &exp.UserID, &exp.Name, &exp.Tag, &exp.Amount, &exp.Date)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "No se pudo crear el gasto a partir del recurrente", err)
		return
	}

	_, err = tx.Exec(
		`UPDATE monthly_expenses SET last_applied_at=$1, last_applied_expense_id=$2 WHERE id=$3 AND user_id=$4`,
		now, exp.ID, itemID, userID,
	)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "No se pudo marcar el gasto recurrente como aplicado", err)
		return
	}

	if err := tx.Commit(); err != nil {
		respondError(c, http.StatusInternalServerError, "No se pudo confirmar la aplicación del gasto recurrente", err)
		return
	}

	item.LastAppliedAt = &now
	item.LastExpenseID = &exp.ID
	c.JSON(http.StatusCreated, gin.H{
		"expense":        exp,
		"monthlyExpense": item,
	})
}
