package handlers

import (
  //"fmt"
  models "inventory-service/models/sqlc"
  "log/slog"
  "net/http"
  "github.com/gin-gonic/gin"
  "github.com/jackc/pgx/v5"
)

type Handlers struct {
  db *pgx.Conn
  queries *models.Queries
  getInventoryRequest
}

func NewHandlers(db *pgx.Conn) *Handlers {
  return &Handlers{
    db:      db,
    queries: models.New(db),
  }
}
type getInventoryRequest struct {
  Name string `json:"name"`
}

func (h *Handlers) GetInventory(ctx *gin.Context) {
  _, existed := ctx.Get("claims")
  if !existed {
    ctx.JSON(http.StatusInternalServerError, gin.H{
      "error": "Claims not found in context",
    })
  }
  ctx.JSON(200, gin.H{"message": "Get Inventory"})
}

func (h *Handlers) ListInventory(ctx *gin.Context) {
  _, existed := ctx.Get("claims")
  if !existed {
    ctx.JSON(http.StatusInternalServerError, gin.H{
      "error": "Claims not found in context",
    })
  }
  inventories, err := h.queries.ListInventory(ctx, models.ListInventoryParams{
    Limit: 10,
    Offset: 0,
  })
  if err != nil {
    slog.Error("Got an error while listing inventories: ", slog.Any(err.Error(), "err"))
  }

  // for _, inventory := range inventories{
  //   fmt.Printf("%v\n", inventory)
  // }

  ctx.JSON(200, gin.H{
    "message": "List Inventory", 
    "data": inventories,
  })
}

func (h *Handlers) UpdateInventory(c *gin.Context) {
  c.JSON(200, gin.H{"message": "Update Inventory"})
}

func (h *Handlers) CreateInventory(c *gin.Context) {
  c.JSON(200, gin.H{"message": "Create Inventory"})
}

func (h *Handlers) DeleteInventory(c *gin.Context) {
  c.JSON(200, gin.H{"message": "Delete Inventory"})
}