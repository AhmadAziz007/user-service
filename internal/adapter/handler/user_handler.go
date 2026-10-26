package handler

import (
	"ecommerce/user-service/config"
	"ecommerce/user-service/internal/adapter"
	"ecommerce/user-service/internal/adapter/handler/request"
	"ecommerce/user-service/internal/adapter/handler/response"
	"ecommerce/user-service/internal/core/domain/entity"
	"ecommerce/user-service/internal/core/service"
	"ecommerce/user-service/utils/conv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type UserHandlerInterface interface {
	SignIn(c echo.Context) error
	CreateUserAccount(c echo.Context) error
	ForgotPassword(c echo.Context) error
	VerifyAccount(c echo.Context) error
	UpdatePassword(c echo.Context) error
	GetProfileUser(c echo.Context) error
	UpdateDataUser(c echo.Context) error

	//Modul Customers Admin
	GetCustomerAll(c echo.Context) error
	GetCustomerByID(c echo.Context) error
	GetCustomerForAuth(c echo.Context) error
	CreateCustomer(c echo.Context) error
	UpdateCustomer(c echo.Context) error
	DeleteCustomer(c echo.Context) error
}

type userHandler struct {
	userService service.UserServiceInterface
}

// DeleteCustomer implements [UserHandlerInterface].
func (u *userHandler) DeleteCustomer(c echo.Context) error {
	var (
		resp = response.DefaultResponseWithPaginations{}
		ctx  = c.Request().Context()
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[UserHandler-1] DeleteCustomer: %s", "data token not found")
		resp.Message = "data token not valid"
		resp.Data = nil
		return c.JSON(http.StatusUnauthorized, resp)
	}
	jwtUserData := entity.JwtUserData{}
	if err := json.Unmarshal([]byte(user), &jwtUserData); err != nil {
		resp.Message = "invalid session data"
		resp.Data = nil
		return c.JSON(http.StatusUnauthorized, resp)
	}

	idParamStr := c.Param("id")
	if idParamStr == "" {
		log.Infof("[UserHandler-2] DeleteCustomer: %s", "missing or invalid customer ID")
		resp.Message = "missing or invalid customer ID"
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	id, err := conv.StringToInt64(idParamStr)
	if err != nil {
		log.Infof("[UserHandler-3] DeleteCustomer: %s", "invalid customer ID")
		resp.Message = "invalid customer ID"
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	err = u.userService.DeleteCustomer(ctx, id)
	if err != nil {
		log.Infof("[UserHandler-4] DeleteCustomer: %v", err)
		if err.Error() == "404" {
			resp.Message = "Customer not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp.Message = "Customer deleted successfully"
	resp.Data = nil
	return c.JSON(http.StatusOK, resp)
}

// UpdateCustomer implements [UserHandlerInterface].
func (u *userHandler) UpdateCustomer(c echo.Context) error {
	var (
		resp = response.DefaultResponseWithPaginations{}
		ctx  = c.Request().Context()
		req  = request.CustomerRequest{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[UserHandler-1] UpdateCustomer: %s", "data token not found")
		resp.Message = "data token not valid"
		resp.Data = nil
		return c.JSON(http.StatusUnauthorized, resp)
	}
	jwtUserData := entity.JwtUserData{}
	if err := json.Unmarshal([]byte(user), &jwtUserData); err != nil {
		resp.Message = "invalid session data"
		resp.Data = nil
		return c.JSON(http.StatusUnauthorized, resp)
	}

	if err = c.Bind(&req); err != nil {
		log.Errorf("[UserHandler-2] UpdateCustomer: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	if err = c.Validate(&req); err != nil {
		log.Errorf("[UserHandler-3] UpdateCustomer: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	latString := ""
	lngString := ""
	if req.Lat != 0 {
		latString = strconv.FormatFloat(req.Lat, 'g', -1, 64)
	}

	if req.Lng != 0 {
		lngString = strconv.FormatFloat(req.Lng, 'g', -1, 64)
	}
	phoneString := fmt.Sprintf("%v", req.Phone)

	idParamStr := c.Param("id")
	if idParamStr == "" {
		log.Infof("[UserHandler-4] UpdateCustomer: %s", "missing or invalid customer ID")
		resp.Message = "missing or invalid customer ID"
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	id, err := conv.StringToInt64(idParamStr)
	if err != nil {
		log.Infof("[UserHandler-5] UpdateCustomer: %s", "invalid customer ID")
		resp.Message = "invalid customer ID"
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	reqEntity := entity.UserEntity{
		ID:        id,
		Name:      req.Name,
		Email:     req.Email,
		Password:  req.Password,
		Phone:     phoneString,
		Address:   req.Address,
		Lat:       latString,
		Lng:       lngString,
		Photo:     req.Photo,
		UpdatedBy: &jwtUserData.UserID,
	}

	err = u.userService.UpdateCustomer(ctx, reqEntity)
	if err != nil {
		log.Errorf("[UserHandler-6] UpdateCustomer: %v", err)
		if err.Error() == "404" {
			resp.Message = "Customer not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp.Message = "Success"
	resp.Data = nil

	return c.JSON(http.StatusOK, resp)
}

// CreateCustomer implements [UserHandlerInterface].
func (u *userHandler) CreateCustomer(c echo.Context) error {
	var (
		resp = response.DefaultResponseWithPaginations{}
		ctx  = c.Request().Context()
		req  = request.CustomerRequest{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[UserHandler-1] CreateCustomer: %s", "data token not found")
		resp.Message = "data token not valid"
		resp.Data = nil
		return c.JSON(http.StatusUnauthorized, resp)
	}
	jwtUserData := entity.JwtUserData{}
	if err := json.Unmarshal([]byte(user), &jwtUserData); err != nil {
		resp.Message = "invalid session data"
		resp.Data = nil
		return c.JSON(http.StatusUnauthorized, resp)
	}

	if err = c.Bind(&req); err != nil {
		log.Errorf("[UserHandler-2] CreateCustomer: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	if err = c.Validate(&req); err != nil {
		log.Errorf("[UserHandler-3] CreateCustomer: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	if req.Password != req.PasswordConfirmation {
		log.Infof("[UserHandler-4] CreateCustomer: %s", "password and confirm password does not match")
		resp.Message = "password and confirm password does not match"
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	latString := strconv.FormatFloat(req.Lat, 'g', -1, 64)
	lngString := strconv.FormatFloat(req.Lng, 'g', -1, 64)

	reqEntity := entity.UserEntity{
		Name:      req.Name,
		Email:     req.Email,
		Password:  req.Password,
		Phone:     req.Phone,
		Address:   req.Address,
		Lat:       latString,
		Lng:       lngString,
		Photo:     req.Photo,
		RoleID:    req.RoleID,
		CreatedBy: &jwtUserData.UserID,
	}

	err = u.userService.CreateCustomer(ctx, reqEntity)
	if err != nil {
		log.Fatalf("[UserHandler-5] CreateCustomer: %v", err)
		resp.Message = "failed to create customer"
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp.Message = "success"
	resp.Data = nil
	resp.Pagination = nil

	return c.JSON(http.StatusCreated, resp)
}

func (u *userHandler) GetCustomerByID(c echo.Context) error {
	var (
		resp     = response.DefaultResponseWithPaginations{}
		ctx      = c.Request().Context()
		respUser = response.CustomerResponse{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[UserHandler-1] GetCustomerByID: %s", "data token not found")
		resp.Message = "data token not valid"
		resp.Data = nil
		return c.JSON(http.StatusUnauthorized, resp)
	}

	idParam := c.Param("id")
	if idParam == "" {
		log.Errorf("[UserHandler-2] GetCustomerByID: %s", "id invalid")
		resp.Message = "id invalid"
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Errorf("[UserHandler-3] GetCustomerByID: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	result, err := u.userService.GetCustomerByID(ctx, id)
	if err != nil {
		log.Errorf("[UserHandler-4] GetCustomerByID: %v", err)
		if err.Error() == "404" {
			resp.Message = "Customer not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp.Message = "success get customer by id"
	respUser.ID = result.ID
	respUser.RoleID = result.RoleID
	respUser.Name = result.Name
	respUser.Email = result.Email
	respUser.Phone = result.Phone
	respUser.Address = result.Address
	respUser.Photo = result.Photo
	respUser.CreatedBy = result.CreatedBy
	respUser.UpdatedBy = result.UpdatedBy
	respUser.CreatedAt = result.CreatedAt.Format("2006-01-02 15:04:05")
	respUser.UpdatedAt = result.UpdatedAt.Format("2006-01-02 15:04:05")
	respUser.Lat = result.Lat
	respUser.Lng = result.Lng

	resp.Data = respUser
	resp.Pagination = nil

	return c.JSON(http.StatusOK, resp)
}

// GetCustomerForAuth returns the requested customer while preventing a
// customer token from reading another customer's profile.
func (u *userHandler) GetCustomerForAuth(c echo.Context) error {
	sessionJSON, ok := c.Get("user").(string)
	if !ok || sessionJSON == "" {
		return c.JSON(http.StatusUnauthorized, response.DefaultResponse{Message: "missing or invalid token", Data: nil})
	}

	session := entity.JwtUserData{}
	if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
		return c.JSON(http.StatusUnauthorized, response.DefaultResponse{Message: "invalid session data", Data: nil})
	}

	requestedID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.DefaultResponse{Message: "id invalid", Data: nil})
	}
	if session.RoleName == "Customer" && session.UserID != requestedID {
		return c.JSON(http.StatusForbidden, response.DefaultResponse{Message: "customer cannot access another customer's profile", Data: nil})
	}

	return u.GetCustomerByID(c)
}

// GetCustomerAll implements UserHandlerInterface.
func (u *userHandler) GetCustomerAll(c echo.Context) error {
	var (
		resp     = response.DefaultResponseWithPaginations{}
		ctx      = c.Request().Context()
		respUser = []response.CustomerListResponse{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[UserHandler-1] GetCustomerAll: %s", "data token not found")
		resp.Message = "data token not found"
		resp.Data = nil
		return c.JSON(http.StatusNotFound, resp)
	}

	search := c.QueryParam("search")
	orderBy := "created_at"
	if c.QueryParam("order_by") != "" {
		orderBy = c.QueryParam("order_by")
	}

	orderType := c.QueryParam("order_type")
	if orderType != "asc" && orderType != "desc" {
		orderType = "desc"
	}

	pageStr := c.QueryParam("page")
	var page int64 = 1
	if pageStr != "" {
		page, _ = conv.StringToInt64(pageStr)
		if page <= 0 {
			page = 1
		}
	}

	limitStr := c.QueryParam("limit")
	var limit int64 = 10
	if limitStr != "" {
		limit, _ = conv.StringToInt64(limitStr)
		if limit <= 0 {
			limit = 10
		}
	}

	reqEntity := entity.QueryStringCustomer{
		Search:    search,
		Page:      page,
		Limit:     limit,
		OrderBy:   orderBy,
		OrderType: orderType,
	}

	results, countData, totalPages, err := u.userService.GetCustomerAll(ctx, reqEntity)
	if err != nil {
		log.Errorf("[UserHandler-2] GetCustomerAll: %v", err)
		if err.Error() == "404" {
			resp.Message = "Data not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	for _, val := range results {
		respUser = append(respUser, response.CustomerListResponse{
			ID:        val.ID,
			Name:      val.Name,
			Email:     val.Email,
			Photo:     val.Photo,
			Phone:     val.Phone,
			CreatedBy: val.CreatedBy,
			UpdatedBy: val.UpdatedBy,
			CreatedAt: val.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: val.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	resp.Message = "Data retrieved successfully"
	resp.Data = respUser
	resp.Pagination = &response.Pagination{
		Page:       page,
		TotalCount: countData,
		PerPage:    limit,
		TotalPage:  totalPages,
	}

	return c.JSON(http.StatusOK, resp)
}

// UpdateDataUser implements UserHandlerInterface.
func (u *userHandler) UpdateDataUser(c echo.Context) error {
	var (
		resp        = response.DefaultResponse{}
		ctx         = c.Request().Context()
		req         = request.UpdateDataUserRequest{}
		jwtUserData = entity.JwtUserData{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[UserHandler-1] UpdateDataUser: %s", "data token not found")
		resp.Message = "data token not found"
		resp.Data = nil
		return c.JSON(http.StatusNotFound, resp)
	}

	err := json.Unmarshal([]byte(user), &jwtUserData)
	if err != nil {
		log.Errorf("[UserHandler-2] UpdateDataUser: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	userID := jwtUserData.UserID

	if err = c.Bind(&req); err != nil {
		log.Errorf("[UserHandler-3] UpdateDataUser: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	if err = c.Validate(&req); err != nil {
		log.Errorf("[UserHandler-4] UpdateDataUser: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	reqEntity := entity.UserEntity{
		ID:      userID,
		Name:    req.Name,
		Email:   req.Email,
		Address: req.Address,
		Lat:     req.Lat,
		Lng:     req.Lng,
		Phone:   req.Phone,
		Photo:   req.Photo,
	}

	err = u.userService.UpdateDataUser(ctx, reqEntity)
	if err != nil {
		log.Errorf("[UserHandler-5] UpdateDataUser: %v", err)
		if err.Error() == "404" {
			resp.Message = "User not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp.Message = "Success"
	resp.Data = nil
	return c.JSON(http.StatusOK, resp)
}

// GetProfileUser implements UserHandlerInterface.
func (u *userHandler) GetProfileUser(c echo.Context) error {
	var (
		resp        = response.DefaultResponse{}
		respProfile = response.ProfileResponse{}
		ctx         = c.Request().Context()
		jwtUserData = entity.JwtUserData{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[UserHandler-1] GetProfileUser: %s", "data token not found")
		resp.Message = "data token not found"
		resp.Data = nil
		return c.JSON(http.StatusNotFound, resp)
	}

	err := json.Unmarshal([]byte(user), &jwtUserData)
	if err != nil {
		log.Errorf("[UserHandler-2] GetProfileUser: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	userID := jwtUserData.UserID

	dataUser, err := u.userService.GetProfileUser(ctx, userID)
	if err != nil {
		log.Errorf("[UserHandler-3] GetProfileUser: %v", err)
		if err.Error() == "404" {
			resp.Message = "user not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	respProfile.Address = dataUser.Address
	respProfile.Name = dataUser.Name
	respProfile.Email = dataUser.Email
	respProfile.ID = dataUser.ID
	respProfile.Lat = dataUser.Lat
	respProfile.Lng = dataUser.Lng
	respProfile.Phone = dataUser.Phone
	respProfile.Photo = dataUser.Photo
	respProfile.RoleName = dataUser.RoleName

	resp.Message = "success"
	resp.Data = respProfile

	return c.JSON(http.StatusOK, resp)
}

// UpdatePassword implements UserHandlerInterface.
func (u *userHandler) UpdatePassword(c echo.Context) error {
	var (
		resp = response.DefaultResponse{}
		req  = request.UpdatePasswordRequest{}
		ctx  = c.Request().Context()
	)

	tokenString := c.QueryParam("token")
	if tokenString == "" {
		log.Infof("[UserHandler-1] UpdatePassword: %s", "missing or invalid token")
		resp.Message = "missing or invalid token"
		resp.Data = nil
		return c.JSON(http.StatusUnauthorized, resp)
	}

	if err := c.Bind(&req); err != nil {
		log.Infof("[UserHandler-2] UpdatePassword: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	if err = c.Validate(req); err != nil {
		log.Errorf("[UserHandler-3] UpdatePassword: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	if req.NewPassword != req.ConfirmPassword {
		log.Infof("[UserHandler-4] UpdatePassword: %s", "new password and confirm password does not match")
		resp.Message = "new password and confirm password does not match"
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	reqEntity := entity.UserEntity{
		Password: req.NewPassword,
		Token:    tokenString,
	}

	err = u.userService.UpdatePassword(ctx, reqEntity)
	if err != nil {
		log.Errorf("[UserHandler-5] UpdatePassword: %v", err)
		if err.Error() == "404" {
			resp.Message = "User not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}

		if err.Error() == "401" {
			resp.Message = "Token expired or invalid"
			resp.Data = nil
			return c.JSON(http.StatusUnauthorized, resp)
		}
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp.Data = nil
	resp.Message = "Password updated successfully"

	return c.JSON(http.StatusOK, resp)
}

// VerifyToken implements UserHandlerInterface.
func (u *userHandler) VerifyAccount(c echo.Context) error {
	var (
		resp       = response.DefaultResponse{}
		respSignIn = response.SignInResponse{}
		ctx        = c.Request().Context()
	)

	tokenString := c.QueryParam("token")
	if tokenString == "" {
		log.Infof("[UserHandler-1] VerifyAccount: %s", "missing or invalid token")
		resp.Message = "missing or invalid token"
		resp.Data = nil
		return c.JSON(http.StatusUnauthorized, resp)
	}

	user, err := u.userService.VerifyToken(ctx, tokenString)
	if err != nil {
		log.Errorf("[UserHandler-2] VerifyAccount: %v", err)
		if err.Error() == "404" {
			resp.Message = "User not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}

		if err.Error() == "401" {
			resp.Message = "Token expired or invalid"
			resp.Data = nil
			return c.JSON(http.StatusUnauthorized, resp)
		}
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	respSignIn.ID = user.ID
	respSignIn.Name = user.Name
	respSignIn.Email = user.Email
	respSignIn.Role = user.RoleName
	respSignIn.Lat = user.Lat
	respSignIn.Lng = user.Lng
	respSignIn.Phone = user.Phone
	respSignIn.AccessToken = user.Token

	resp.Message = "Success"
	resp.Data = respSignIn

	return c.JSON(http.StatusOK, resp)

}

// ForgotPassword implements UserHandlerInterface.
func (u *userHandler) ForgotPassword(c echo.Context) error {
	var (
		req  = request.ForgotPasswordRequest{}
		resp = response.DefaultResponse{}
		ctx  = c.Request().Context()
	)

	if err = c.Bind(&req); err != nil {
		log.Errorf("[UserHandler-1] ForgotPassword: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	if err = c.Validate(req); err != nil {
		log.Errorf("[UserHandler-2] ForgotPassword: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	reqEntity := entity.UserEntity{
		Email: req.Email,
	}

	err = u.userService.ForgotPassword(ctx, reqEntity)
	if err != nil {
		log.Errorf("[UserHandler-3] ForgotPassword: %v", err)
		if err.Error() == "404" {
			resp.Message = "User not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp.Message = "Success"
	resp.Data = nil
	return c.JSON(http.StatusOK, resp)
}

// CreateUserAccount implements UserHandlerInterface.
func (u *userHandler) CreateUserAccount(c echo.Context) error {
	var (
		req  = request.SignUpRequest{}
		resp = response.DefaultResponse{}
		ctx  = c.Request().Context()
	)

	if err = c.Bind(&req); err != nil {
		return response.RespondWithError(c, http.StatusUnprocessableEntity, "[UserHandler-1] CreateUserAccount", err)
	}

	if err = c.Validate(req); err != nil {
		return response.RespondWithError(c, http.StatusUnprocessableEntity, "[UserHandler-2] CreateUserAccount", err)
	}

	if req.Password != req.PasswordConfirmation {
		err = errors.New("passwords do not match")
		return response.RespondWithError(c, http.StatusUnprocessableEntity, "[UserHandler-3] CreateUserAccount", err)
	}

	reqEntity := entity.UserEntity{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	err = u.userService.CreateUserAccount(ctx, reqEntity)
	if err != nil {
		return response.RespondWithError(c, http.StatusInternalServerError, "[UserHandler-4] CreateUserAccount", err)
	}

	resp.Message = "Success"
	return c.JSON(http.StatusCreated, resp)
}

var err error

// SignIn implements UserHandlerInterface.
func (u *userHandler) SignIn(c echo.Context) error {
	var (
		req        = request.SignInRequest{}
		resp       = response.DefaultResponse{}
		respSignIn = response.SignInResponse{}
		ctx        = c.Request().Context()
	)

	if err = c.Bind(&req); err != nil {
		log.Errorf("[UserHandler-1] SignIn: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	if err = c.Validate(req); err != nil {
		log.Errorf("[UserHandler-2] SignIn: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	reqEntity := entity.UserEntity{
		Email:    req.Email,
		Password: req.Password,
	}
	user, token, err := u.userService.SignIn(ctx, reqEntity)
	if err != nil {
		if err.Error() == "404" {
			log.Errorf("[UserHandler-3] SignIn: %s", "User Not Found")
			resp.Message = "User not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}
		log.Errorf("[UserHandler-4] SignIn: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	respSignIn.ID = user.ID
	respSignIn.Name = user.Name
	respSignIn.Email = user.Email
	respSignIn.Role = user.RoleName
	respSignIn.Lat = user.Lat
	respSignIn.Lng = user.Lng
	respSignIn.Phone = user.Phone
	respSignIn.AccessToken = token

	resp.Message = "Success"
	resp.Data = respSignIn

	return c.JSON(http.StatusOK, resp)
}

func NewUserHandler(e *echo.Echo, userService service.UserServiceInterface, cfg *config.Config, jwtService service.JwtServiceInterface) UserHandlerInterface {
	userHandler := &userHandler{userService: userService}

	e.Use(middleware.Recover())
	e.POST("/signin", userHandler.SignIn)
	e.POST("/signup", userHandler.CreateUserAccount)
	e.POST("/forgot-password", userHandler.ForgotPassword)
	e.GET("/verify-account", userHandler.VerifyAccount)
	e.PUT("/update-password", userHandler.UpdatePassword)

	mid := adapter.NewMiddlewareAdapter(cfg, jwtService)
	// Both Admin and Super Admin can read customer data. Mutations remain restricted below.
	adminGroup := e.Group("/admin", mid.CheckToken())
	adminGroup.GET("/customers", userHandler.GetCustomerAll)
	adminGroup.GET("/customers/:id", userHandler.GetCustomerByID)

	superAdminGroup := e.Group("/admin", mid.CheckToken(), mid.RequireRole("Super Admin"))
	superAdminGroup.POST("/customers", userHandler.CreateCustomer)
	superAdminGroup.PUT("/customers/:id", userHandler.UpdateCustomer)
	superAdminGroup.DELETE("/customers/:id", userHandler.DeleteCustomer)
	adminGroup.GET("/check", func(c echo.Context) error {
		return c.String(200, "OK")
	})

	authGroup := e.Group("/auth", mid.CheckToken())
	authGroup.PUT("/profile", userHandler.UpdateDataUser)
	authGroup.GET("/profile", userHandler.GetProfileUser)
	authGroup.GET("/customers/:id", userHandler.GetCustomerForAuth)

	return userHandler
}
