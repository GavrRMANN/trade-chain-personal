package httpapi

import (
	"net/http"
	"strconv"

	"trade-chain/internal/service"
)

func pagination(r *http.Request) (int, int, error) {
	offset, err := intQuery(r, "offset", 0)
	if err != nil {
		return 0, 0, service.ErrInvalidInput
	}
	limit, err := intQuery(r, "limit", 20)
	if err != nil {
		return 0, 0, service.ErrInvalidInput
	}
	return offset, limit, nil
}
func intQuery(r *http.Request, key string, def int) (int, error) {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def, nil
	}
	return strconv.Atoi(v)
}
