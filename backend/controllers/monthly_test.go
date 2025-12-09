package controllers

import (
	"bytes"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestListMonthlyExpenses(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.GET("/monthly-expenses", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.ListMonthlyExpenses(c)
	})

	mock.ExpectQuery("SELECT id, user_id, name, tag, amount, last_applied_at, last_applied_expense_id FROM monthly_expenses").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "tag", "amount", "last_applied_at", "last_applied_expense_id"}).
			AddRow(1, 1, "Rent", "Housing", 1000.0, nil, nil))

	req, _ := http.NewRequest("GET", "/monthly-expenses", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Rent")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateMonthlyExpense(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.POST("/monthly-expenses", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.CreateMonthlyExpense(c)
	})

	mock.ExpectQuery("INSERT INTO monthly_expenses").
		WithArgs(int64(1), "Rent", "Housing", 1000.0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "tag", "amount", "last_applied_at", "last_applied_expense_id"}).
			AddRow(1, 1, "Rent", "Housing", 1000.0, nil, nil))

	body := `{"name": "Rent", "tag": "Housing", "amount": 1000.0}`
	req, _ := http.NewRequest("POST", "/monthly-expenses", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "Rent")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateMonthlyExpense_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.POST("/monthly-expenses", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.CreateMonthlyExpense(c)
	})

	// Missing required fields
	body := `{"name": "Rent"}`
	req, _ := http.NewRequest("POST", "/monthly-expenses", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Los datos del gasto recurrente no son válidos")
}

func TestDeleteMonthlyExpense_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.DELETE("/monthly-expenses/:id", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.DeleteMonthlyExpense(c)
	})

	mock.ExpectExec("DELETE FROM monthly_expenses").
		WithArgs(int64(1), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req, _ := http.NewRequest("DELETE", "/monthly-expenses/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteMonthlyExpense_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.DELETE("/monthly-expenses/:id", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.DeleteMonthlyExpense(c)
	})

	mock.ExpectExec("DELETE FROM monthly_expenses").
		WithArgs(int64(999), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	req, _ := http.NewRequest("DELETE", "/monthly-expenses/999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "No se encontró el gasto recurrente solicitado")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteMonthlyExpense_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.DELETE("/monthly-expenses/:id", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.DeleteMonthlyExpense(c)
	})

	req, _ := http.NewRequest("DELETE", "/monthly-expenses/invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "El identificador del gasto recurrente no es válido")
}

func TestApplyMonthlyExpense_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.POST("/monthly-expenses/:id/apply", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.ApplyMonthlyExpense(c)
	})

	// Never applied before (nulls)
	mock.ExpectQuery("SELECT id, user_id, name, tag, amount, last_applied_at, last_applied_expense_id FROM monthly_expenses").
		WithArgs(int64(1), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "tag", "amount", "last_applied_at", "last_applied_expense_id"}).
			AddRow(1, 1, "Rent", "Housing", 1000.0, nil, nil))

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO expenses").
		WithArgs(int64(1), "Rent", "Housing", 1000.0, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "tag", "amount", "expense_date"}).
			AddRow(10, 1, "Rent", "Housing", 1000.0, "2023-12-09"))
	mock.ExpectExec("UPDATE monthly_expenses SET last_applied_at").
		WithArgs(sqlmock.AnyArg(), int64(10), int64(1), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	req, _ := http.NewRequest("POST", "/monthly-expenses/1/apply", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "Rent")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestApplyMonthlyExpense_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.POST("/monthly-expenses/:id/apply", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.ApplyMonthlyExpense(c)
	})

	mock.ExpectQuery("SELECT id, user_id, name, tag, amount, last_applied_at, last_applied_expense_id FROM monthly_expenses").
		WithArgs(int64(999), int64(1)).
		WillReturnError(sql.ErrNoRows)

	req, _ := http.NewRequest("POST", "/monthly-expenses/999/apply", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "No se encontró el gasto recurrente")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestApplyMonthlyExpense_AlreadyAppliedThisMonth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.POST("/monthly-expenses/:id/apply", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.ApplyMonthlyExpense(c)
	})

	// Already applied this month
	now := time.Now()
	mock.ExpectQuery("SELECT id, user_id, name, tag, amount, last_applied_at, last_applied_expense_id FROM monthly_expenses").
		WithArgs(int64(1), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "tag", "amount", "last_applied_at", "last_applied_expense_id"}).
			AddRow(1, 1, "Rent", "Housing", 1000.0, now, int64(5)))

	req, _ := http.NewRequest("POST", "/monthly-expenses/1/apply", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Este gasto recurrente ya se aplicó en el mes actual")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestApplyMonthlyExpense_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.POST("/monthly-expenses/:id/apply", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.ApplyMonthlyExpense(c)
	})

	req, _ := http.NewRequest("POST", "/monthly-expenses/invalid/apply", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "El identificador del gasto recurrente no es válido")
}
