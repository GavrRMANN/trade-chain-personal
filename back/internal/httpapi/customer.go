package httpapi

import (
	"net/http"

	"trade-chain/internal/domain"
	"trade-chain/internal/service"

	"github.com/go-chi/chi/v5"
)

type customerHandler struct{ s service.CustomerService }

func mountCustomerRoutes(r chi.Router, s service.CustomerService) {
	h := customerHandler{s}
	r.Route("/customers", func(r chi.Router) {
		r.Get("/", h.list)
		// overview объявляется до /{id}: chi отдаёт статическому сегменту
		// приоритет над параметром, но порядок здесь ещё и читается как
		// «сначала списки, потом одна запись».
		r.Get("/overview", h.overview)
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
	v, e := h.s.GetByID(r.Context(), chi.URLParam(r, "id"))
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
	out, e := h.s.Update(r.Context(), chi.URLParam(r, "id"), &v)
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
	if e := h.s.Delete(r.Context(), chi.URLParam(r, "id")); e != nil {
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
