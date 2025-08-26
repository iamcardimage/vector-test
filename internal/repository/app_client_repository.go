package repository

import (
	"time"
	appdb "vector/internal/db/app"
	"vector/internal/models"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type appClientRepository struct {
	database *gorm.DB
}

func NewAppClientRepository(database *gorm.DB) AppClientRepository {
	return &appClientRepository{database: database}
}

func (r *appClientRepository) GetCurrent(clientID int) (models.ClientVersion, error) {
	return appdb.GetClientCurrent(r.database, clientID)
}

func (r *appClientRepository) GetSecondPartCurrent(clientID int) (models.SecondPartVersion, error) {
	return appdb.GetSecondPartCurrent(r.database, clientID)
}

func (r *appClientRepository) ListSecondPartHistory(clientID int) ([]models.SecondPartVersion, error) {
	return appdb.ListSecondPartHistory(r.database, clientID)
}

func (r *appClientRepository) CreateSecondPartDraft(clientID int, riskLevel *string, createdBy *int, dataOverride *datatypes.JSON) (models.SecondPartVersion, error) {
	return appdb.CreateSecondPartDraft(r.database, clientID, riskLevel, createdBy, dataOverride)
}

func (r *appClientRepository) SubmitSecondPart(clientID int, userID *int) (models.SecondPartVersion, error) {
	return appdb.SubmitSecondPart(r.database, clientID, userID)
}

func (r *appClientRepository) ApproveSecondPart(clientID int, approvedBy *int) (models.SecondPartVersion, error) {
	return appdb.ApproveSecondPart(r.database, clientID, approvedBy)
}

func (r *appClientRepository) RejectSecondPart(clientID int, userID *int, reason string) (models.SecondPartVersion, error) {
	return appdb.RejectSecondPart(r.database, clientID, userID, reason)
}

func (r *appClientRepository) RequestDocsSecondPart(clientID int, userID *int, reason string) (models.SecondPartVersion, error) {
	return appdb.RequestDocsSecondPart(r.database, clientID, userID, reason)
}

func (r *appClientRepository) ListClientsWithSP(page, perPage int, needsSecondPart *bool, spStatus *string, dueBefore *time.Time) ([]models.ClientWithSP, int64, error) {

	dbItems, total, err := appdb.ListClientsWithSP(r.database, page, perPage, needsSecondPart, spStatus, dueBefore)
	if err != nil {
		return nil, 0, err
	}

	// Конвертируем appdb.ClientWithSP в models.ClientWithSP
	items := make([]models.ClientWithSP, len(dbItems))
	for i, dbItem := range dbItems {
		items[i] = models.ClientWithSP{
			ClientID:          dbItem.ClientID,
			Surname:           dbItem.Surname,
			Name:              dbItem.Name,
			Patronymic:        dbItem.Patronymic,
			ExternalRiskLevel: dbItem.ExternalRiskLevel,
			NeedsSecondPart:   dbItem.NeedsSecondPart,
			SecondPartCreated: dbItem.SecondPartCreated,
			ClientVersion:     dbItem.ClientVersion,
			SpStatus:          dbItem.SpStatus,
			SpDueAt:           dbItem.SpDueAt,
			SpClientVersion:   dbItem.SpClientVersion,
		}
	}

	return items, total, nil
}
