package service

import (
	"context"
	"errors"
	"time"
	"vector/internal/models"
	"vector/internal/repository"

	"gorm.io/datatypes"
)

type AppService struct {
	clientRepo       repository.AppClientRepository
	checkRepo        repository.CheckRepository
	recalcRepo       repository.RecalcRepository
	syncContractRepo repository.SyncContractRepository
}

func NewAppService(
	clientRepo repository.AppClientRepository,
	userRepo repository.UserRepository,
	checkRepo repository.CheckRepository,
	recalcRepo repository.RecalcRepository,
	syncContractRepo repository.SyncContractRepository,
) *AppService {
	return &AppService{
		clientRepo:       clientRepo,
		checkRepo:        checkRepo,
		recalcRepo:       recalcRepo,
		syncContractRepo: syncContractRepo,
	}
}

// ========== МЕТОДЫ ДЛЯ КЛИЕНТОВ ==========

func (s *AppService) GetClientCurrent(clientID int) (models.ClientVersion, error) {
	return s.clientRepo.GetCurrent(clientID)
}

func (s *AppService) GetClientHistory(clientID int) ([]models.ClientVersion, error) {
	return s.clientRepo.GetClientHistory(clientID)
}

func (s *AppService) GetClientVersion(clientID int, version int) (models.ClientVersion, error) {
	return s.clientRepo.GetClientVersion(clientID, version)
}

func (s *AppService) ListClientsWithSP(page, perPage int, needsSecondPart *bool, spStatus *string, dueBefore *time.Time) ([]models.ClientWithSP, int64, error) {
	return s.clientRepo.ListClientsWithSP(page, perPage, needsSecondPart, spStatus, dueBefore)
}

// ========== МЕТОДЫ ДЛЯ ВТОРОЙ ЧАСТИ ==========

func (s *AppService) GetSecondPartCurrent(clientID int) (models.SecondPartVersion, error) {
	return s.clientRepo.GetSecondPartCurrent(clientID)
}

func (s *AppService) ListSecondPartHistory(clientID int) ([]models.SecondPartVersion, error) {
	return s.clientRepo.ListSecondPartHistory(clientID)
}

func (s *AppService) CreateSecondPartDraft(clientID int, riskLevel *string, createdBy *int, dataOverride *datatypes.JSON) (models.SecondPartVersion, error) {
	return s.clientRepo.CreateSecondPartDraft(clientID, riskLevel, createdBy, dataOverride)
}

func (s *AppService) SubmitSecondPart(clientID int, userID *int) (models.SecondPartVersion, error) {
	return s.clientRepo.SubmitSecondPart(clientID, userID)
}

func (s *AppService) ApproveSecondPart(clientID int, approvedBy *int) (models.SecondPartVersion, error) {
	return s.clientRepo.ApproveSecondPart(clientID, approvedBy)
}

func (s *AppService) RejectSecondPart(clientID int, userID *int, reason string) (models.SecondPartVersion, error) {
	return s.clientRepo.RejectSecondPart(clientID, userID, reason)
}

func (s *AppService) RequestDocsSecondPart(clientID int, userID *int, reason string) (models.SecondPartVersion, error) {
	return s.clientRepo.RequestDocsSecondPart(clientID, userID, reason)
}

// ========== МЕТОДЫ ДЛЯ КОНТРАКТОВ ==========

func (s *AppService) GetContract(contractID int) (models.Contract, error) {
	ctx := context.Background()
	contract, err := s.syncContractRepo.GetCurrentContract(ctx, contractID)
	if err != nil {
		return models.Contract{}, err
	}
	if contract == nil {
		return models.Contract{}, errors.New("contract not found")
	}
	return *contract, nil
}

func (s *AppService) ListContracts(page, perPage int, userID *int, status *string) ([]models.Contract, int64, error) {
	return s.syncContractRepo.ListContracts(page, perPage, userID, status)
}

// ========== МЕТОДЫ ДЛЯ ПРОВЕРОК ==========

func (s *AppService) CreateSecondPartCheck(clientID, spVersion int, kind string, payload *datatypes.JSON, runBy *int) (models.SecondPartCheck, error) {
	return s.checkRepo.CreateSecondPartCheck(clientID, spVersion, kind, payload, runBy)
}

func (s *AppService) UpdateCheckResult(checkID uint, status string, result *datatypes.JSON) (models.SecondPartCheck, error) {
	return s.checkRepo.UpdateResult(checkID, status, result)
}

func (s *AppService) ListChecksByClient(clientID int, spVersion *int) ([]models.SecondPartCheck, error) {
	return s.checkRepo.ListByClient(clientID, spVersion)
}

// ========== МЕТОДЫ ДЛЯ ПЕРЕСЧЕТОВ ==========

func (s *AppService) RecalcNeedsSecondPart() (int64, error) {
	return s.recalcRepo.RecalcNeedsSecondPart()
}

func (s *AppService) RecalcPassportExpiry() (int64, error) {
	return s.recalcRepo.RecalcPassportExpiry()
}

func (s *AppService) RecalcAll() error {
	return s.recalcRepo.RecalcAll()
}
