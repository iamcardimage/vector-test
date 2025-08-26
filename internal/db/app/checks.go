package app

import (
	"time"
	"vector/internal/models"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ===== CHECKS SYSTEM =====

// CreateSecondPartCheck создает новую проверку
func CreateSecondPartCheck(gdb *gorm.DB, clientID, spVersion int, kind string, payload *datatypes.JSON, runBy *int) (models.SecondPartCheck, error) {
	ch := models.SecondPartCheck{
		ClientID:          clientID,
		SecondPartVersion: spVersion,
		Kind:              kind,
		Status:            "pending",
		Payload:           datatypes.JSON([]byte(`{}`)),
		RunAt:             time.Now().UTC(),
		RunByUserID:       runBy,
	}
	if payload != nil {
		ch.Payload = *payload
	}
	return ch, gdb.Create(&ch).Error
}

// UpdateSecondPartCheckResult обновляет результат проверки
func UpdateSecondPartCheckResult(gdb *gorm.DB, checkID uint, status string, result *datatypes.JSON) (models.SecondPartCheck, error) {
	var ch models.SecondPartCheck
	if err := gdb.Where("id = ?", checkID).Take(&ch).Error; err != nil {
		return ch, err
	}
	now := time.Now().UTC()
	ch.Status = status
	ch.FinishedAt = &now
	if result != nil {
		ch.Result = *result
	}
	return ch, gdb.Save(&ch).Error
}

// ListSecondPartChecks возвращает список проверок для клиента
func ListSecondPartChecks(gdb *gorm.DB, clientID int, spVersion *int) ([]models.SecondPartCheck, error) {
	var xs []models.SecondPartCheck
	q := gdb.Where("client_id = ?", clientID)
	if spVersion != nil {
		q = q.Where("second_part_version = ?", *spVersion)
	}
	return xs, q.Order("id DESC").Find(&xs).Error
}
