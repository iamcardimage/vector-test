// internal/models/sync_models.go - ТОЛЬКО для sync сервиса
package models

// ClientListItem представляет элемент списка клиентов для sync сервиса
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
