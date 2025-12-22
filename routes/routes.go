package routes

import (
	handlers "inventory-service/handlers"
	// "inventory-service/middlewares"
	"inventory-service/observability"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type Route struct {
	db       *pgx.Conn
	handlers *handlers.Handlers
}

func NewRoute(db *pgx.Conn, prometheusMetrics *observability.PrometheusMetrics) *Route {
	return &Route{
		db:       db,
		handlers: handlers.NewHandlers(db, prometheusMetrics),
	}
}

func (r *Route) AddHealthRoutes(router *gin.Engine) {

	router.GET("/healthz", r.handlers.HealthzHandler)
	router.GET("/readyz", r.handlers.ReadyzHandler)

}

func (r *Route) AddInventoryRoutes(router *gin.Engine) {
	v1 := router.Group("/v1")
	{
		inventory := v1.Group("/inventory")
		// inventory.Use(middlewares.ClerkAuth(r.db))
		{
			inventory.GET("/:id", r.handlers.GetInventory)
			inventory.GET("/list", r.handlers.ListInventory)
			inventory.POST("/create", r.handlers.CreateInventory)
			inventory.PUT("/:id", r.handlers.UpdateInventory)
			inventory.DELETE("/:id", r.handlers.DeleteInventory)
		}
	}
}
