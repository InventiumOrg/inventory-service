package routes

import (
	"inventory-service/handlers"
	"inventory-service/middlewares"

	"github.com/gin-gonic/gin"
)

type Route struct {
  handlers *handlers.Handlers
}

func (r *Route) AddRoutes(router *gin.Engine) *gin.Engine {
	router.Use(middlewares.Authenticate())
	router.GET("/inventory/:id", r.handlers.GetInventory)
  router.GET("/inventory", r.handlers.ListInventory)
  router.PUT("/inventory/:id", r.handlers.UpdateInventory)
	router.POST("/inventory", r.handlers.CreateInventory)
  router.DELETE("/inventory/:id", r.handlers.DeleteInventory)
  return router
}
