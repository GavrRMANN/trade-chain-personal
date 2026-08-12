package service

import (
	"context"

	"trade-chain/internal/domain"
	"trade-chain/internal/repository"
)

type notificationService struct {
	chains repository.ChainRepository
	reads  repository.NotificationRepository
}

func NewNotificationService(chains repository.ChainRepository, reads repository.NotificationRepository) NotificationService {
	return &notificationService{chains: chains, reads: reads}
}

func (s *notificationService) ListReads(ctx context.Context, customerID string) ([]domain.NotificationRead, error) {
	if blank(customerID) {
		return nil, ErrInvalidInput
	}
	reads, err := s.reads.ListReads(ctx, customerID)
	return reads, normalizeError(err)
}

func (s *notificationService) MarkRead(ctx context.Context, customerID, chainID string, kind domain.NotificationKind) error {
	if blank(customerID) || blank(chainID) || !isNotificationKind(kind) {
		return ErrInvalidInput
	}

	chain, err := s.chains.GetByID(ctx, chainID, customerID)
	if err != nil {
		return normalizeError(err)
	}
	if !isChainParticipant(chain, customerID) || notificationKindFor(chain, customerID) != kind {
		return ErrForbidden
	}
	return normalizeError(s.reads.MarkRead(ctx, customerID, chainID, kind))
}

func (s *notificationService) MarkAllRead(ctx context.Context, customerID string) error {
	if blank(customerID) {
		return ErrInvalidInput
	}

	chains, err := s.chains.GetByCustomerID(ctx, customerID)
	if err != nil {
		return normalizeError(err)
	}
	for _, chain := range chains {
		kind := notificationKindFor(&chain, customerID)
		if err := s.reads.MarkRead(ctx, customerID, chain.ChainID, kind); err != nil {
			return normalizeError(err)
		}
	}
	return nil
}

func isChainParticipant(chain *domain.Chain, customerID string) bool {
	return chain.InitiatorID == customerID || (chain.RecipientID != nil && *chain.RecipientID == customerID)
}

func notificationKindFor(chain *domain.Chain, customerID string) domain.NotificationKind {
	isIncoming := chain.InitiatorID != customerID
	status := domain.ChainStatus(chain.Status)
	if (status == domain.ChainPending || status == domain.ChainCountered) && isIncoming {
		return domain.NotificationIncomingOffer
	}
	if status == domain.ChainPending || status == domain.ChainCountered {
		return domain.NotificationOutgoingOffer
	}
	if status == domain.ChainActive {
		return domain.NotificationInProgress
	}
	return domain.NotificationFinished
}

func isNotificationKind(kind domain.NotificationKind) bool {
	switch kind {
	case domain.NotificationIncomingOffer, domain.NotificationOutgoingOffer, domain.NotificationInProgress, domain.NotificationFinished:
		return true
	default:
		return false
	}
}
