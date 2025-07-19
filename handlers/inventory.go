package handlers

import (
	"github.com/gin-gonic/gin"
)
type Handlers struct{
	getInventoryRequest
}
type getInventoryRequest struct {
	Name string `json:"name"`
}

func (h *Handlers) GetInventory(ctx *gin.Context) {
	ctx.JSON(200, gin.H{"message": "Get Inventory"})
}

func (h *Handlers) ListInventory(ctx *gin.Context) {
	ctx.JSON(200, gin.H{"message": "List Inventory"})
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