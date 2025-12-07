package controllers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

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
