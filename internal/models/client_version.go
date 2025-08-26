package models

import (
	"time"

	"gorm.io/datatypes"
)

type ClientVersion struct {
	// Основные поля (как было)
	ClientID       int `gorm:"not null;index"`
	Version        int `gorm:"not null"`
	Surname        string
	Name           string
	Patronymic     string
	Birthday       string
	BirthPlace     string
	ContactEmail   string
	Inn            string
	Snils          string
	CreatedLKAt    string
	UpdatedLKAt    string
	PassIssuerCode string
	PassSeries     string
	PassNumber     string
	PassIssueDate  string
	PassIssuer     string
	MainPhone      string

	// =================================
	// НОВЫЕ ПОЛЯ - КОРНЕВОЙ УРОВЕНЬ
	// =================================

	// Идентификация и авторизация
	ID                int    `gorm:"column:external_id;index"` // ID из внешней системы
	Login             string `gorm:"index"`
	LockedAt          *time.Time
	CurrentSignInAt   *time.Time
	SignInCount       *int
	NeedToSetPassword *bool

	// Статус и блокировки
	Blocked       *bool
	BlockedReason string
	BlockType     string

	// Личные данные
	Male               *bool
	IsRfResident       *bool
	DocumentType       string
	DocumentCountry    string
	LegalCapacity      string
	IsRfTaxpayer       *bool
	PifsPortfolioCode  *int
	ExternalIDStr      string `gorm:"column:external_id_str;index"` // строковый external_id
	IsValidInfo        *bool
	QualifiedInvestor  *bool
	RiskLevel          string `gorm:"index"` // "low"/"medium"/"high"
	FillStage          string
	IsFilled           *bool
	EsiaID             *int
	IdentificationType string
	AgentID            *int
	AgentPointID       *int
	TaxStatus          string
	IsAmericanNational *bool

	// Адрес (из корневого уровня)
	Country  string
	Region   string
	Index    *int
	City     string
	Street   string
	House    string
	Corps    *int
	Flat     *int
	District string

	// Подписи и цифровой профиль
	SignatureType              string
	DataReceivedDigitalProfile *bool

	// JSON поля для сложных структур
	FromCompanySettings     datatypes.JSON `gorm:"type:jsonb"` // from_company_settings
	Settings                datatypes.JSON `gorm:"type:jsonb"` // settings
	PersonInfo              datatypes.JSON `gorm:"type:jsonb"` // person_info целиком
	Manager                 datatypes.JSON `gorm:"type:jsonb"` // manager
	Checks                  datatypes.JSON `gorm:"type:jsonb"` // checks целиком
	Note                    datatypes.JSON `gorm:"type:jsonb"` // note
	AdSource                datatypes.JSON `gorm:"type:jsonb"` // ad_source
	SignatureAllowedNumbers datatypes.JSON `gorm:"type:jsonb"` // массив номеров

	// риск из внешней базы (НЕ тот, что во второй части)
	ExternalRiskLevel string

	// хэш только по списку триггер-полей (для решения о 2-й части)
	SecondPartTriggerHash string `gorm:"not null"`

	NeedsSecondPart   bool           `gorm:"not null"`
	SecondPartCreated bool           `gorm:"not null"`
	Hash              string         `gorm:"not null"` // общий версионный хэш
	Status            string         // unchanged/changed
	Raw               datatypes.JSON `gorm:"type:jsonb"` // ВСЕ данные как есть
	SyncedAt          time.Time      `gorm:"not null"`
	ValidFrom         time.Time      `gorm:"not null"`
	ValidTo           *time.Time
	IsCurrent         bool `gorm:"not null;index"`
}

func (ClientVersion) TableName() string {
	return "core.clients_versions"
}
