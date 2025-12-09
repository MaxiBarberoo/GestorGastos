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

func TestListExpenses(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.GET("/expenses", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.ListExpenses(c)
	})

	mock.ExpectQuery("SELECT id, user_id, name, tag, amount, expense_date FROM expenses").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "tag", "amount", "expense_date"}).
			AddRow(1, 1, "Groceries", "Food", 50.0, time.Now()))

	req, _ := http.NewRequest("GET", "/expenses", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Groceries")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateExpense(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.POST("/expenses", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.CreateExpense(c)
	})

	// Use AnyArg for the date to avoid timezone issues in test
	mock.ExpectQuery("INSERT INTO expenses").
		WithArgs(int64(1), "Groceries", "Food", 50.0, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "tag", "amount", "expense_date"}).
			AddRow(1, 1, "Groceries", "Food", 50.0, time.Now()))

	body := `{"name": "Groceries", "tag": "Food", "amount": 50.0, "date": "2023-10-27"}`
	req, _ := http.NewRequest("POST", "/expenses", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "Groceries")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteExpense(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.DELETE("/expenses/:id", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.DeleteExpense(c)
	})

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE monthly_expenses").
		WithArgs(int64(1), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected is fine
	mock.ExpectExec("DELETE FROM expenses").
		WithArgs(int64(1), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	req, _ := http.NewRequest("DELETE", "/expenses/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListExpenses_WithDateFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.GET("/expenses", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.ListExpenses(c)
	})

	mock.ExpectQuery("SELECT id, user_id, name, tag, amount, expense_date FROM expenses").
		WithArgs(int64(1), "2023-01-01", "2023-12-31").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "tag", "amount", "expense_date"}).
			AddRow(1, 1, "Groceries", "Food", 50.0, time.Now()))

	req, _ := http.NewRequest("GET", "/expenses?from=2023-01-01&to=2023-12-31", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Groceries")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListExpenses_InvalidFromDate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.GET("/expenses", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.ListExpenses(c)
	})

	req, _ := http.NewRequest("GET", "/expenses?from=invalid-date", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "El parámetro 'from' debe usar el formato YYYY-MM-DD")
}

func TestListExpenses_InvalidToDate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.GET("/expenses", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.ListExpenses(c)
	})

	req, _ := http.NewRequest("GET", "/expenses?to=bad-format", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "El parámetro 'to' debe usar el formato YYYY-MM-DD")
}

func TestCreateExpense_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.POST("/expenses", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.CreateExpense(c)
	})

	// Missing required fields
	body := `{"name": "Groceries"}`
	req, _ := http.NewRequest("POST", "/expenses", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Los datos del gasto no son válidos")
}

func TestCreateExpense_InvalidDateFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.POST("/expenses", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.CreateExpense(c)
	})

	body := `{"name": "Groceries", "tag": "Food", "amount": 50.0, "date": "27-10-2023"}`
	req, _ := http.NewRequest("POST", "/expenses", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "La fecha del gasto no tiene el formato correcto")
}

func TestDeleteExpense_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.DELETE("/expenses/:id", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.DeleteExpense(c)
	})

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE monthly_expenses").
		WithArgs(int64(1), int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM expenses").
		WithArgs(int64(999), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected = not found

	req, _ := http.NewRequest("DELETE", "/expenses/999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "No se encontró el gasto solicitado")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteExpense_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.DELETE("/expenses/:id", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.DeleteExpense(c)
	})

	req, _ := http.NewRequest("DELETE", "/expenses/invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "El identificador del gasto no es válido")
}

func TestListExpenses_DBError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.GET("/expenses", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.ListExpenses(c)
	})

	mock.ExpectQuery("SELECT id, user_id, name, tag, amount, expense_date FROM expenses").
		WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)

	req, _ := http.NewRequest("GET", "/expenses", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "No se pudo obtener la lista de gastos")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateExpense_DBError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.POST("/expenses", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.CreateExpense(c)
	})

	mock.ExpectQuery("INSERT INTO expenses").
		WithArgs(int64(1), "Groceries", "Food", 50.0, sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	body := `{"name": "Groceries", "tag": "Food", "amount": 50.0, "date": "2023-10-27"}`
	req, _ := http.NewRequest("POST", "/expenses", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "No se pudo guardar el gasto")
	assert.NoError(t, mock.ExpectationsWereMet())
}
