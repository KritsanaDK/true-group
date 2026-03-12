package main

import (
	"context"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"tdg/internal/cache"
	"tdg/internal/handler"
	"tdg/internal/infra/database"
	"tdg/internal/infra/logger"
	"tdg/internal/repository"
	"tdg/internal/service"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
)

func init() {
	// Force IPv4 to prevent "socket not connected" errors
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 30 * time.Second,
			}
			// Force TCP4 instead of TCP6
			return d.DialContext(ctx, "tcp4", address)
		},
	}
}

// loadEnv loads and logs environment variables
func loadEnv() (env, version, appName, dbUser, dbPassword, dbHost, dbPort, dbName, redisAddr, redisPassword string, debug bool) {
	env = os.Getenv("ENV")
	version = os.Getenv("VERSION")
	appName = os.Getenv("NAME")

	dbUser = os.Getenv("DB_USER")
	dbPassword = os.Getenv("DB_PASSWORD")
	dbHost = os.Getenv("DB_HOST")
	dbPort = os.Getenv("DB_PORT")
	dbName = os.Getenv("DB_NAME")
	redisAddr = os.Getenv("REDIS_ADDR")
	redisPassword = os.Getenv("REDIS_PASSWORD")

	logger.Debug("Environment Variables",
		"ENV", env,
		"VERSION", version,
		"NAME", appName,
		"DB_USER", dbUser,
		"DB_HOST", dbHost,
		"DB_PORT", dbPort,
		"DB_NAME", dbName,
		"REDIS_ADDR", redisAddr,
		"REDIS_PASSWORD", redisPassword,
	)

	var err error
	debug, err = strconv.ParseBool(os.Getenv("DEBUG"))
	if err != nil {
		debug = false
	}
	return
}

func main() {
	env, version, appName, dbUser, dbPassword, dbHost, dbPort, dbName, redisAddr, redisPassword, debug := loadEnv()

	// Initialize logger
	if err := logger.InitialzeLoggerSystem(); err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	logger.Debug(env, version, appName, debug)

	ctx := context.Background()

	// ✅ Database - Connect in goroutine
	var db *pgxpool.Pool
	var redisDb *redis.Client
	var wg sync.WaitGroup
	errChan := make(chan error, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("starting database connection...")

		var err error
		db, err = database.ConnectPostgres(ctx, database.ConfigPostgres{
			User:     dbUser,
			Password: dbPassword,
			Host:     dbHost,
			Port:     dbPort,
			Name:     dbName,
		})
		if err != nil {
			errChan <- err
			return
		}
		logger.Info("database connection established")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("starting Redis connection...")

		var err error
		redisDb, err = database.ConnectRedis(ctx, database.ConfigRedis{
			RedisEndPoint: redisAddr,
			Password:      redisPassword,
		})

		if err != nil {
			errChan <- err
			return
		}

		logger.Info("Redis connection established")
	}()

	// Wait for database connection to complete
	wg.Wait()
	close(errChan)

	if err := <-errChan; err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	defer func() {
		logger.Info("closing database connection")
		db.Close()
	}()

	trueRepository, err := repository.NewTrueRepository(ctx, db)
	if err != nil {
		log.Fatal(err)
		return
	}

	redisClient, err := cache.NewRedisClient(ctx, redisDb)
	if err != nil {
		log.Fatal(err)
		return
	}

	// --- DI ---
	repo := service.AllRepository{
		ITrueRepository: trueRepository,
		ICacheClient:    redisClient,
	}

	trueService := service.NewTrueService(ctx, debug, repo)

	svc := service.AllService{
		ITrueService: trueService,
	}

	trueHandler := handler.NewTrueHandler(debug, svc)

	// --- Echo ---
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			echo.GET, echo.POST, echo.PUT, echo.PATCH, echo.DELETE,
		},
	}))

	e.GET("/health", trueHandler.Health)

	e.GET("/users/:user_id/recommendations", trueHandler.GetUserRecommendations)
	e.GET("/recommendations/batch", trueHandler.GetBatchRecommendations)
	// --- Start ---
	e.Logger.Fatal(e.Start(":8080"))
}
