package handler

import (
	"ecommerce/user-service/config"
	"ecommerce/user-service/internal/adapter"
	"ecommerce/user-service/internal/adapter/handler/request"
	"ecommerce/user-service/internal/adapter/handler/response"
	"ecommerce/user-service/internal/core/domain/entity"
	"ecommerce/user-service/internal/core/service"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type RoleHandlerInterface interface {
	GetAll(c echo.Context) error
	GetByID(c echo.Context) error
	Create(c echo.Context) error
	Update(c echo.Context) error
	Delete(c echo.Context) error
}

type roleHandler struct {
	roleService service.RoleServiceInterface
}

func (r *roleHandler) GetAll(c echo.Context) error {
	var (
		respRole = []response.RoleResponse{}
		resp     = response.DefaultResponse{}
		ctx      = c.Request().Context()
	)

	search := c.QueryParam("search")
	roles, err := r.roleService.GetAll(ctx, search)
	if err != nil {
		log.Errorf("[Rolehandler-3] GetAll: %v", err)
		if err.Error() == "404" {
			resp.Message = "Role not found"
			return c.JSON(http.StatusNotFound, resp)
		}
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	for _, role := range roles {
		respRole = append(respRole, response.RoleResponse{
			ID:   role.ID,
			Name: role.Name,
		})
	}

	resp.Message = "success"
	resp.Data = respRole
	return c.JSON(http.StatusOK, resp)
}

func (r *roleHandler) GetByID(c echo.Context) error {
	var (
		respRole = response.RoleResponse{}
		resp     = response.DefaultResponse{}
		ctx      = c.Request().Context()
	)

	roleIDString := c.Param("id")
	if roleIDString == "" {
		log.Errorf("[Rolehandler-1] GetByID: %v", "missing or invalid role ID")
		resp.Message = "missing or invalid role ID"
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	roleID, err := strconv.Atoi(roleIDString)
	if err != nil || roleID < 1 {
		log.Errorf("[Rolehandler-2] GetByID: %v", err)
		resp.Message = "invalid role ID"
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	role, err := r.roleService.GetByID(ctx, int64(roleID))
	if err != nil {
		log.Errorf("[RoleHandler-3] GetByID: %v", err)
		if err.Error() == "404" {
			resp.Message = "Role not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}
		resp.Message = err.Error()
		return c.JSON(http.StatusInternalServerError, resp)
	}

	respRole.ID = role.ID
	respRole.Name = role.Name
	resp.Message = "Success"
	resp.Data = respRole
	return c.JSON(http.StatusOK, resp)
}

func (r *roleHandler) Create(c echo.Context) error {
	var (
		resp        = response.DefaultResponse{}
		req         = request.RoleRequest{}
		ctx         = c.Request().Context()
		jwtUserData = entity.JwtUserData{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[RoleHandler-1] Create: %s", "user not found")
		resp.Message = "data token not found"
		resp.Data = nil
		return c.JSON(http.StatusNotFound, resp)
	}

	err := json.Unmarshal([]byte(user), &jwtUserData)
	if err != nil {
		log.Errorf("[RoleHandler-2] Create: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	if jwtUserData.RoleName != "Super Admin" {
		log.Errorf("[RoleHandler-1 Create] %v", "online Super Admin can create role")
		resp.Message = "online Super Admin can create role"
		resp.Data = nil
		return c.JSON(http.StatusForbidden, resp)
	}

	if err := c.Bind(&req); err != nil {
		log.Errorf("[Rolehandler-3] Create: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	if err := c.Validate(req); err != nil {
		log.Errorf("[Rolehandler-4] Create: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	roleEntity := entity.RoleEntity{
		Name: req.Name,
	}

	err = r.roleService.Create(ctx, roleEntity)
	if err != nil {
		log.Errorf("[Rolehandler-5] Create: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp.Message = "Role created successfully"
	resp.Data = nil
	return c.JSON(http.StatusCreated, resp)
}

func (r *roleHandler) Update(c echo.Context) error {
	var (
		resp = response.DefaultResponse{}
		req  = request.RoleRequest{}
		ctx  = c.Request().Context()
	)

	roleIDString := c.Param("id")
	if roleIDString == "" {
		log.Errorf("[Rolehandler-1] Update: %v", "missing or invalid role ID")
		resp.Message = "missing or invalid role ID"
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	roleID, err := strconv.Atoi(roleIDString)
	if err != nil || roleID < 1 {
		log.Errorf("[Rolehandler-2] Update: %v", err)
		resp.Message = "invalid role ID"
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	if err := c.Bind(&req); err != nil {
		log.Errorf("[Rolehandler-3] Update: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	if err := c.Validate(req); err != nil {
		log.Errorf("[Rolehandler-4] Update: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	reqEntity := entity.RoleEntity{
		ID:   int64(roleID),
		Name: req.Name,
	}

	err = r.roleService.Update(ctx, reqEntity)
	if err != nil {
		log.Errorf("[Rolehandler-3] Update: %v", err)
		if err.Error() == "404" {
			resp.Message = "Role not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}

		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp.Message = "Role updated successfully"
	resp.Data = nil
	return c.JSON(http.StatusOK, resp)
}

func (r *roleHandler) Delete(c echo.Context) error {
	var (
		resp = response.DefaultResponse{}
		ctx  = c.Request().Context()
	)

	roleIDString := c.Param("id")
	if roleIDString == "" {
		log.Infof("[Rolehandler-1] Delete: %s", "missing or invalid role ID")
		resp.Message = "missing or invalid role ID"
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	roleID, err := strconv.Atoi(roleIDString)
	if err != nil || roleID < 1 {
		log.Errorf("[Rolehandler-2] Delete: %v", err)
		resp.Message = "invalid role ID"
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	err = r.roleService.Delete(ctx, int64(roleID))
	if err != nil {
		log.Errorf("[Rolehandler-3] Delete: %v", err)
		if err.Error() == "404" {
			resp.Message = "Role not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}
		if err.Error() == "400" {
			resp.Message = "Role is associated with users"
			return c.JSON(http.StatusBadRequest, resp)
		}

		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp.Message = "Role deleted successfully"
	resp.Data = nil
	return c.JSON(http.StatusOK, resp)
}

func NewRoleHandler(e *echo.Echo, roleService service.RoleServiceInterface, cfg *config.Config, jwtService service.JwtServiceInterface) RoleHandlerInterface {
	role := &roleHandler{roleService: roleService}

	e.Use(middleware.Recover())
	mid := adapter.NewMiddlewareAdapter(cfg, jwtService)

	// Both Admin and Super Admin can read role data. Mutations remain restricted below.
	adminGroup := e.Group("/admin", mid.CheckToken())
	adminGroup.GET("/roles", role.GetAll)
	adminGroup.GET("/roles/:id", role.GetByID)

	superAdminGroup := e.Group("/admin", mid.CheckToken(), mid.RequireRole("Super Admin"))
	superAdminGroup.POST("/roles", role.Create)
	superAdminGroup.PUT("/roles/:id", role.Update)
	superAdminGroup.DELETE("/roles/:id", role.Delete)

	return role
}
