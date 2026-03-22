package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type FamilyHandler struct {
	repo interface{} // TODO: Create FamilyRepository
}

func NewFamilyHandler(repo interface{}) *FamilyHandler {
	return &FamilyHandler{repo: repo}
}

func (h *FamilyHandler) GetAll(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

func (h *FamilyHandler) GetByID(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

func (h *FamilyHandler) Create(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

func (h *FamilyHandler) Update(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

func (h *FamilyHandler) GetFamilyTree(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}
