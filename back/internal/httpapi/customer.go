package httpapi

import (
	"net/http"

	"trade-chain/internal/auth"
	"trade-chain/internal/domain"
	"trade-chain/internal/service"

	"github.com/go-chi/chi/v5"
)

type customerHandler struct {
	s service.CustomerService
}

func mountCustomerRoutes(r chi.Router, s service.CustomerService) {
	h := customerHandler{s}

	r.Route("/customers", func(r chi.Router) {
		r.Get("/", h.list)
		r.Get("/overview", h.overview)

		// Рекомендации текущего пользователя. Кто этот пользователь, знает
		// только аутентификация: без неё в контексте нет идентификатора, и
		// ручки отвечали бы ошибкой запроса даже на верный токен.
		r.Group(func(r chi.Router) {
			r.Use(auth.AuthMiddleware)

			r.Post("/me/recommendations", h.createRecommendation)
			r.Get("/me/recommendations", h.getMyRecommendations)
			r.Patch("/me/recommendations", h.updateRecommendations)
			r.Delete("/me/recommendations/{categoryID}", h.deleteRecommendation)
		})

		// Рекомендации конкретного пользователя.
		r.Get("/{customerID}/recommendations", h.getRecommendations)

		r.Get("/{id}", h.get)
		r.Patch("/{id}", h.update)
		r.Delete("/{id}", h.delete)
	})
}

// get godoc
// @Summary Get customer by ID
// @Description Get customer details
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Success 200 {object} domain.Customer
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers/{id} [get]
func (h customerHandler) get(w http.ResponseWriter, r *http.Request) {
	v, e := h.s.GetByID(
		r.Context(),
		chi.URLParam(r, "id"),
	)
	if e != nil {
		writeError(w, e)
		return
	}

	writeJSON(w, http.StatusOK, v)
}

// update godoc
// @Summary Update customer
// @Description Update customer information
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Param request body domain.UpdateCustomerDTO true "Updated customer data"
// @Success 200 {object} domain.Customer
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers/{id} [patch]
func (h customerHandler) update(w http.ResponseWriter, r *http.Request) {
	var v domain.UpdateCustomerDTO

	if decodeJSON(r, &v) != nil {
		writeError(w, service.ErrInvalidInput)
		return
	}

	out, e := h.s.Update(
		r.Context(),
		chi.URLParam(r, "id"),
		&v,
	)
	if e != nil {
		writeError(w, e)
		return
	}

	writeJSON(w, http.StatusOK, out)
}

// delete godoc
// @Summary Delete customer
// @Description Soft delete customer (set is_active=false)
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Success 204 "No content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers/{id} [delete]
func (h customerHandler) delete(w http.ResponseWriter, r *http.Request) {
	if e := h.s.Delete(
		r.Context(),
		chi.URLParam(r, "id"),
	); e != nil {
		writeError(w, e)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// list godoc
// @Summary List customers
// @Description List customers with pagination
// @Tags customers
// @Accept json
// @Produce json
// @Param offset query int false "Offset" default(0)
// @Param limit query int false "Limit" default(20) maximum(100)
// @Success 200 {array} domain.Customer
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers [get]
func (h customerHandler) list(w http.ResponseWriter, r *http.Request) {
	o, l, e := pagination(r)
	if e != nil {
		writeError(w, e)
		return
	}

	v, e := h.s.List(r.Context(), o, l)
	if e != nil {
		writeError(w, e)
		return
	}

	writeJSON(w, http.StatusOK, v)
}

// overview godoc
// @Summary List customers with activity stats
// @Description List customers together with rating, review count, total and active products, and exchange chains
// @Tags customers
// @Accept json
// @Produce json
// @Param offset query int false "Offset" default(0)
// @Param limit query int false "Limit" default(20) maximum(100)
// @Success 200 {array} domain.CustomerOverview
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers/overview [get]
func (h customerHandler) overview(w http.ResponseWriter, r *http.Request) {
	o, l, e := pagination(r)
	if e != nil {
		writeError(w, e)
		return
	}

	v, e := h.s.ListOverview(r.Context(), o, l)
	if e != nil {
		writeError(w, e)
		return
	}

	writeJSON(w, http.StatusOK, v)
}

// getRecommendations godoc
// @Summary Get customer recommendations
// @Description Get category preferences of a customer
// @Tags customers
// @Accept json
// @Produce json
// @Param customerID path string true "Customer ID"
// @Success 200 {array} domain.CustomerWishlistOption
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers/{customerID}/recommendations [get]
func (h customerHandler) getRecommendations(w http.ResponseWriter, r *http.Request) {
	v, e := h.s.GetCustomerWishlistOptions(
		r.Context(),
		chi.URLParam(r, "customerID"),
	)
	if e != nil {
		writeError(w, e)
		return
	}

	writeJSON(w, http.StatusOK, v)
}

// getMyRecommendations godoc
// @Summary Get current customer recommendations
// @Description Get category preferences of the current customer
// @Tags customers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} domain.CustomerWishlistOption
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers/me/recommendations [get]
func (h customerHandler) getMyRecommendations(w http.ResponseWriter, r *http.Request) {
	customerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, service.ErrForbidden)
		return
	}

	v, e := h.s.GetCustomerWishlistOptions(
		r.Context(),
		customerID,
	)
	if e != nil {
		writeError(w, e)
		return
	}

	writeJSON(w, http.StatusOK, v)
}

// createRecommendation godoc
// @Summary Add customer recommendations
// @Description Add category preferences to the current customer
// @Tags customers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.UpdateCustomerWishlistDTO true "Customer wishlist options"
// @Success 201 {array} domain.CustomerWishlistOption
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers/me/recommendations [post]
func (h customerHandler) createRecommendation(w http.ResponseWriter, r *http.Request) {
	customerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, service.ErrForbidden)
		return
	}

	var dto domain.UpdateCustomerWishlistDTO

	if decodeJSON(r, &dto) != nil {
		writeError(w, service.ErrInvalidInput)
		return
	}

	for _, categoryID := range dto.CategoryIDs {
		if err := h.s.AddCustomerWishlistOption(
			r.Context(),
			customerID,
			categoryID,
		); err != nil {
			writeError(w, err)
			return
		}
	}

	v, err := h.s.GetCustomerWishlistOptions(
		r.Context(),
		customerID,
	)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, v)
}

// updateRecommendations godoc
// @Summary Update customer recommendations
// @Description Replace category preferences of the current customer
// @Tags customers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.UpdateCustomerWishlistDTO true "Customer wishlist options"
// @Success 200 {array} domain.CustomerWishlistOption
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers/me/recommendations [patch]
func (h customerHandler) updateRecommendations(w http.ResponseWriter, r *http.Request) {
	customerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, service.ErrForbidden)
		return
	}

	var dto domain.UpdateCustomerWishlistDTO

	if decodeJSON(r, &dto) != nil {
		writeError(w, service.ErrInvalidInput)
		return
	}

	if err := h.s.ReplaceCustomerWishlistOptions(
		r.Context(),
		customerID,
		&dto,
	); err != nil {
		writeError(w, err)
		return
	}

	v, err := h.s.GetCustomerWishlistOptions(
		r.Context(),
		customerID,
	)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, v)
}

// deleteRecommendation godoc
// @Summary Delete customer recommendation
// @Description Delete a category from the current customer's preferences
// @Tags customers
// @Produce json
// @Security BearerAuth
// @Param categoryID path string true "Category ID"
// @Success 204 "No content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers/me/recommendations/{categoryID} [delete]
func (h customerHandler) deleteRecommendation(w http.ResponseWriter, r *http.Request) {
	customerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, service.ErrForbidden)
		return
	}

	if e := h.s.DeleteCustomerWishlistOption(
		r.Context(),
		customerID,
		chi.URLParam(r, "categoryID"),
	); e != nil {
		writeError(w, e)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
