package service

import (
	"time"
	"vector/internal/models"
	"vector/internal/repository"

	"gorm.io/datatypes"
)

type AppService struct {
	clientRepo repository.AppClientRepository
	userRepo   repository.UserRepository
	checkRepo  repository.CheckRepository
	recalcRepo repository.RecalcRepository
}

func NewAppService(
	clientRepo repository.AppClientRepository,
	userRepo repository.UserRepository,
	checkRepo repository.CheckRepository,
	recalcRepo repository.RecalcRepository,
) *AppService {
	return &AppService{
		clientRepo: clientRepo,
		userRepo:   userRepo,
		checkRepo:  checkRepo,
		recalcRepo: recalcRepo,
	}
}

func (s *AppService) GetUserByToken(token string) (models.AppUser, error) {
	return s.userRepo.GetByToken(token)
}

func (s *AppService) CreateUser(email, role, token string) (models.AppUser, error) {
	return s.userRepo.Create(email, role, token)
}

func (s *AppService) ListUsers() ([]models.AppUser, error) {
	return s.userRepo.List()
}

func (s *AppService) UpdateUserRole(id uint, role string) (models.AppUser, error) {
	return s.userRepo.UpdateRole(id, role)
}

func (s *AppService) RotateUserToken(id uint) (models.AppUser, error) {
	return s.userRepo.RotateToken(id)
}

func (s *AppService) DeleteUser(id uint) error {
	return s.userRepo.Delete(id)
}

func (s *AppService) GetClientCurrent(clientID int) (models.ClientVersion, error) {
	return s.clientRepo.GetCurrent(clientID)
}

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

func (s *AppService) ListClientsWithSP(page, perPage int, needsSecondPart *bool, spStatus *string, dueBefore *time.Time) ([]models.ClientWithSP, int64, error) {
	return s.clientRepo.ListClientsWithSP(page, perPage, needsSecondPart, spStatus, dueBefore)
}

func (s *AppService) CreateSecondPartCheck(clientID, spVersion int, kind string, payload *datatypes.JSON, runBy *int) (models.SecondPartCheck, error) {
	return s.checkRepo.CreateSecondPartCheck(clientID, spVersion, kind, payload, runBy)
}

func (s *AppService) UpdateCheckResult(checkID uint, status string, result *datatypes.JSON) (models.SecondPartCheck, error) {
	return s.checkRepo.UpdateResult(checkID, status, result)
}

func (s *AppService) ListChecksByClient(clientID int, spVersion *int) ([]models.SecondPartCheck, error) {
	return s.checkRepo.ListByClient(clientID, spVersion)
}

func (s *AppService) RecalcNeedsSecondPart() (int64, error) {
	return s.recalcRepo.RecalcNeedsSecondPart()
}

func (s *AppService) RecalcPassportExpiry() (int64, error) {
	return s.recalcRepo.RecalcPassportExpiry()
}

func (s *AppService) RecalcAll() error {
	return s.recalcRepo.RecalcAll()
}
