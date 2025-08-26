package repository

import (
	"time"
	"vector/internal/models"

	"gorm.io/datatypes"
)

type AppClientRepository interface {
	GetCurrent(clientID int) (models.ClientVersion, error)
	GetSecondPartCurrent(clientID int) (models.SecondPartVersion, error)
	ListSecondPartHistory(clientID int) ([]models.SecondPartVersion, error)
	CreateSecondPartDraft(clientID int, riskLevel *string, createdBy *int, dataOverride *datatypes.JSON) (models.SecondPartVersion, error)
	SubmitSecondPart(clientID int, userID *int) (models.SecondPartVersion, error)
	ApproveSecondPart(clientID int, approvedBy *int) (models.SecondPartVersion, error)
	RejectSecondPart(clientID int, userID *int, reason string) (models.SecondPartVersion, error)
	RequestDocsSecondPart(clientID int, userID *int, reason string) (models.SecondPartVersion, error)

	ListClientsWithSP(page, perPage int, needsSecondPart *bool, spStatus *string, dueBefore *time.Time) ([]models.ClientWithSP, int64, error)
}

type UserRepository interface {
	GetByToken(token string) (models.AppUser, error)
	Create(email, role, token string) (models.AppUser, error)
	List() ([]models.AppUser, error)
	UpdateRole(id uint, role string) (models.AppUser, error)
	RotateToken(id uint) (models.AppUser, error)
	Delete(id uint) error
	Seed() error
}

type CheckRepository interface {
	CreateSecondPartCheck(clientID, spVersion int, kind string, payload *datatypes.JSON, runBy *int) (models.SecondPartCheck, error)
	UpdateResult(checkID uint, status string, result *datatypes.JSON) (models.SecondPartCheck, error)
	ListByClient(clientID int, spVersion *int) ([]models.SecondPartCheck, error)
}

type RecalcRepository interface {
	RecalcNeedsSecondPart() (int64, error)
	RecalcPassportExpiry() (int64, error)
	RecalcAll() error
}
