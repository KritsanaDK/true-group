package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"tdg/internal/service"

	"github.com/labstack/echo/v4"
)

type trueHandler struct {
	BaseHandler
	service service.ITrueService
}

func NewTrueHandler(debug bool, s service.AllService) *trueHandler {
	return &trueHandler{
		BaseHandler: BaseHandler{Debug: debug},
		service:     s.ITrueService,
	}
}

func (h *trueHandler) GetUserRecommendations(c echo.Context) error {

	// Step 1️⃣ Get user_id
	userIDStr := c.Param("user_id")

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid user_id",
		})
	}

	// Step 2️⃣ Query parameter
	limitStr := c.QueryParam("limit")

	limit := 10
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error":   "invalid_parameter",
				"message": "Invalid limit parameter",
			})
		}

		limit = l
	}

	resp, err := h.service.GetRecommendations(
		userID,
		limit,
	)

	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error":   err.Error(),
				"message": fmt.Sprintf("User with ID %d does not exist", userID),
			})
		}

		if errors.Is(err, service.ErrModelUnavailable) {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{
				"error":   "model_unavailable",
				"message": "Recommendation model is temporarily unavailable",
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "internal_error",
			"message": "An unexpected error occurred",
		})
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *trueHandler) GetBatchRecommendations(c echo.Context) error {
	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil || page <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "invalid_parameter",
			"message": "page must be a positive integer",
		})
	}

	limit, err := strconv.Atoi(c.QueryParam("limit"))
	if err != nil || limit <= 0 {
		limit = 20
	} else if limit > 50 {
		limit = 50
	}

	resp, err := h.service.GetBatchRecommendations(page, limit)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "internal_error",
			"message": "An unexpected error occurred",
		})
	}

	return c.JSON(http.StatusOK, resp)
}
