package routes

import (
	"github.com/gin-gonic/gin"

	"gestor-gastos/config"
	"gestor-gastos/controllers"
	"gestor-gastos/middleware"
)

func Setup(cfg *config.Config, handler *controllers.Handler) *gin.Engine {
	router := gin.Default()
	router.Use(middleware.CORS(cfg.FrontendOrigin))

	api := router.Group("/api")
	auth := api.Group("/auth")
	auth.POST("/register", handler.Register)
	auth.POST("/login", handler.Login)
	auth.GET("/me", middleware.Auth(handler.JWTSecret), handler.Me)

	protected := api.Group("/")
	protected.Use(middleware.Auth(handler.JWTSecret))
	{
		protected.GET("/expenses", handler.ListExpenses)
		protected.POST("/expenses", handler.CreateExpense)
		protected.DELETE("/expenses/:id", handler.DeleteExpense)

		protected.GET("/monthly-expenses", handler.ListMonthlyExpenses)
		protected.POST("/monthly-expenses", handler.CreateMonthlyExpense)
		protected.DELETE("/monthly-expenses/:id", handler.DeleteMonthlyExpense)
		protected.POST("/monthly-expenses/:id/apply", handler.ApplyMonthlyExpense)
	}

	return router
}
