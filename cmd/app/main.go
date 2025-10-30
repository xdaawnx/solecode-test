package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"solecode/cmd/app/cli"
	_ "solecode/docs/swagger"
	"solecode/pkg/cache"
	soleCodeCache "solecode/pkg/cache"
	"solecode/pkg/config"
	"solecode/pkg/database"
	soleCodeHttp "solecode/src/delivery/http"
	repo "solecode/src/repository"
	uc "solecode/src/usecase"
)

// @title SoleCode User API
// @version 1.0
// @description A REST API for user with MySQL and Redis
// @termsOfService http://swagger.io/terms/

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

func main() {
	// If command line arguments are provided, execute CLI
	if len(os.Args) > 1 {
		cli.Execute()
		return
	}

	// Otherwise run as server
	runServer()
}

func runServer() {
	// Load configuration
	cfg, err := config.LoadConfig("conf/conf.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.NewMySQLDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize cache (Redis or NullCache as fallback)
	var cacheImpl soleCodeCache.CacheItf
	redisCache, err := cache.NewRedisCache(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis, using null cache: %v", err)
	} else {
		defer redisCache.Close()
		cacheImpl = redisCache
		log.Println("Redis cache connected successfully")
	}

	dom := repo.InitRepository(db)

	// Initialize all use cases
	uc := uc.InitUsecase(*dom, cacheImpl)
	// Initialize HTTP handler
	userHandler := soleCodeHttp.NewUserHandler(*uc)

	// Initialize router
	router := soleCodeHttp.NewRouter(userHandler)

	// Get the base handler
	handler := router.GetHandler()

	// Create server with timeouts
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      handler,
		ReadTimeout:  cfg.Server.Timeout,
		WriteTimeout: cfg.Server.Timeout,
		IdleTimeout:  60 * time.Second,
	}
	log.Printf("Server starting on port %s", cfg.Server.Port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}
