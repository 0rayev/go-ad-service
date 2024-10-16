/*
This file contains the HTTP handlers that manage incoming requests and responses.
It acts as the entry point for the API, where requests are processed and
appropriate responses are sent back to the client.
*/
package ad

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// Handler struct holds a reference to the AdService
type Handler struct {
	Service *AdService
}

// NewHandler is a constructor for Handler
func NewHandler(service *AdService) *Handler {
	return &Handler{Service: service}
}

// GetAdByID handles fetching a single ad by its ID, with tracing
func (h *Handler) GetAdByID(c *gin.Context) {
	// Start a span for the handler
	tracer := otel.Tracer("ad-service.handler")
	ctx, span := tracer.Start(c.Request.Context(), "GetAdByIDHandler")
	defer span.End()

	// Parse the ID from the URL parameter and handle errors
	id, err := strconv.Atoi(c.Param("id"))
	// Check for non-numeric or non-positive IDs
	if err != nil || id <= 0 {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", "Invalid ID parameter"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Fetch the ad using the service layer, passing the trace context
	ad, err := h.Service.GetAdByID(id, ctx)
	if err != nil {
		// Check if the error is due to "not found" or an internal issue
		if err == sql.ErrNoRows {
			// Handle case where the ad is not found
			span.RecordError(err)
			span.SetAttributes(attribute.Int("ad_id", id), attribute.String("error", "Ad not found"))
			c.JSON(http.StatusNotFound, gin.H{"error": "Ad not found"})
		} else {
			// Handle internal server errors
			span.RecordError(err)
			span.SetAttributes(attribute.String("error", "Failed to fetch ad by ID"))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch ad by ID"})
		}
		return
	}

	span.SetAttributes(attribute.Int("ad_id", ad.ID), attribute.String("status", "success"))
	c.JSON(http.StatusOK, ad)
}

// AddAd handles the creation of a new ad, with tracing
func (h *Handler) AddAd(c *gin.Context) {
	// Start a span for the handler
	tracer := otel.Tracer("ad-service.handler")
	ctx, span := tracer.Start(c.Request.Context(), "AddAdHandler")
	defer span.End()

	var ad Ad
	if err := c.ShouldBindJSON(&ad); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", "Invalid request body"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate title and description
	if ad.Title == "" || ad.Description == "" {
		span.RecordError(errors.New("title or description cannot be empty"))
		span.SetAttributes(attribute.String("error", "Title or description missing"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title and description are required"})
		return
	}

	// Validate price (have to be positive)
	if ad.Price <= 0 {
		span.RecordError(errors.New("invalid price value"))
		span.SetAttributes(attribute.String("error", "Price cannot be zero or negative"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Price cannot be zero or negative"})
		return
	}

	if err := h.Service.AddAd(&ad, ctx); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", "Failed to add ad"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add ad"})
		return
	}

	span.SetAttributes(attribute.Int("ad_id", ad.ID), attribute.String("status", "success"))
	c.JSON(http.StatusCreated, ad)
}

// GetAllAds handles fetching all ads, with tracing
// Expected URL: http://localhost:8080/ads?page=1&limit=10&sort_by=created_at&order=asc
func (h *Handler) GetAllAds(c *gin.Context) {

	// Start a span for the handler
	tracer := otel.Tracer("ad-service.handler")
	ctx, span := tracer.Start(c.Request.Context(), "GetAllAdsHandler")
	defer span.End()

	// Paginating and sorting
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page value. Must be a positive integer."})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit <= 0 {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit value. Must be a positive integer."})
		return
	}

	sortBy := c.DefaultQuery("sort_by", "created_at")
	// Validate if sortBy is one of the allowed fields
	validSortFields := map[string]bool{
		"id":         true,
		"title":      true,
		"price":      true,
		"created_at": true,
		"is_active":  true,
	}
	if !validSortFields[sortBy] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sort_by value. Must be one of 'id', 'title', 'price', 'created_at', 'is_active'."})
		return
	}

	order := c.DefaultQuery("order", "asc")
	if order != "asc" && order != "desc" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order value. Must be either 'asc' or 'desc'."})
		return
	}
	// Fetch ads from the service using the validated parameters
	ads, err := h.Service.GetAllAds(page, limit, sortBy, order, ctx)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", "Failed to fetch ads"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch ads"})
		return
	}

	span.SetAttributes(attribute.String("status", "success"))
	c.JSON(http.StatusOK, ads)
}

// UpdateAd handles updating an existing ad, with tracing
func (h *Handler) UpdateAd(c *gin.Context) {
	// Start a span for the handler
	tracer := otel.Tracer("ad-service.handler")
	ctx, span := tracer.Start(c.Request.Context(), "UpdateAdHandler")
	defer span.End()

	// Validate ID
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", "Invalid ad ID"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ad ID"})
		return
	}

	var ad Ad
	if err := c.ShouldBindJSON(&ad); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", "Invalid request body"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate title and description
	if ad.Title == "" || ad.Description == "" {
		span.RecordError(errors.New("title or description cannot be empty"))
		span.SetAttributes(attribute.String("error", "Title or description missing"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title and description are required"})
		return
	}

	// Validate price (have to be positive)
	if ad.Price <= 0 {
		span.RecordError(errors.New("invalid price value"))
		span.SetAttributes(attribute.String("error", "Price cannot be zero or negative"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Price cannot be zero or negative"})
		return
	}
	err = h.Service.UpdateAd(id, &ad, ctx)
	if err != nil {
		if errors.Is(err, ErrAdNotFound) {
			span.RecordError(err)
			span.SetAttributes(attribute.Int("ad_id", id), attribute.String("error", "Ad not found"))
			c.JSON(http.StatusNotFound, gin.H{"error": "Ad not found"})
			return
		}
		span.RecordError(err)
		span.SetAttributes(attribute.Int("ad_id", id), attribute.String("error", "Failed to update ad"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update ad"})
		return
	}

	span.SetAttributes(attribute.Int("ad_id", id), attribute.String("status", "success"))
	c.JSON(http.StatusOK, gin.H{"message": "Ad updated"})
}

// DeleteAd handles deleting an ad by ID, with tracing
func (h *Handler) DeleteAd(c *gin.Context) {
	// Start a span for the handler
	tracer := otel.Tracer("ad-service.handler")
	ctx, span := tracer.Start(c.Request.Context(), "DeleteAdHandler")
	defer span.End()

	id, err := strconv.Atoi(c.Param("id"))
	// Check for non-numeric or non-positive IDs
	if err != nil || id <= 0 {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", "Invalid ad ID"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ad ID"})
		return
	}
	err = h.Service.DeleteAd(id, ctx)
	if err != nil {
		if errors.Is(err, ErrAdNotFound) {
			span.RecordError(err)
			span.SetAttributes(attribute.Int("ad_id", id), attribute.String("error", "Ad not found"))
			c.JSON(http.StatusNotFound, gin.H{"error": "Ad not found"})
			return
		}
		span.RecordError(err)
		span.SetAttributes(attribute.Int("ad_id", id), attribute.String("error", "Failed to delete ad"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete ad"})
		return
	}

	span.SetAttributes(attribute.Int("ad_id", id), attribute.String("status", "success"))
	c.JSON(http.StatusOK, gin.H{"message": "Ad deleted"})
}
