package sync

import (
	"context"
	"time"
	"vector/internal/models"
	"vector/internal/pkg/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UpsertContracts добавляет или обновляет договоры в core.contracts
func UpsertContracts(gdb *gorm.DB, contracts []models.Contract) error {
	if len(contracts) == 0 {
		return nil
	}

	// ИСПРАВИЛИ: используем clause.OnConflict
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

// GetCurrentContract возвращает текущий договор по ID
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

// ListContracts возвращает список договоров с пагинацией
func ListContracts(gdb *gorm.DB, page, perPage int, userID *int, status *string) ([]models.Contract, int64, error) {
	var contracts []models.Contract
	var total int64

	query := gdb.Table("core.contracts")

	// Фильтры
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	// Считаем общее количество
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Получаем данные с пагинацией
	offset := (page - 1) * perPage
	err := query.
		Order("external_id DESC").
		Offset(offset).
		Limit(perPage).
		Find(&contracts).Error

	return contracts, total, err
}

// ApplyContractsBatch применяет изменения договоров
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
		// Проверяем существующий договор
		existing, err := GetCurrentContract(gdb, data.ContractID)
		if err != nil {
			continue
		}

		// Парсим новые данные
		newContract := utils.ParseContract(data.RawData)
		newContract.ExternalID = data.ContractID
		newContract.Hash = data.Hash
		newContract.SyncedAt = now

		if existing == nil {
			// Новый договор
			contracts = append(contracts, newContract)
			stats.Created++
		} else if existing.Hash != data.Hash {
			// Изменения есть
			newContract.ID = existing.ID // Сохраняем внутренний ID
			contracts = append(contracts, newContract)
			stats.Updated++
		} else {
			// Без изменений
			stats.Unchanged++
		}
	}

	// Сохраняем все изменения
	if len(contracts) > 0 {
		if err := UpsertContracts(gdb, contracts); err != nil {
			return stats, err
		}
	}

	return stats, nil
}

// Структуры для данных
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
