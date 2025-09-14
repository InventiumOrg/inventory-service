package api

import (
	routes "inventory-service/routes"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
  "github.com/gin-contrib/cors"
  "time"
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
  s.router.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"http://localhost:3000"},
    AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
    AllowHeaders:     []string{"Origins", "Content-Type", "Authorization", "Bearer"},
    AllowCredentials: true,
    MaxAge:           12 * time.Hour,
  }))
  s.routes.AddRoutes(s.router)
  return s.router.Run(addr)
}
