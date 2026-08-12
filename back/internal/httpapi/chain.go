package httpapi

import (
	"errors"
	"net/http"

	"trade-chain/internal/auth"
	"trade-chain/internal/domain"
	"trade-chain/internal/service"

	"github.com/go-chi/chi/v5"
)

type chainHandler struct{ s service.ChainService }

type ChainStatusRequest struct {
	Status domain.ChainStatus `json:"status"`
}

func mountChainRoutes(r chi.Router, s service.ChainService) {
	h := chainHandler{s}
	r.Route("/chains", func(r chi.Router) {
		r.With(auth.AuthMiddleware).Post("/", h.create)
		// Статичный сегмент объявляется до шаблона, иначе chi разберёт
		// «my» как идентификатор звена.
		r.With(auth.AuthMiddleware).Get("/my", h.mine)

		// Временно сделал получение цепочек запросом, требующим авторизацию, чтобы внутри неё получать айдишник. Ну плюс странно, что это получается публичный запрос
		r.With(auth.AuthMiddleware).Get("/{id}", h.get)

		r.Get("/{id}/full", h.full)
		r.With(auth.AuthMiddleware).Patch("/{id}/status", h.status)
		r.With(auth.AuthMiddleware).Post("/{id}/confirm", h.confirm)
		r.With(auth.AuthMiddleware).Get("/{id}/messages", h.messages)
		r.With(auth.AuthMiddleware).Post("/{id}/messages", h.sendMessage)
		r.With(auth.AuthMiddleware).Delete("/{id}", h.delete)
		r.With(auth.AuthMiddleware).Get("/by-product/{productID}", h.byProduct)
	})
}

// create godoc
// @Summary Create a chain
// @Description Create a new exchange chain
// @Tags chains
// @Accept json
// @Produce json
// @Param request body domain.Chain true "Chain data (initiator_id will be set from token)"
// @Success 201 {object} domain.Chain
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /chains [post]
func (h chainHandler) create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, service.ErrForbidden)
		return
	}

	var v domain.Chain
	if decodeJSON(r, &v) != nil {
		writeError(w, service.ErrInvalidInput)
		return
	}
	v.InitiatorID = userID

	out, err := h.s.Create(r.Context(), &v)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

// get godoc
// @Summary Get chain by ID
// @Description Get chain details
// @Tags chains
// @Accept json
// @Produce json
// @Param id path string true "Chain ID"
// @Success 200 {object} domain.Chain
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /chains/{id} [get]
func (h chainHandler) get(w http.ResponseWriter, r *http.Request) {
	customerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, errors.New("Ошибка при получении ID пользователя"))
		return
	}

	v, err := h.s.GetByID(
		r.Context(),
		chi.URLParam(r, "id"),
		customerID,
	)
	if err != nil {
		writeError(w, err)
		return
	}

	orientChainForCustomer(v, customerID)

	writeJSON(w, http.StatusOK, v)
}

// full godoc
// @Summary Get full chain
// @Description Get full chain details (all linked chains)
// @Tags chains
// @Accept json
// @Produce json
// @Param id path string true "Chain ID"
// @Success 200 {array} domain.Chain
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /chains/{id}/full [get]
func (h chainHandler) full(w http.ResponseWriter, r *http.Request) {
	v, err := h.s.GetFullChain(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

// byProduct godoc
// @Summary Get chains by product
// @Description Get chains associated with a product (either from or to)
// @Tags chains
// @Accept json
// @Produce json
// @Param productID path string true "Product ID"
// @Success 200 {array} domain.Chain
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /chains/by-product/{productID} [get]
func (h chainHandler) byProduct(w http.ResponseWriter, r *http.Request) {
	customerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, errors.New("Ошибка при получении ID пользователя"))
		return
	}
	v, err := h.s.GetByProductID(r.Context(), chi.URLParam(r, "productID"), customerID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

// status godoc
// @Summary Update chain status
// @Description Update chain status (pending -> active -> completed, or cancelled/rejected)
// @Tags chains
// @Accept json
// @Produce json
// @Param id path string true "Chain ID"
// @Param request body ChainStatusRequest true "New status"
// @Success 204 "No content"
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /chains/{id}/status [patch]
func (h chainHandler) status(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, service.ErrForbidden)
		return
	}

	var req ChainStatusRequest
	if decodeJSON(r, &req) != nil {
		writeError(w, service.ErrInvalidInput)
		return
	}
	if err := h.s.UpdateStatus(r.Context(), chi.URLParam(r, "id"), req.Status, userID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// delete godoc
// @Summary Delete chain
// @Description Delete a chain (soft delete maybe, but hard delete in repo)
// @Tags chains
// @Accept json
// @Produce json
// @Param id path string true "Chain ID"
// @Success 204 "No content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /chains/{id} [delete]
func (h chainHandler) delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, service.ErrForbidden)
		return
	}
	if err := h.s.Delete(r.Context(), chi.URLParam(r, "id"), userID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
