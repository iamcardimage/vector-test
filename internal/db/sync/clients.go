package sync

import (
	"vector/internal/models"

	"gorm.io/gorm"
)

type ClientListItem struct {
	ClientID          int    `gorm:"column:client_id" json:"id"`
	Surname           string `json:"surname"`
	Name              string `json:"name"`
	Patronymic        string `json:"patronymic"`
	Birthday          string `json:"birthday"`
	BirthPlace        string `json:"birth_place"`
	ContactEmail      string `json:"contact_email"`
	Inn               string `json:"inn"`
	Snils             string `json:"snils"`
	CreatedLKAt       string `json:"created_lk_at"`
	UpdatedLKAt       string `json:"updated_lk_at"`
	PassIssuerCode    string `json:"pass_issuer_code"`
	PassSeries        string `json:"pass_series"`
	PassNumber        string `json:"pass_number"`
	PassIssueDate     string `json:"pass_issue_date"`
	PassIssuer        string `json:"pass_issuer"`
	MainPhone         string `json:"main_phone"`
	ExternalRiskLevel string `gorm:"column:external_risk_level" json:"external_risk_level"`
	NeedsSecondPart   bool   `gorm:"column:needs_second_part" json:"needs_second_part"`
	SecondPartCreated bool   `gorm:"column:second_part_created" json:"second_part_created"`
}

func ListCurrentClients(gdb *gorm.DB, page, perPage int, needsSecondPart *bool) (items []ClientListItem, total int64, err error) {
	q := gdb.Model(&models.ClientVersion{}).
		Where("is_current = ?", true)

	if needsSecondPart != nil {
		q = q.Where("needs_second_part = ?", *needsSecondPart)
	}

	if err = q.Count(&total).Error; err != nil {
		return
	}

	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 500 {
		perPage = 100
	}
	offset := (page - 1) * perPage

	err = q.
		Select([]string{
			"client_id",
			"surname",
			"name",
			"patronymic",
			"birthday",
			"birth_place",
			"contact_email",
			"external_risk_level",
			"needs_second_part",
			"second_part_created",
			"inn",
			"snils",
			"created_lk_at",
			"updated_lk_at",
			"pass_issuer_code",
			"pass_series",
			"pass_number",
			"pass_issue_date",
			"pass_issuer",
			"main_phone",
		}).
		Order("client_id ASC").
		Limit(perPage).
		Offset(offset).
		Scan(&items).Error

	return
}
