package adapter

import (
	"ecommerce/user-service/config"
	"ecommerce/user-service/internal/adapter/handler/response"
	"ecommerce/user-service/internal/core/domain/entity"
	"ecommerce/user-service/internal/core/service"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type MiddlewareAdapterInterface interface {
	CheckToken() echo.MiddlewareFunc
	RequireRole(roleName string) echo.MiddlewareFunc
}

// RequireRole authorizes a request after CheckToken has loaded its Redis session.
func (m *middlewareAdapter) RequireRole(roleName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			resp := response.DefaultResponse{}
			sessionJSON, ok := c.Get("user").(string)
			if !ok || sessionJSON == "" {
				resp.Message = "missing or invalid token"
				return c.JSON(http.StatusUnauthorized, resp)
			}

			session := entity.JwtUserData{}
			if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
				log.Errorf("[Middleware-4] RequireRole: %v", err)
				resp.Message = "invalid session data"
				return c.JSON(http.StatusUnauthorized, resp)
			}

			if session.RoleName != roleName {
				resp.Message = "only user admin to can access"
				return c.JSON(http.StatusForbidden, resp)
			}

			return next(c)
		}
	}
}

type middlewareAdapter struct {
	cfg        *config.Config
	jwtService service.JwtServiceInterface
}

// CheckToken implements [MiddlewareAdapterInterface].
func (m *middlewareAdapter) CheckToken() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			respErr := response.DefaultResponse{}
			redisConn := config.NewConfig().NewRedisClient()
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				log.Errorf("[MiddlewareAdapter-1] CheckToken: %s", "missing or invalid token")
				respErr.Message = "missing or invalid token"
				respErr.Data = nil
				return c.JSON(http.StatusUnauthorized, respErr)
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			_, err := m.jwtService.ValidateToken(tokenString)
			if err != nil {
				log.Errorf("[MiddlewareAdapter-2] CheckToken: %s", err.Error())
				respErr.Message = err.Error()
				respErr.Data = nil
				return c.JSON(http.StatusUnauthorized, respErr)
			}

			getSession, err := redisConn.Get(c.Request().Context(), tokenString).Result()
			if err != nil || len(getSession) == 0 {
				log.Errorf("[MiddlewareAdapter-3] CheckToken: %s", err.Error())
				respErr.Message = err.Error()
				respErr.Data = nil
				return c.JSON(http.StatusUnauthorized, respErr)
			}

			jwtUserData := entity.JwtUserData{}
			err = json.Unmarshal([]byte(getSession), &jwtUserData)
			if err != nil {
				log.Errorf("[MiddlewareAdapter-4] CheckToken: %v", err)
				respErr.Message = err.Error()
				respErr.Data = nil
				return c.JSON(http.StatusInternalServerError, respErr)
			}

			path := c.Request().URL.Path
			segments := strings.Split(strings.Trim(path, "/"), "/")
			if jwtUserData.RoleName == "Customer" && segments[0] == "admin" {
				log.Infof("[MiddlewareAdapter-5] CheckToken: %s", "customer cannot access admin routes")
				respErr.Message = "customer cannot access admin routes"
				respErr.Data = nil
				return c.JSON(http.StatusForbidden, respErr)
			}

			c.Set("user", getSession)
			return next(c)
		}
	}
}

func NewMiddlewareAdapter(cfg *config.Config, jwtService service.JwtServiceInterface) MiddlewareAdapterInterface {
	return &middlewareAdapter{
		cfg:        cfg,
		jwtService: jwtService,
	}
}
