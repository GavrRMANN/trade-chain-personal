package httpapi

import (
	"net/http"
	"os"
	"strings"

	"trade-chain/internal/auth"
	"trade-chain/internal/domain"
	"trade-chain/internal/service"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

type authHandler struct {
	customerService service.CustomerService
}

func NewAuthHandler(cs service.CustomerService) *authHandler {
	return &authHandler{customerService: cs}
}

func (h *authHandler) mountAuth(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", h.login)
		r.Post("/register", h.register)
		r.Post("/demo-login", h.demoLogin)

		r.Group(func(r chi.Router) {
			r.Use(auth.AuthMiddleware)
			r.Get("/me", h.me)
		})
	})
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	User  domain.Customer `json:"user"`
	Token string          `json:"token"`
}

// login godoc
// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/login [post]
func (h *authHandler) login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if decodeJSON(r, &req) != nil {
		writeError(w, service.ErrInvalidInput)
		return
	}
	customer, err := h.customerService.GetByEmail(r.Context(), req.Email)
	if err != nil {
		writeError(w, service.ErrInvalidInput)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(customer.PasswordHash), []byte(req.Password)); err != nil {
		writeError(w, service.ErrInvalidInput)
		return
	}
	token, err := auth.GenerateToken(customer.CustomerID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, AuthResponse{User: *customer, Token: token})
}

type DemoLoginRequest struct {
	CustomerID string `json:"customer_id"`
}

// demoLoginEnabled сообщает, поднят ли вход по выбору участника.
//
// Значение читается на каждом запросе, а не один раз при старте: маршрут
// объявлен всегда, и выключенный флаг должен давать понятную ошибку вместо
// 404 от роутера — иначе клиент не отличает «выключено» от «опечатка в пути».
func demoLoginEnabled() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("DEMO_LOGIN_ENABLED")), "true")
}

// demoLogin godoc
// @Summary Login as a demo customer
// @Description Issue a JWT for the given customer without a password. Available only when DEMO_LOGIN_ENABLED=true
// @Tags auth
// @Accept json
// @Produce json
// @Param request body DemoLoginRequest true "Customer to sign in as"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /auth/demo-login [post]
//
// Пароль здесь не спрашивается намеренно. Чтобы увидеть обмен, нужны две
// стороны с товарами, желаниями и историей: пустой свежезарегистрированный
// аккаунт не показывает ни каталога, ни цепочки. Это ход ради демонстрации,
// а не часть продуктовой авторизации, поэтому вход живёт за отдельным флагом
// и в окружении без DEMO_LOGIN_ENABLED отвечает 403.
func (h *authHandler) demoLogin(w http.ResponseWriter, r *http.Request) {
	if !demoLoginEnabled() {
		writeError(w, service.ErrForbidden)
		return
	}
	var req DemoLoginRequest
	if decodeJSON(r, &req) != nil {
		writeError(w, service.ErrInvalidInput)
		return
	}
	customer, err := h.customerService.GetByID(r.Context(), req.CustomerID)
	if err != nil {
		writeError(w, err)
		return
	}
	token, err := auth.GenerateToken(customer.CustomerID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, AuthResponse{User: *customer, Token: token})
}

// register godoc
// @Summary Register user
// @Description Register a new user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.CreateCustomerDTO true "Registration data"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/register [post]
func (h *authHandler) register(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateCustomerDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, service.ErrInvalidInput)
		return
	}
	customer, err := h.customerService.Create(r.Context(), &req)
	if err != nil {
		writeError(w, err)
		return
	}
	token, err := auth.GenerateToken(customer.CustomerID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, AuthResponse{User: *customer, Token: token})
}

// me godoc
// @Summary Get current user info
// @Description Get information about the authenticated user (validates token)
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} domain.Customer
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/me [get]
func (h *authHandler) me(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, service.ErrForbidden)
		return
	}
	customer, err := h.customerService.GetByID(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, customer)
}
