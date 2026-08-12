package httpapi

import (
	"net/http"

	"trade-chain/internal/domain"
	"trade-chain/internal/service"

	"github.com/go-chi/chi/v5"
)

type categoryHandler struct{ s service.CategoryService }

func mountCategoryRoutes(r chi.Router, s service.CategoryService) {
	h := categoryHandler{s}
	r.Route("/categories", func(r chi.Router) {
		r.Post("/", h.create)
		r.Get("/", h.list)
		r.Get("/{id}", h.get)
		r.Get("/{id}/subcategories", h.subcategories)
		r.Put("/{id}", h.update)
		r.Delete("/{id}", h.delete)
	})
}

// create godoc
// @Summary Create category
// @Description Create a new category
// @Tags categories
// @Accept json
// @Produce json
// @Param request body domain.Category true "Category data"
// @Success 201 {object} domain.Category
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /categories [post]
func (h categoryHandler) create(w http.ResponseWriter, r *http.Request) {
	var v domain.Category
	if decodeJSON(r, &v) != nil {
		writeError(w, service.ErrInvalidInput)
		return
	}
	out, e := h.s.Create(r.Context(), &v)
	if e != nil {
		writeError(w, e)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

// get godoc
// @Summary Get category by ID
// @Description Get category details
// @Tags categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Success 200 {object} domain.Category
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /categories/{id} [get]
func (h categoryHandler) get(w http.ResponseWriter, r *http.Request) {
	v, e := h.s.GetByID(r.Context(), chi.URLParam(r, "id"))
	if e != nil {
		writeError(w, e)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

// subcategories godoc
// @Summary Get subcategories
// @Description Get subcategories of a category
// @Tags categories
// @Accept json
// @Produce json
// @Param id path string true "Parent Category ID"
// @Success 200 {array} domain.Category
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /categories/{id}/subcategories [get]
func (h categoryHandler) subcategories(w http.ResponseWriter, r *http.Request) {
	v, e := h.s.GetSubcategories(r.Context(), chi.URLParam(r, "id"))
	if e != nil {
		writeError(w, e)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

// update godoc
// @Summary Update category
// @Description Update category information
// @Tags categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Param request body domain.Category true "Updated category data"
// @Success 200 {object} domain.Category
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /categories/{id} [put]
func (h categoryHandler) update(w http.ResponseWriter, r *http.Request) {
	var v domain.Category
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
// @Summary Delete category
// @Description Delete a category
// @Tags categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Success 204 "No content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /categories/{id} [delete]
func (h categoryHandler) delete(w http.ResponseWriter, r *http.Request) {
	if e := h.s.Delete(r.Context(), chi.URLParam(r, "id")); e != nil {
		writeError(w, e)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// list godoc
// @Summary List categories
// @Description List all categories
// @Tags categories
// @Accept json
// @Produce json
// @Success 200 {array} domain.Category
// @Failure 500 {object} ErrorResponse
// @Router /categories [get]
func (h categoryHandler) list(w http.ResponseWriter, r *http.Request) {
	v, e := h.s.List(r.Context())
	if e != nil {
		writeError(w, e)
		return
	}
	writeJSON(w, http.StatusOK, v)
}
