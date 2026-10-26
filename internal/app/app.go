package app

import (
	"context"
	"ecommerce/user-service/config"
	"ecommerce/user-service/internal/adapter/handler"
	"ecommerce/user-service/internal/adapter/message"
	"ecommerce/user-service/internal/adapter/repository"
	"ecommerce/user-service/internal/adapter/storage"
	"ecommerce/user-service/internal/core/service"
	"ecommerce/user-service/utils/validator"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10/translations/en"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	middlewareGateway "ecommerce/user-service/internal/middleware"
)

func RunServer() {
	cfg := config.NewConfig()
	consumerCtx, cancelConsumer := context.WithCancel(context.Background())
	defer cancelConsumer()

	stopConsumer, err := message.StartEmailVerificationConsumer(consumerCtx, cfg)
	if err != nil {
		log.Fatalf("[RunServer-1] start email consumer: %v", err)
	}
	defer func() {
		if err := stopConsumer(); err != nil {
			log.Printf("[RunServer] stop email consumer: %v", err)
		}
	}()

	db, err := cfg.ConnectPostgres()
	if err != nil {
		log.Fatalf("[RunServer-2] %v", err)
		return
	}

	userRepo := repository.NewUserRepository(db.DB)
	tokenRepo := repository.NewVerificationTokenRepository(db.DB)
	roleRepo := repository.NewRoleRepository(db.DB)

	jwtService := service.NewJwtService(cfg)
	userService := service.NewUserService(userRepo, cfg, jwtService, tokenRepo)
	roleService := service.NewRoleService(roleRepo)

	e := echo.New()
	e.Use(middleware.CORS())
	e.Use(middlewareGateway.GatewayValidationMiddleware())

	storageHandler := storage.NewSupabase(cfg)

	customerValidator := validator.NewValidator()
	en.RegisterDefaultTranslations(customerValidator.Validator, customerValidator.Translator)
	e.Validator = customerValidator

	handler.NewUserHandler(e, userService, cfg, jwtService)
	handler.NewUploadImage(e, cfg, storageHandler, jwtService)
	handler.NewRoleHandler(e, roleService, cfg, jwtService)

	go func() {
		if cfg.App.AppPort == "" {
			cfg.App.AppPort = os.Getenv("APP_PORT")
		}

		err = e.Start(":" + cfg.App.AppPort)
		if err != nil {
			log.Fatalf("[RunServer-3] %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	signal.Notify(quit, syscall.SIGTERM)

	<-quit

	log.Printf("[RunServer-4] Shutting Down server of 5 second...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	e.Shutdown(ctx)
}
