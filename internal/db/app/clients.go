package app

import (
	"strings"
	"time"
	"vector/internal/models"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func GetClientCurrent(gdb *gorm.DB, id int) (models.ClientVersion, error) {
	var v models.ClientVersion
	err := gdb.Where("client_id = ? AND is_current = true", id).Take(&v).Error
	return v, err
}

func GetSecondPartCurrent(gdb *gorm.DB, id int) (models.SecondPartVersion, error) {
	var sp models.SecondPartVersion
	err := gdb.Where("client_id = ? AND is_current = true", id).Take(&sp).Error
	return sp, err
}

func ListSecondPartHistory(gdb *gorm.DB, id int) ([]models.SecondPartVersion, error) {
	var vs []models.SecondPartVersion
	err := gdb.Where("client_id = ?", id).Order("version ASC").Find(&vs).Error
	return vs, err
}

func CreateSecondPartDraft(
	gdb *gorm.DB,
	clientID int,
	riskLevel *string,
	createdBy *int,
	dataOverride *datatypes.JSON,
) (models.SecondPartVersion, error) {
	now := time.Now().UTC()
	var out models.SecondPartVersion

	err := gdb.Transaction(func(tx *gorm.DB) error {

		curClient, err := GetClientCurrent(tx, clientID)
		if err != nil {
			return err
		}

		var curSP models.SecondPartVersion
		err = tx.Where("client_id = ? AND is_current = true", clientID).Take(&curSP).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		nextVersion := 1
		prefill := datatypes.JSON([]byte(`{}`))
		if err == nil {
			nextVersion = curSP.Version + 1
			prefill = curSP.Data
			if err := tx.Model(&models.SecondPartVersion{}).
				Where("client_id = ? AND is_current = true", clientID).
				Updates(map[string]any{
					"is_current": false,
					"valid_to":   now,
				}).Error; err != nil {
				return err
			}
		}

		data := prefill
		if dataOverride != nil {
			data = *dataOverride
		}

		var dueAt *time.Time
		var risk string
		if riskLevel != nil && *riskLevel != "" {
			risk = *riskLevel
			years := 0
			if risk == "low" {
				years = 3
			} else {
				years = 1
			}
			t := now.AddDate(years, 0, 0)
			dueAt = &t
		}

		sp := models.SecondPartVersion{
			ClientID:      clientID,
			ClientVersion: curClient.Version,
			Version:       nextVersion,
			IsCurrent:     true,
			ValidFrom:     now,
			Status:        "draft",
			Data:          data,
			RiskLevel:     risk,
			DueAt:         dueAt,
		}
		if createdBy != nil {
			sp.CreatedByUserID = createdBy
		}

		if err := tx.Create(&sp).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.ClientVersion{}).
			Where("client_id = ? AND is_current = true", clientID).
			Update("second_part_created", true).Error; err != nil {
			return err
		}

		out = sp
		return nil
	})
	return out, err
}

func TransitionSecondPartStatus(
	gdb *gorm.DB,
	clientID int,
	newStatus string, // submitted|approved|rejected|doc_requested
	actorID *int,
	reason *string,
) (models.SecondPartVersion, error) {
	now := time.Now().UTC()
	var out models.SecondPartVersion

	err := gdb.Transaction(func(tx *gorm.DB) error {

		curClient, err := GetClientCurrent(tx, clientID)
		if err != nil {
			return err
		}

		var curSP models.SecondPartVersion
		err = tx.Where("client_id = ? AND is_current = true", clientID).Take(&curSP).Error
		if err == gorm.ErrRecordNotFound {
			spDraft, err := CreateSecondPartDraft(tx, clientID, nil, actorID, nil)
			if err != nil {
				return err
			}
			curSP = spDraft
		} else if err != nil {
			return err
		}

		if err := tx.Model(&models.SecondPartVersion{}).
			Where("client_id = ? AND is_current = true", clientID).
			Updates(map[string]any{
				"is_current": false,
				"valid_to":   now,
			}).Error; err != nil {
			return err
		}

		next := models.SecondPartVersion{
			ClientID:      clientID,
			ClientVersion: curClient.Version,
			Version:       curSP.Version + 1,
			IsCurrent:     true,
			ValidFrom:     now,
			Status:        newStatus,
			Data:          curSP.Data,
			RiskLevel:     curSP.RiskLevel,
			DueAt:         curSP.DueAt,
			Reason:        "",
		}

		if newStatus == "approved" {
			years := 1
			if strings.ToLower(curSP.RiskLevel) == "low" {
				years = 3
			}
			t := now.AddDate(years, 0, 0)
			next.DueAt = &t

			if err := tx.Model(&models.ClientVersion{}).
				Where("client_id = ? AND is_current = true", clientID).
				Update("needs_second_part", false).Error; err != nil {
				return err
			}
		}

		if actorID != nil {
			if newStatus == "approved" {
				next.ApprovedByUserID = actorID
			} else {
				next.UpdatedByUserID = actorID
			}
		}
		if reason != nil {
			next.Reason = *reason
		}

		if err := tx.Create(&next).Error; err != nil {
			return err
		}

		out = next
		return nil
	})

	return out, err
}

func SubmitSecondPart(gdb *gorm.DB, clientID int, userID *int) (models.SecondPartVersion, error) {
	return TransitionSecondPartStatus(gdb, clientID, "submitted", userID, nil)
}

func ApproveSecondPart(gdb *gorm.DB, clientID int, approvedBy *int) (models.SecondPartVersion, error) {
	return TransitionSecondPartStatus(gdb, clientID, "approved", approvedBy, nil)
}

func RejectSecondPart(gdb *gorm.DB, clientID int, userID *int, reason string) (models.SecondPartVersion, error) {
	return TransitionSecondPartStatus(gdb, clientID, "rejected", userID, &reason)
}

func RequestDocsSecondPart(gdb *gorm.DB, clientID int, userID *int, reason string) (models.SecondPartVersion, error) {
	return TransitionSecondPartStatus(gdb, clientID, "doc_requested", userID, &reason)
}

type ClientWithSP struct {
	ClientID          int    `gorm:"column:client_id" json:"id"`
	Surname           string `json:"surname"`
	Name              string `json:"name"`
	Patronymic        string `json:"patronymic"`
	ExternalRiskLevel string `gorm:"column:external_risk_level" json:"external_risk_level"`
	NeedsSecondPart   bool   `gorm:"column:needs_second_part" json:"needs_second_part"`
	SecondPartCreated bool   `gorm:"column:second_part_created" json:"second_part_created"`

	ClientVersion   int        `gorm:"column:client_version" json:"client_version"`
	SpStatus        *string    `gorm:"column:sp_status" json:"sp_status,omitempty"`
	SpDueAt         *time.Time `gorm:"column:sp_due_at" json:"sp_due_at,omitempty"`
	SpClientVersion *int       `gorm:"column:sp_client_version" json:"sp_client_version,omitempty"`
}

func ListClientsWithSP(
	gdb *gorm.DB,
	page, perPage int,
	needsSecondPart *bool,
	spStatus *string,
	dueBefore *time.Time,
) (items []ClientWithSP, total int64, err error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 500 {
		perPage = 100
	}
	offset := (page - 1) * perPage

	base := gdb.Table("core.clients_versions AS c").
		Select(`
			c.client_id,
			c.surname, c.name, c.patronymic,
			c.external_risk_level,
			c.needs_second_part,
			c.second_part_created,
			c.version AS client_version,
			sp.status AS sp_status,
			sp.due_at AS sp_due_at,
			sp.client_version AS sp_client_version
		`).
		Joins(`
			LEFT JOIN core.second_part_versions AS sp
				ON sp.client_id = c.client_id
				AND sp.is_current = true
		`).
		Where("c.is_current = true")

	if needsSecondPart != nil {
		base = base.Where("c.needs_second_part = ?", *needsSecondPart)
	}
	if spStatus != nil && *spStatus != "" {
		base = base.Where("sp.status = ?", *spStatus)
	}
	if dueBefore != nil {
		base = base.Where("sp.due_at IS NOT NULL AND sp.due_at <= ?", *dueBefore)
	}

	if err = gdb.Table("(?) AS sub", base.Session(&gorm.Session{NewDB: true})).Count(&total).Error; err != nil {
		return
	}

	err = base.
		Order("c.client_id ASC").
		Limit(perPage).
		Offset(offset).
		Scan(&items).Error
	return
}
