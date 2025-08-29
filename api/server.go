package api

import (
	// middlewares "inventory-service/middlewares"
	routes "inventory-service/routes"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type Server struct {
  router *gin.Engine
  routes *routes.Route
  db *pgx.Conn
}

func NewServer(db *pgx.Conn) *Server {
  server := &Server{
    router: gin.Default(),
    db: db,
  }
  server.routes = routes.NewRoute(db)
  return server
}

func (s *Server) Run(addr string) error {
  s.router.SetTrustedProxies(nil)
  s.routes.AddRoutes(s.router)
  return s.router.Run(addr)
}
