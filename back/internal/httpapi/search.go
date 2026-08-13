package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"trade-chain/internal/auth"
	"trade-chain/internal/domain"
	"trade-chain/internal/search"
	"trade-chain/internal/service"

	"github.com/go-chi/chi/v5"
)

const defaultChainDepth = 10

type searchHandler struct{ s *search.SearchService }

// Поиск цепочки требует знать, чем владеет спрашивающий, поэтому маршрут
// закрыт аутентификацией.
func mountSearchRoutes(r chi.Router, s *search.SearchService) {
	h := searchHandler{s}

	r.Route("/search", func(r chi.Router) {
		r.Use(auth.AuthMiddleware)

		r.Get("/chain", h.chain)
		r.Get("/candidates", h.candidates)
	})
}

// CandidatesResponse — ответ подбора следующего шага обмена.
type CandidatesResponse struct {
	Products []domain.Product `json:"products"`
}

// ChainSearchResponse — ответ поиска цепочки.
type ChainSearchResponse struct {
	Chain  []domain.Product `json:"chain"`
	Length int              `json:"length"`
}

// chain godoc
// @Summary Find exchange chain
// @Description Строит путь от товаров текущего пользователя до целевого товара
// @Tags search
// @Produce json
// @Param source_product_id query string false "Source product ID"
// @Param target_product_id query string true "Target product ID"
// @Param max_depth query int false "Max depth, default 10"
// @Success 200 {object} ChainSearchResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /search/chain [get]
func (h searchHandler) chain(w http.ResponseWriter, r *http.Request) {
	target := strings.TrimSpace(r.URL.Query().Get("target_product_id"))
	source := strings.TrimSpace(r.URL.Query().Get("source_product_id"))
	if target == "" {
		writeError(w, service.ErrInvalidInput)
		return
	}

	depth := defaultChainDepth
	if raw := strings.TrimSpace(r.URL.Query().Get("max_depth")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			writeError(w, service.ErrInvalidInput)
			return
		}
		depth = parsed
	}

	customerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, service.ErrForbidden)
		return
	}

	var result *search.ProductSearchResult
	var err error
	if source == "" {
		result, err = h.s.FindChainToTarget(r.Context(), customerID, target, depth)
	} else {
		result, err = h.s.FindChain(r.Context(), customerID, source, target, depth)
	}
	if err != nil {
		writeError(w, err)
		return
	}

	// BFS без найденного пути отдаёт (nil, nil) — это не ошибка, а «цепочки
	// нет». Разыменовывать result здесь нельзя, иначе пустой ответ падает в
	// панику и 500. Отдаём документированный пустой результат.
	chain := []domain.Product{}
	length := 0
	if result != nil {
		if result.Products != nil {
			chain = result.Products
		}
		length = result.Length
	}

	writeJSON(w, http.StatusOK, ChainSearchResponse{
		Chain:  chain,
		Length: length,
	})
}

// candidates godoc
// @Summary Find exchange candidates
// @Description Подбирает следующий шаг обмена для товара: сперва совпадения по вишлисту, затем остальные активные товары
// @Tags search
// @Produce json
// @Param product_id query string true "Product ID"
// @Param limit query int false "Max candidates, default 8"
// @Param direct query bool false "Только вещи с прямым обменом, без добора каталогом"
// @Success 200 {object} CandidatesResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /search/candidates [get]
func (h searchHandler) candidates(w http.ResponseWriter, r *http.Request) {
	productID := strings.TrimSpace(r.URL.Query().Get("product_id"))
	if productID == "" {
		writeError(w, service.ErrInvalidInput)
		return
	}

	limit := 0
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			writeError(w, service.ErrInvalidInput)
			return
		}
		limit = parsed
	}

	directOnly := false
	if raw := strings.TrimSpace(r.URL.Query().Get("direct")); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			writeError(w, service.ErrInvalidInput)
			return
		}
		directOnly = parsed
	}

	if _, ok := auth.UserIDFromContext(r.Context()); !ok {
		writeError(w, service.ErrForbidden)
		return
	}

	products, err := h.s.FindCandidates(r.Context(), productID, limit, directOnly)
	if err != nil {
		writeError(w, err)
		return
	}
	if products == nil {
		products = []domain.Product{}
	}

	writeJSON(w, http.StatusOK, CandidatesResponse{Products: products})
}
