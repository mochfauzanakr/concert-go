package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"concert-go/internal/domain/model"
	"concert-go/internal/middleware"
	"concert-go/internal/repository/mock"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestRequireRoleMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	roleRepo := mock.NewRoleRepositoryMock()

	setupRouter := func() *gin.Engine {
		r := gin.New()
		r.GET("/admin-only", middleware.RequireRole(roleRepo, "SUPER_ADMIN", "ADMIN"), func(c *gin.Context) {
			c.String(http.StatusOK, "welcome admin")
		})
		return r
	}

	t.Run("no session returns 401", func(t *testing.T) {
		r := setupRouter()
		req := httptest.NewRequest(http.MethodGet, "/admin-only", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("nil role returns 403", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set(middleware.UserSessionContextKey, model.UserSession{
				UserID: uuid.New(),
				Email:  "norole@example.com",
				RoleID: nil,
			})
		})
		r.GET("/admin-only", middleware.RequireRole(roleRepo, "ADMIN"), func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/admin-only", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})

	t.Run("insufficient role returns 403", func(t *testing.T) {
		r := gin.New()
		buyer, _ := roleRepo.FindByName(context.Background(), "BUYER")
		r.Use(func(c *gin.Context) {
			c.Set(middleware.UserSessionContextKey, model.UserSession{
				UserID: uuid.New(),
				Email:  "buyer@example.com",
				RoleID: &buyer.ID,
			})
		})
		r.GET("/admin-only", middleware.RequireRole(roleRepo, "ADMIN"), func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/admin-only", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})

	t.Run("matching role returns 200", func(t *testing.T) {
		r := gin.New()
		admin, _ := roleRepo.FindByName(context.Background(), "ADMIN")
		r.Use(func(c *gin.Context) {
			c.Set(middleware.UserSessionContextKey, model.UserSession{
				UserID: uuid.New(),
				Email:  "admin@example.com",
				RoleID: &admin.ID,
			})
		})
		r.GET("/admin-only", middleware.RequireRole(roleRepo, "SUPER_ADMIN", "ADMIN"), func(c *gin.Context) {
			c.String(http.StatusOK, "welcome admin")
		})

		req := httptest.NewRequest(http.MethodGet, "/admin-only", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})
}
