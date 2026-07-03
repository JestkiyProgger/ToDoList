package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	core_config "github.com/JestkiyProgger/ToDoList/internal/core/config"
	core_logger "github.com/JestkiyProgger/ToDoList/internal/core/logger"
	"github.com/JestkiyProgger/ToDoList/internal/core/repository/postgres/pool/pgx"
	core_http_middleware "github.com/JestkiyProgger/ToDoList/internal/core/transport/http/middleware"
	core_http_server "github.com/JestkiyProgger/ToDoList/internal/core/transport/http/server"
	tasks_postgres_repository "github.com/JestkiyProgger/ToDoList/internal/features/tasks/repository/postgres"
	tasks_service "github.com/JestkiyProgger/ToDoList/internal/features/tasks/service"
	tasks_transport_http "github.com/JestkiyProgger/ToDoList/internal/features/tasks/transport/http"
	users_postgres_repository "github.com/JestkiyProgger/ToDoList/internal/features/users/repository/postgres"
	users_service "github.com/JestkiyProgger/ToDoList/internal/features/users/service"
	users_transport_http "github.com/JestkiyProgger/ToDoList/internal/features/users/transport/http"
	"go.uber.org/zap"
)

func main() {
	cfg := core_config.NewConfigMust()
	time.Local = cfg.TimeZone

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger, err := core_logger.NewLogger(core_logger.NewConfigMust())
	if err != nil {
		fmt.Println("Failed to init logger:", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Debug("application time zone", zap.Any("zone", time.Local))

	logger.Debug("initializing postgres pool")
	pool, err := core_pgx_pool.NewPool(ctx, core_pgx_pool.NewConfigMust())
	if err != nil {
		logger.Fatal("failed to init postgres connection pool", zap.Error(err))
	}
	defer pool.Close()

	logger.Debug("initializing feature", zap.String("feature", "users"))
	usersRepository := users_postgres_repository.NewUsersRepository(pool)
	usersService := users_service.NewUsersService(usersRepository)
	usersTransportHttp := users_transport_http.NewUsersHTTPHandler(usersService)

	logger.Debug("initializing feature", zap.String("feature", "tasks"))
	tasksRepository := tasks_postgres_repository.NewTasksRepository(pool)
	tasksService := tasks_service.NewTasksService(tasksRepository)
	tasksTransportHttp := tasks_transport_http.NewTaskHTTPHandler(tasksService)

	logger.Debug("initializing HTTP server")
	httpServer := core_http_server.NewHTTPServer(
		core_http_server.NewConfigMust(),
		logger,
		core_http_middleware.RequestID(),
		core_http_middleware.Logger(logger),
		core_http_middleware.Trace(),
		core_http_middleware.Panic(),
	)
	apiVersionRouter := core_http_server.NewAPIVersionRouter(core_http_server.ApiVersion1)
	apiVersionRouter.RegisterRoutes(usersTransportHttp.Routes()...)
	apiVersionRouter.RegisterRoutes(tasksTransportHttp.Routes()...)
	httpServer.RegisterAPIRouters(apiVersionRouter)

	if err := httpServer.Run(ctx); err != nil {
		logger.Error("HTTP server run error", zap.Error(err))
	}
}
