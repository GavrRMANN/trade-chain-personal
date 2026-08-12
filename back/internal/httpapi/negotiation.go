package httpapi

import (
	"net/http"

	"trade-chain/internal/auth"
	"trade-chain/internal/domain"
	"trade-chain/internal/service"

	"github.com/go-chi/chi/v5"
)

// ConfirmRequest — решение стороны об итоге обмена.
type ConfirmRequest struct {
	Success bool `json:"success"`
}

// MessageRequest — реплика в переписке по сделке.
type MessageRequest struct {
	Body string `json:"body"`
}

func orientChainForCustomer(chain *domain.Chain, customerID string) {
	if chain.RecipientID == nil || *chain.RecipientID != customerID {
		return
	}

	if chain.ToProductID == nil {
		return
	}

	fromProductID := chain.FromProductID
	chain.FromProductID = *chain.ToProductID
	chain.ToProductID = &fromProductID
}

// mine godoc
// @Summary List my exchanges
// @Description Обмены пользователя: и предложенные им, и полученные
// @Tags chains
// @Produce json
// @Success 200 {array} domain.Chain
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /chains/my [get]
func (h chainHandler) mine(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, service.ErrForbidden)
		return
	}

	v, err := h.s.GetByCustomerID(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	for i := range v {
		orientChainForCustomer(&v[i], userID)
	}
	writeJSON(w, http.StatusOK, v)
}

// confirm godoc
// @Summary Confirm exchange outcome
// @Description Подтвердить итог обмена. Состоявшимся обмен считается только
// @Description после подтверждения обеими сторонами; несостоявшимся — по
// @Description заявлению одной.
// @Tags chains
// @Accept json
// @Produce json
// @Param id path string true "Chain ID"
// @Param request body ConfirmRequest true "Итог встречи"
// @Success 200 {object} domain.Chain
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /chains/{id}/confirm [post]
func (h chainHandler) confirm(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, service.ErrForbidden)
		return
	}

	var req ConfirmRequest
	if decodeJSON(r, &req) != nil {
		writeError(w, service.ErrInvalidInput)
		return
	}

	v, err := h.s.Confirm(r.Context(), chi.URLParam(r, "id"), userID, req.Success, "")
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

// messages godoc
// @Summary Read exchange chat
// @Description Переписка по сделке. Доступна только её участникам.
// @Tags chains
// @Produce json
// @Param id path string true "Chain ID"
// @Success 200 {array} domain.ChainMessage
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /chains/{id}/messages [get]
func (h chainHandler) messages(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, service.ErrForbidden)
		return
	}

	v, err := h.s.Messages(r.Context(), chi.URLParam(r, "id"), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

// sendMessage godoc
// @Summary Write to exchange chat
// @Description Написать второй стороне по этой сделке. По закрытой сделке
// @Description переписка недоступна.
// @Tags chains
// @Accept json
// @Produce json
// @Param id path string true "Chain ID"
// @Param request body MessageRequest true "Текст сообщения"
// @Success 201 {object} domain.ChainMessage
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /chains/{id}/messages [post]
func (h chainHandler) sendMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, service.ErrForbidden)
		return
	}

	var req MessageRequest
	if decodeJSON(r, &req) != nil {
		writeError(w, service.ErrInvalidInput)
		return
	}

	v, err := h.s.SendMessage(r.Context(), chi.URLParam(r, "id"), userID, req.Body)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, v)
}
