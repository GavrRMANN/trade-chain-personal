package httpapi

import (
	"net/http"

	"trade-chain/internal/domain"
	"trade-chain/internal/service"

	"github.com/go-chi/chi/v5"
)

type wishlistHandler struct{ s service.WishlistService }

type OptionRequest struct {
	CategoryID string `json:"category_id"`
}

func mountWishlistRoutes(r chi.Router, s service.WishlistService) {
	h := wishlistHandler{s}
	r.Route("/wishlists", func(r chi.Router) {
		r.Post("/", h.create)
		r.Get("/by-product/{productID}", h.byProduct)
		r.Get("/{id}", h.get)
		r.Delete("/{id}", h.delete)
		r.Get("/{id}/options", h.options)
		r.Post("/{id}/options", h.addOption)
		r.Delete("/{id}/options/{categoryID}", h.removeOption)
	})
}

// create godoc
// @Summary Create wishlist
// @Description Create a wishlist for a product
// @Tags wishlists
// @Accept json
// @Produce json
// @Param request body domain.Wishlist true "Wishlist data"
// @Success 201 {object} domain.Wishlist
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /wishlists [post]
func (h wishlistHandler) create(w http.ResponseWriter, r *http.Request) {
	var v domain.Wishlist
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
// @Summary Get wishlist by ID
// @Description Get wishlist details
// @Tags wishlists
// @Accept json
// @Produce json
// @Param id path string true "Wishlist ID"
// @Success 200 {object} domain.Wishlist
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /wishlists/{id} [get]
func (h wishlistHandler) get(w http.ResponseWriter, r *http.Request) {
	v, e := h.s.GetByID(r.Context(), chi.URLParam(r, "id"))
	if e != nil {
		writeError(w, e)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

// byProduct godoc
// @Summary Get wishlist by product
// @Description Get wishlist associated with a product
// @Tags wishlists
// @Accept json
// @Produce json
// @Param productID path string true "Product ID"
// @Success 200 {object} domain.Wishlist
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /wishlists/by-product/{productID} [get]
func (h wishlistHandler) byProduct(w http.ResponseWriter, r *http.Request) {
	v, e := h.s.GetByProductID(r.Context(), chi.URLParam(r, "productID"))
	if e != nil {
		writeError(w, e)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

// delete godoc
// @Summary Delete wishlist
// @Description Delete a wishlist
// @Tags wishlists
// @Accept json
// @Produce json
// @Param id path string true "Wishlist ID"
// @Success 204 "No content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /wishlists/{id} [delete]
func (h wishlistHandler) delete(w http.ResponseWriter, r *http.Request) {
	if e := h.s.Delete(r.Context(), chi.URLParam(r, "id")); e != nil {
		writeError(w, e)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// options godoc
// @Summary Get wishlist options (categories)
// @Description Get list of categories that the wishlist owner wants to receive
// @Tags wishlists
// @Accept json
// @Produce json
// @Param id path string true "Wishlist ID"
// @Success 200 {array} domain.Category
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /wishlists/{id}/options [get]
func (h wishlistHandler) options(w http.ResponseWriter, r *http.Request) {
	v, e := h.s.GetOptions(r.Context(), chi.URLParam(r, "id"))
	if e != nil {
		writeError(w, e)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

// addOption godoc
// @Summary Add category option to wishlist
// @Description Add a category that the wishlist owner wants to receive
// @Tags wishlists
// @Accept json
// @Produce json
// @Param id path string true "Wishlist ID"
// @Param request body OptionRequest true "Category ID"
// @Success 204 "No content"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /wishlists/{id}/options [post]
func (h wishlistHandler) addOption(w http.ResponseWriter, r *http.Request) {
	var v OptionRequest
	if decodeJSON(r, &v) != nil {
		writeError(w, service.ErrInvalidInput)
		return
	}
	if e := h.s.AddCategoryOption(r.Context(), chi.URLParam(r, "id"), v.CategoryID); e != nil {
		writeError(w, e)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// removeOption godoc
// @Summary Remove category option from wishlist
// @Description Remove a category from wishlist options
// @Tags wishlists
// @Accept json
// @Produce json
// @Param id path string true "Wishlist ID"
// @Param categoryID path string true "Category ID"
// @Success 204 "No content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /wishlists/{id}/options/{categoryID} [delete]
func (h wishlistHandler) removeOption(w http.ResponseWriter, r *http.Request) {
	if e := h.s.RemoveCategoryOption(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "categoryID")); e != nil {
		writeError(w, e)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
