package sync

import (
	"context"
	"time"
	"vector/internal/models"
	"vector/internal/pkg/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func UpsertContracts(gdb *gorm.DB, contracts []models.Contract) error {
	if len(contracts) == 0 {
		return nil
	}

	return gdb.Table("core.contracts").
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "external_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"user_id", "comment", "created_at", "updated_at", "inner_code",
				"is_personal_invest_account", "is_personal_invest_account_new",
				"kind", "rialto_code", "signed_at", "closed_at", "status",
				"contract_owner_type", "contract_owner_id", "anketa", "owner_id",
				"calculated_profile_id", "depo_accounts_type", "strategy_id",
				"strategy_name", "tariff_id", "tariff_name", "user_login",
				"raw", "hash", "synced_at",
			}),
		}).
		Create(&contracts).Error
}

func GetCurrentContract(gdb *gorm.DB, contractID int) (*models.Contract, error) {
	var contract models.Contract
	err := gdb.Table("core.contracts").
		Where("external_id = ?", contractID).
		Take(&contract).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &contract, err
}

func ListContracts(gdb *gorm.DB, page, perPage int, userID *int, status *string) ([]models.Contract, int64, error) {
	var contracts []models.Contract
	var total int64

	query := gdb.Table("core.contracts")

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	err := query.
		Order("external_id DESC").
		Offset(offset).
		Limit(perPage).
		Find(&contracts).Error

	return contracts, total, err
}

func ApplyContractsBatch(gdb *gorm.DB, ctx context.Context, contractsData []ApplyContractData) (ApplyStats, error) {
	stats := ApplyStats{
		Created:   0,
		Updated:   0,
		Unchanged: 0,
	}

	if len(contractsData) == 0 {
		return stats, nil
	}

	now := time.Now().UTC()
	contracts := make([]models.Contract, 0, len(contractsData))

	for _, data := range contractsData {
		existing, err := GetCurrentContract(gdb, data.ContractID)
		if err != nil {
			continue
		}

		newContract := utils.ParseContract(data.RawData)
		newContract.ExternalID = data.ContractID
		newContract.Hash = data.Hash
		newContract.SyncedAt = now

		if existing == nil {
			contracts = append(contracts, newContract)
			stats.Created++
		} else if existing.Hash != data.Hash {
			newContract.ID = existing.ID
			contracts = append(contracts, newContract)
			stats.Updated++
		} else {

			stats.Unchanged++
		}
	}

	if len(contracts) > 0 {
		if err := UpsertContracts(gdb, contracts); err != nil {
			return stats, err
		}
	}

	return stats, nil
}

type ApplyContractData struct {
	ContractID int    `json:"contract_id"`
	RawData    []byte `json:"raw_data"`
	Hash       string `json:"hash"`
}

type ApplyStats struct {
	Created   int `json:"created"`
	Updated   int `json:"updated"`
	Unchanged int `json:"unchanged"`
}
