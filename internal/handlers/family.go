package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hnfnfl/family-tree/internal/repository"
)

type FamilyHandler struct {
	repo *repository.FamilyRepository
}

func NewFamilyHandler(repo *repository.FamilyRepository) *FamilyHandler {
	return &FamilyHandler{repo: repo}
}

func (h *FamilyHandler) GetAll(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	families, err := h.repo.GetAll(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   families,
		"limit":  limit,
		"offset": offset,
		"total":  len(families),
	})
}

func (h *FamilyHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Family ID is required"})
		return
	}

	family, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Family not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": family})
}

func (h *FamilyHandler) Create(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	type CreateFamilyRequest struct {
		Name        string  `json:"name" binding:"required"`
		Description *string `json:"description,omitempty"`
	}

	var req CreateFamilyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	family := repository.Family{
		Name:        req.Name,
		Description: req.Description,
	}

	created, err := h.repo.Create(c.Request.Context(), &family, fmt.Sprintf("%v", userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": created})
}

func (h *FamilyHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Family ID is required"})
		return
	}

	var updates repository.Family
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.repo.Update(c.Request.Context(), id, &updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": updated})
}

func (h *FamilyHandler) GetFamilyTree(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Family ID is required"})
		return
	}

	tree, err := h.repo.GetFamilyTree(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Family tree not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tree})
}

type AddMemberRequest struct {
	PersonID string `json:"person_id" binding:"required"`
}

func (h *FamilyHandler) AddMember(c *gin.Context) {
	familyID := c.Param("id")
	if familyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Family ID is required"})
		return
	}

	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.AddMember(c.Request.Context(), familyID, req.PersonID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Member added to family successfully",
		"family_id": familyID,
		"person_id": req.PersonID,
	})
}
