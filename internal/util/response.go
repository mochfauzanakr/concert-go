package util

import (
	"errors"
	"log"
	"net/http"

	"concert-go/internal/domain/payload/response"
	"concert-go/internal/exception"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RespondOK formats a 200 response with optional pagination
func RespondOK(c *gin.Context, message string, data interface{}, pagination *response.Pagination) {
	resp := response.GlobalResponse{
		Message:    message,
		Data:       data,
		Pagination: pagination,
		ReqID:      uuid.New().String(),
		Status:     "T",
	}
	c.JSON(http.StatusOK, resp)
}

// RespondBadRequest formats a 400 response
func RespondBadRequest(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusBadRequest, response.ErrorResponse{
		Message: message,
		Data:    data,
	})
}

// RespondUnauthorized formats a 401 response
func RespondUnauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, response.ErrorResponse{
		Message: message,
	})
}

// RespondUnprocessable formats a 422 response
func RespondUnprocessable(c *gin.Context) {
	c.JSON(http.StatusUnprocessableEntity, response.ErrorResponse{
		Message: "Invalid payload",
	})
}

// RespondInternalError formats a 500 response
func RespondInternalError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, response.ErrorResponse{
		Message: "Internal server error",
	})
}

// RespondError maps AppException to its HTTP code or falls back to 500
func RespondError(c *gin.Context, err error) {
	var appErr *exception.AppException
	if errors.As(err, &appErr) {
		c.JSON(appErr.Code, response.ErrorResponse{
			Message: appErr.Message,
		})
		return
	}

	log.Printf("[Internal Error] %v", err)
	RespondInternalError(c)
}
