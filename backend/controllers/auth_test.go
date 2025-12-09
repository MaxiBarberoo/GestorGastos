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
	"golang.org/x/crypto/bcrypt"
)

func TestRegister_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.POST("/register", handler.Register)

	// Expect check for existing email
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("test@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Expect insert user
	mock.ExpectQuery("INSERT INTO users").
		WithArgs("Test User", "test@example.com", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "created_at"}).
			AddRow(1, "Test User", "test@example.com", time.Now()))

	body := `{"name": "Test User", "email": "test@example.com", "password": "password123"}`
	req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "token")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRegister_EmailExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.POST("/register", handler.Register)

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("test@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	body := `{"name": "Test User", "email": "test@example.com", "password": "password123"}`
	req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "El correo ingresado ya está registrado")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLogin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.POST("/login", handler.Login)

	password := "password123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	mock.ExpectQuery("SELECT id, name, email, password_hash, created_at FROM users").
		WithArgs("test@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password_hash", "created_at"}).
			AddRow(1, "Test User", "test@example.com", string(hash), time.Now()))

	body := `{"email": "test@example.com", "password": "password123"}`
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "token")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLogin_InvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.POST("/login", handler.Login)

	// Case 1: User not found
	mock.ExpectQuery("SELECT id, name, email, password_hash, created_at FROM users").
		WithArgs("wrong@example.com").
		WillReturnError(sql.ErrNoRows)

	body := `{"email": "wrong@example.com", "password": "password123"}`
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Credenciales inválidas")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRegister_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.POST("/register", handler.Register)

	// Invalid JSON - missing required fields
	body := `{"name": "Test User"}`
	req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Los datos enviados no son válidos")
}

func TestLogin_WrongPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.POST("/login", handler.Login)

	// Hash for "correctpassword"
	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)

	mock.ExpectQuery("SELECT id, name, email, password_hash, created_at FROM users").
		WithArgs("test@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password_hash", "created_at"}).
			AddRow(1, "Test User", "test@example.com", string(hash), time.Now()))

	body := `{"email": "test@example.com", "password": "wrongpassword"}`
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Credenciales inválidas")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLogin_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.POST("/login", handler.Login)

	// Missing password
	body := `{"email": "test@example.com"}`
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Los datos de acceso no son válidos")
}

func TestMe_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.GET("/me", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.Me(c)
	})

	mock.ExpectQuery("SELECT id, name, email, created_at FROM users").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "created_at"}).
			AddRow(1, "Test User", "test@example.com", time.Now()))

	req, _ := http.NewRequest("GET", "/me", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Test User")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMe_UserNotInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.GET("/me", handler.Me) // No userID set in context

	req, _ := http.NewRequest("GET", "/me", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "No se encontró la sesión del usuario")
}

func TestMe_DBError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.GET("/me", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.Me(c)
	})

	mock.ExpectQuery("SELECT id, name, email, created_at FROM users").
		WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)

	req, _ := http.NewRequest("GET", "/me", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "No se pudo obtener el perfil del usuario")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRegister_DBErrorOnEmailCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.POST("/register", handler.Register)

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("test@example.com").
		WillReturnError(sql.ErrConnDone)

	body := `{"name": "Test User", "email": "test@example.com", "password": "password123"}`
	req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "No se pudo verificar si el email ya existe")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRegister_DBErrorOnInsert(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	handler := NewHandler(db, "secret")
	router := gin.Default()
	router.POST("/register", handler.Register)

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("test@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectQuery("INSERT INTO users").
		WithArgs("Test User", "test@example.com", sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	body := `{"name": "Test User", "email": "test@example.com", "password": "password123"}`
	req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "No se pudo guardar el usuario")
	assert.NoError(t, mock.ExpectationsWereMet())
}
