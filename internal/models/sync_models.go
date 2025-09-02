package models

import "time"

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
	Version           int    `json:"version"`
}

type SecondPartResponse struct {
	ClientVersion int        `json:"client_version" example:"1"`
	Version       int        `json:"version" example:"3"`
	Status        string     `json:"status" example:"draft"`
	RiskLevel     string     `json:"risk_level" example:""`
	DueAt         *time.Time `json:"due_at"`
	IsCurrent     bool       `json:"is_current" example:"true"`
	Stale         bool       `json:"stale,omitempty"`
}
type ClientDetailResponse struct {
	ClientListItem
	SecondPart *SecondPartResponse `json:"second_part,omitempty"`
}

type GetClientResponse struct {
	ID            int    `json:"id" example:"12345"`
	ClientID      int    `json:"client_id" example:"12345"`
	Version       int    `json:"version" example:"1"`
	ExternalID    int    `json:"external_id" example:"54321"`
	ExternalIDStr string `json:"external_id_str" example:"EXT_12345"`

	Surname    string `json:"surname" example:"Иванов"`
	Name       string `json:"name" example:"Иван"`
	Patronymic string `json:"patronymic" example:"Иванович"`

	Birthday   string `json:"birthday" example:"1990-01-01"`
	BirthPlace string `json:"birth_place" example:"Test City"`

	Inn             string `json:"inn" example:"123456789012"`
	Snils           string `json:"snils" example:"12345678901"`
	PassSeries      string `json:"pass_series" example:"1234"`
	PassNumber      string `json:"pass_number" example:"123456"`
	PassIssueDate   string `json:"pass_issue_date" example:"01.01.2020"`
	PassIssuer      string `json:"pass_issuer" example:"MVD Test District"`
	PassIssuerCode  string `json:"pass_issuer_code" example:"100-001"`
	DocumentType    string `json:"document_type" example:"passport"`
	DocumentCountry string `json:"document_country" example:"RU"`

	ContactEmail string `json:"contact_email" example:"test@example.com"`
	MainPhone    string `json:"main_phone" example:"1234567890"`
	Login        string `json:"login" example:"user_login"`

	Blocked       *bool  `json:"blocked,omitempty" example:"false"`
	BlockedReason string `json:"blocked_reason,omitempty" example:""`
	BlockType     string `json:"block_type,omitempty" example:""`

	Male               *bool `json:"male,omitempty" example:"true"`
	IsRfResident       *bool `json:"is_rf_resident,omitempty" example:"true"`
	IsRfTaxpayer       *bool `json:"is_rf_taxpayer,omitempty" example:"true"`
	IsValidInfo        *bool `json:"is_valid_info,omitempty" example:"true"`
	QualifiedInvestor  *bool `json:"qualified_investor,omitempty" example:"false"`
	IsFilled           *bool `json:"is_filled,omitempty" example:"true"`
	IsAmericanNational *bool `json:"is_american_national,omitempty" example:"false"`
	NeedToSetPassword  *bool `json:"need_to_set_password,omitempty" example:"false"`

	LegalCapacity      string `json:"legal_capacity,omitempty" example:"full"`
	PifsPortfolioCode  *int   `json:"pifs_portfolio_code,omitempty" example:"123"`
	RiskLevel          string `json:"risk_level,omitempty" example:"low"`
	ExternalRiskLevel  string `json:"external_risk_level,omitempty" example:"low"`
	FillStage          string `json:"fill_stage,omitempty" example:"completed"`
	IdentificationType string `json:"identification_type,omitempty" example:"passport"`
	TaxStatus          string `json:"tax_status,omitempty" example:"resident"`

	EsiaID       *int `json:"esia_id,omitempty" example:"98765"`
	AgentID      *int `json:"agent_id,omitempty" example:"456"`
	AgentPointID *int `json:"agent_point_id,omitempty" example:"789"`

	Country  string `json:"country,omitempty" example:"Россия"`
	Region   string `json:"region,omitempty" example:"Московская область"`
	Index    *int   `json:"index,omitempty" example:"123456"`
	City     string `json:"city,omitempty" example:"Москва"`
	Street   string `json:"street,omitempty" example:"ул. Тверская"`
	House    string `json:"house,omitempty" example:"10"`
	Corps    *int   `json:"corps,omitempty" example:"1"`
	Flat     *int   `json:"flat,omitempty" example:"25"`
	District string `json:"district,omitempty" example:"Тверской"`

	SignatureType              string     `json:"signature_type,omitempty" example:"electronic"`
	DataReceivedDigitalProfile *bool      `json:"data_received_digital_profile,omitempty" example:"true"`
	LockedAt                   *time.Time `json:"locked_at,omitempty" swaggertype:"string" format:"date-time"`
	CurrentSignInAt            *time.Time `json:"current_sign_in_at,omitempty" swaggertype:"string" format:"date-time"`
	SignInCount                *int       `json:"sign_in_count,omitempty" example:"5"`

	CreatedLKAt string     `json:"created_lk_at" example:"2024-01-01T10:00:00.000+03:00"`
	UpdatedLKAt string     `json:"updated_lk_at" example:"2024-01-01T12:00:00.000+03:00"`
	SyncedAt    time.Time  `json:"synced_at" swaggertype:"string" format:"date-time"`
	ValidFrom   time.Time  `json:"valid_from" swaggertype:"string" format:"date-time"`
	ValidTo     *time.Time `json:"valid_to,omitempty" swaggertype:"string" format:"date-time"`
	IsCurrent   bool       `json:"is_current" example:"true"`

	FromCompanySettings     *map[string]interface{} `json:"from_company_settings,omitempty"`
	Settings                *map[string]interface{} `json:"settings,omitempty"`
	PersonInfo              *map[string]interface{} `json:"person_info,omitempty"`
	Manager                 *map[string]interface{} `json:"manager,omitempty"`
	Checks                  *map[string]interface{} `json:"checks,omitempty"`
	Note                    *map[string]interface{} `json:"note,omitempty"`
	AdSource                *map[string]interface{} `json:"ad_source,omitempty"`
	SignatureAllowedNumbers *map[string]interface{} `json:"signature_allowed_numbers,omitempty"`
	Raw                     *map[string]interface{} `json:"raw,omitempty"`

	Hash                  string `json:"hash" example:"abc123def456"`
	Status                string `json:"status" example:"unchanged"`
	SecondPartTriggerHash string `json:"second_part_trigger_hash" example:"xyz789abc123"`

	NeedsSecondPart   bool `json:"needs_second_part" example:"true"`
	SecondPartCreated bool `json:"second_part_created" example:"true"`
	SecondPart        *struct {
		ClientVersion int        `json:"client_version" example:"1"`
		Version       int        `json:"version" example:"1"`
		Status        string     `json:"status" example:"draft"`
		RiskLevel     string     `json:"risk_level" example:"low"`
		IsCurrent     bool       `json:"is_current" example:"true"`
		DueAt         *time.Time `json:"due_at" swaggertype:"string" format:"date-time"`
	} `json:"second_part,omitempty"`
}

type GetContractResponse struct {
	ID         int `json:"id" example:"123"`
	ExternalID int `json:"external_id" example:"54321"`
	UserID     int `json:"user_id" example:"12345"`

	InnerCode         string  `json:"inner_code" example:"BRK-001"`
	Kind              string  `json:"kind" example:"broking"`
	Status            string  `json:"status" example:"active"`
	ContractOwnerType string  `json:"contract_owner_type" example:"individual"`
	ContractOwnerID   int     `json:"contract_owner_id" example:"789"`
	Comment           *string `json:"comment,omitempty" example:"Основной договор"`

	IsPersonalInvestAccount    bool `json:"is_personal_invest_account" example:"true"`
	IsPersonalInvestAccountNew bool `json:"is_personal_invest_account_new" example:"false"`

	RialtoCode *string `json:"rialto_code,omitempty" example:"RLT123"`
	Anketa     *string `json:"anketa,omitempty" example:"ANK456"`
	OwnerID    *int    `json:"owner_id,omitempty" example:"111"`
	UserLogin  *string `json:"user_login,omitempty" example:"user123"`

	CalculatedProfileID *int    `json:"calculated_profile_id,omitempty" example:"222"`
	DepoAccountsType    *string `json:"depo_accounts_type,omitempty" example:"standard"`
	StrategyID          *int    `json:"strategy_id,omitempty" example:"333"`
	StrategyName        *string `json:"strategy_name,omitempty" example:"Conservative"`
	TariffID            *int    `json:"tariff_id,omitempty" example:"444"`
	TariffName          *string `json:"tariff_name,omitempty" example:"Basic"`

	CreatedAt time.Time  `json:"created_at" swaggertype:"string" format:"date-time"`
	UpdatedAt time.Time  `json:"updated_at" swaggertype:"string" format:"date-time"`
	SignedAt  *time.Time `json:"signed_at,omitempty" swaggertype:"string" format:"date-time"`
	ClosedAt  *time.Time `json:"closed_at,omitempty" swaggertype:"string" format:"date-time"`
	SyncedAt  time.Time  `json:"synced_at" swaggertype:"string" format:"date-time"`

	Hash string                  `json:"hash" example:"abc123def456"`
	Raw  *map[string]interface{} `json:"raw,omitempty"`
}

type ListContractsResponse struct {
	Success    bool                  `json:"success" example:"true"`
	Contracts  []GetContractResponse `json:"contracts"`
	Page       int                   `json:"page" example:"1"`
	PerPage    int                   `json:"per_page" example:"10"`
	Total      int64                 `json:"total" example:"150"`
	TotalPages int                   `json:"total_pages" example:"15"`
}

type ListClientsResponse struct {
	Success    bool                   `json:"success" example:"true"`
	Clients    []ClientDetailResponse `json:"clients"`
	Page       int                    `json:"page" example:"1"`
	PerPage    int                    `json:"per_page" example:"10"`
	Total      int64                  `json:"total" example:"150"`
	TotalPages int                    `json:"total_pages" example:"15"`
}

type GetSecondPartResponse struct {
	ClientID         int                     `json:"client_id" example:"123"`
	ClientVersion    int                     `json:"client_version" example:"1"`
	Version          int                     `json:"version" example:"2"`
	Status           string                  `json:"status" example:"draft"`
	RiskLevel        string                  `json:"risk_level" example:"low"`
	IsCurrent        bool                    `json:"is_current" example:"true"`
	DueAt            *time.Time              `json:"due_at,omitempty" swaggertype:"string" format:"date-time"`
	ValidFrom        time.Time               `json:"valid_from" swaggertype:"string" format:"date-time"`
	ValidTo          *time.Time              `json:"valid_to,omitempty" swaggertype:"string" format:"date-time"`
	Data             *map[string]interface{} `json:"data,omitempty"`
	Reason           string                  `json:"reason,omitempty" example:"Additional documents required"`
	CreatedByUserID  *int                    `json:"created_by_user_id,omitempty" example:"456"`
	UpdatedByUserID  *int                    `json:"updated_by_user_id,omitempty" example:"789"`
	ApprovedByUserID *int                    `json:"approved_by_user_id,omitempty" example:"101"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"client not found"`
}

type ClientHistoryResponse struct {
	Success  bool                `json:"success" example:"true"`
	History  []GetClientResponse `json:"history"`
	Total    int                 `json:"total" example:"5"`
	ClientID int                 `json:"client_id" example:"123"`
}

type ClientVersionSummary struct {
	Version        int        `json:"version" example:"3"`
	IsCurrent      bool       `json:"is_current" example:"true"`
	ValidFrom      time.Time  `json:"valid_from" swaggertype:"string" format:"date-time"`
	ValidTo        *time.Time `json:"valid_to,omitempty" swaggertype:"string" format:"date-time"`
	SyncedAt       time.Time  `json:"synced_at" swaggertype:"string" format:"date-time"`
	Status         string     `json:"status" example:"changed"`
	ChangesSummary []string   `json:"changes_summary,omitempty" example:"contact_email,address"`
}

type ClientVersionsListResponse struct {
	Success  bool                   `json:"success" example:"true"`
	Versions []ClientVersionSummary `json:"versions"`
	Total    int                    `json:"total" example:"5"`
	ClientID int                    `json:"client_id" example:"123"`
}

type GetClientVersionResponse struct {
	Success  bool              `json:"success" example:"true"`
	Version  GetClientResponse `json:"version"`
	ClientID int               `json:"client_id" example:"123"`
}
