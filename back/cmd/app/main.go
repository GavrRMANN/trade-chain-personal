package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"trade-chain/internal/auth"
	"trade-chain/internal/events"
	"trade-chain/internal/httpapi"
	"trade-chain/internal/repository"
	"trade-chain/internal/search"
	"trade-chain/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	if err := auth.RequireSigningSecret(); err != nil {
		log.Fatal(err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatal("failed to ping database:", err)
	}

	// Репозитории
	customerRepo := repository.NewCustomerRepository(pool)
	productRepo := repository.NewProductRepository(pool)
	categoryRepo := repository.NewCategoryRepository(pool) // нужно создать, если нет
	wishlistRepo := repository.NewWishlistRepository(pool) // нужно создать
	chainRepo := repository.NewChainRepository(pool)
	negotiationRepo := repository.NewNegotiationRepository(pool)
	reviewRepo := repository.NewReviewRepository(pool)
	notificationRepo := repository.NewNotificationRepository(pool)

	// Сервисы
	customerService := service.NewCustomerService(customerRepo)
	productService := service.NewProductService(productRepo, customerRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	wishlistService := service.NewWishlistService(wishlistRepo, productRepo)
	eventBroker := events.NewBroker(pool)
	chainService := service.NewChainService(chainRepo, productRepo, negotiationRepo, eventBroker)
	offerService := service.NewOfferService(chainService, chainRepo, negotiationRepo)
	reviewService := service.NewReviewService(reviewRepo, customerRepo, productRepo, chainService)
	notificationService := service.NewNotificationService(chainRepo, notificationRepo)

	// Сервис поиска
	searchService := search.NewSearchService(productService, categoryService)

	// HTTP роутер
	deps := httpapi.Dependencies{
		Customers:     customerService,
		Products:      productService,
		Chains:        chainService,
		Offers:        offerService,
		Reviews:       reviewService,
		Categories:    categoryService,
		Wishlists:     wishlistService,
		Notifications: notificationService,
		Search:        searchService,
		Events:        eventBroker,
		CronSecret:    os.Getenv("CRON_SECRET"),

		DB:              pool,
		DemoResetSecret: os.Getenv("DEMO_RESET_SECRET"),
	}
	router := httpapi.NewRouter(deps)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("starting server on port %s", port)
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       65 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
