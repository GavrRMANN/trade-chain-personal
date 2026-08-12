package service

import (
	"errors"
	"fmt"

	"trade-chain/internal/domain"
)

// mapExchangeError переводит нарушение правил обмена в сервисную ошибку.
//
// Транспортный слой выбирает код ответа по сервисной ошибке, а текст отдаёт
// пользователю как есть, поэтому человеческая формулировка правила
// сохраняется: «это действие доступно другой стороне» объясняет отказ,
// а «operation forbidden» — нет.
//
// Через normalizeError результат пропускать нельзя: она сводит всё незнакомое
// к внутренней ошибке, и 403 превратился бы в 500.
func mapExchangeError(err error) error {
	switch {
	case err == nil:
		return nil

	case errors.Is(err, domain.ErrNotParticipant),
		errors.Is(err, domain.ErrWrongActor):
		return fmt.Errorf("%w: %s", ErrForbidden, err)

	case errors.Is(err, domain.ErrAlreadyConfirmed),
		errors.Is(err, domain.ErrOfferDuplicate),
		errors.Is(err, domain.ErrProductUnavailable):
		return fmt.Errorf("%w: %s", ErrConflict, err)

	case errors.Is(err, domain.ErrChainFinal),
		errors.Is(err, domain.ErrInvalidTransition),
		errors.Is(err, domain.ErrSelfExchange),
		errors.Is(err, domain.ErrInvalidSurcharge):
		return fmt.Errorf("%w: %s", ErrInvalidInput, err)

	default:
		return err
	}
}
