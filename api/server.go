package api

import (
  routes "inventory-service/routes"
  "github.com/gin-gonic/gin"
)

type Server struct {
  router *gin.Engine
  route *routes.Route
}

func NewServer() *Server {
  return &Server{
    router: gin.Default(),
  }
}

func (s *Server) Run(addr string) error {
  s.router.SetTrustedProxies(nil)
  s.route = &routes.Route{}
  s.route.AddRoutes(s.router)
  return s.router.Run(addr)
}
