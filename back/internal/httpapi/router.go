package httpapi

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"trade-chain/internal/auth"
	"trade-chain/internal/events"
	"trade-chain/internal/search"
	"trade-chain/internal/service"

	_ "trade-chain/docs"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Dependencies struct {
	Customers     service.CustomerService
	Products      service.ProductService
	Chains        service.ChainService
	Offers        service.OfferService
	Reviews       service.ReviewService
	Categories    service.CategoryService
	Wishlists     service.WishlistService
	Notifications service.NotificationService
	Search        *search.SearchService
	Events        *events.Broker
	CronSecret    string
	// DB и DemoResetSecret нужны только служебному сбросу демо-стенда:
	// без них маршрут /demo/reset просто не поднимается.
	DB              *pgxpool.Pool
	DemoResetSecret string
}

func NewRouter(d Dependencies) http.Handler {
	r := chi.NewRouter()
	allowedOrigins := []string{"http://localhost:3000", "http://localhost:5173"}
	if configuredOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); configuredOrigins != "" {
		allowedOrigins = strings.Split(configuredOrigins, ",")
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
		},
	}))

	// Middleware. Объявляются до первого маршрута: chi не разрешает добавлять
	// их после и падает при старте, если раздача файлов встаёт выше.
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	fs := http.FileServer(http.Dir("./uploads"))
	r.With(middleware.Timeout(15*time.Second)).Handle("/uploads/*", http.StripPrefix("/uploads/", fs))
	// Health check
	r.With(middleware.Timeout(15*time.Second)).Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Swagger UI
	r.With(middleware.Timeout(15*time.Second)).Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// Создаём обработчик для аутентификации
	authHandler := NewAuthHandler(d.Customers)
	mountExpirationRoute(r, d.Chains, d.CronSecret)

	// Все маршруты. Верификацию делаю внутри маунтов
	r.Route("/api/v1", func(r chi.Router) {
		if d.Events != nil {
			r.With(auth.AuthMiddleware).Get("/events", eventsHandler{broker: d.Events}.stream)
		}

		r.Group(func(r chi.Router) {
			r.Use(middleware.Timeout(15 * time.Second))

			authHandler.mountAuth(r) // /auth/login, /auth/register
			mountDemoRoutes(r, d.DB, d.DemoResetSecret)

			if d.Customers != nil {
				mountCustomerRoutes(r, d.Customers)
			}
			if d.Products != nil {
				mountProductRoutes(r, d.Products, d.Wishlists, d.Search)
			}
			if d.Chains != nil {
				mountChainRoutes(r, d.Chains)
			}
			if d.Offers != nil {
				mountExchangeOfferRoutes(r, d.Offers)
			}
			if d.Reviews != nil {
				mountReviewRoutes(r, d.Reviews)
			}
			if d.Categories != nil {
				mountCategoryRoutes(r, d.Categories)
			}
			if d.Wishlists != nil {
				mountWishlistRoutes(r, d.Wishlists)
			}
			if d.Notifications != nil {
				mountNotificationRoutes(r, d.Notifications)
			}
			if d.Search != nil {
				mountSearchRoutes(r, d.Search)
			}
		})

	})

	chi.Walk(r, func(
		method string,
		route string,
		handler http.Handler,
		middlewares ...func(http.Handler) http.Handler,
	) error {
		log.Printf("ROUTE %s %s middleware=%d", method, route, len(middlewares))
		return nil
	})
	return r
}
