package utils

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
	"vector/internal/models"

	"gorm.io/datatypes"
)

func ExtractUserID(raw json.RawMessage) (int, error) {
	var tmp map[string]any
	if err := json.Unmarshal(raw, &tmp); err != nil {
		return 0, err
	}
	idVal, ok := tmp["id"]
	if !ok {
		return 0, errors.New("id field not found")
	}
	switch v := idVal.(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	default:
		return 0, errors.New("invalid id type")
	}
}

// НОВАЯ ФУНКЦИЯ: полный парсинг ClientVersion
func ParseClientVersion(raw json.RawMessage) models.ClientVersion {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return models.ClientVersion{Raw: datatypes.JSON(raw)}
	}

	client := models.ClientVersion{
		Raw: datatypes.JSON(raw), // Всегда сохраняем полные данные
	}

	// Основные поля
	client.Surname = ExtractString(m, "surname")
	client.Name = ExtractString(m, "name")
	client.Patronymic = ExtractString(m, "patronymic")
	client.Birthday = ExtractString(m, "birthday")
	client.BirthPlace = ExtractString(m, "birth_place")
	client.ContactEmail = ExtractString(m, "contact_email")
	client.Inn = ExtractString(m, "inn")
	client.Snils = ExtractString(m, "snils")
	client.CreatedLKAt = ExtractString(m, "created_at")
	client.UpdatedLKAt = ExtractString(m, "updated_at")
	client.PassIssuerCode = ExtractString(m, "pass_issuer_code")
	client.PassSeries = ExtractString(m, "pass_series")
	client.PassNumber = ExtractString(m, "pass_number")
	client.PassIssueDate = ExtractString(m, "pass_issue_date")
	client.PassIssuer = ExtractString(m, "pass_issuer")
	client.MainPhone = ExtractString(m, "main_phone")

	// Новые поля
	client.ID = ExtractInt(m, "id")
	client.Login = ExtractString(m, "login")
	client.LockedAt = ExtractTimePtr(m, "locked_at")
	client.CurrentSignInAt = ExtractTimePtr(m, "current_sign_in_at")
	client.SignInCount = ExtractIntPtr(m, "sign_in_count")
	client.NeedToSetPassword = ExtractBoolPtr(m, "need_to_set_password")
	client.Blocked = ExtractBoolPtr(m, "blocked")
	client.BlockedReason = ExtractString(m, "blocked_reason")
	client.BlockType = ExtractString(m, "block_type")
	client.Male = ExtractBoolPtr(m, "male")
	client.IsRfResident = ExtractBoolPtr(m, "is_rf_resident")
	client.DocumentType = ExtractString(m, "document_type")
	client.DocumentCountry = ExtractString(m, "document_country")
	client.LegalCapacity = ExtractString(m, "legal_capacity")
	client.IsRfTaxpayer = ExtractBoolPtr(m, "is_rf_taxpayer")
	client.PifsPortfolioCode = ExtractIntPtr(m, "pifs_portfolio_code")
	client.ExternalIDStr = ExtractString(m, "external_id")
	client.IsValidInfo = ExtractBoolPtr(m, "is_valid_info")
	client.QualifiedInvestor = ExtractBoolPtr(m, "qualified_investor")
	client.RiskLevel = ExtractString(m, "risk_level")
	client.FillStage = ExtractString(m, "fill_stage")
	client.IsFilled = ExtractBoolPtr(m, "is_filled")
	client.EsiaID = ExtractIntPtr(m, "esia_id")
	client.IdentificationType = ExtractString(m, "identification_type")
	client.AgentID = ExtractIntPtr(m, "agent_id")
	client.AgentPointID = ExtractIntPtr(m, "agent_point_id")
	client.TaxStatus = ExtractString(m, "tax_status")
	client.IsAmericanNational = ExtractBoolPtr(m, "is_american_national")
	client.Country = ExtractString(m, "country")
	client.Region = ExtractString(m, "region")
	client.Index = ExtractIntPtr(m, "index")
	client.City = ExtractString(m, "city")
	client.Street = ExtractString(m, "street")
	client.House = ExtractString(m, "house")
	client.Corps = ExtractIntPtr(m, "corps")
	client.Flat = ExtractIntPtr(m, "flat")
	client.District = ExtractString(m, "district")
	client.SignatureType = ExtractString(m, "signature_type")
	client.DataReceivedDigitalProfile = ExtractBoolPtr(m, "data_recieved_digital_profile")

	// JSON поля
	client.FromCompanySettings = ExtractJSONField(m, "from_company_settings")
	client.Settings = ExtractJSONField(m, "settings")
	client.PersonInfo = ExtractJSONField(m, "person_info")
	client.Manager = ExtractJSONField(m, "manager")
	client.Checks = ExtractJSONField(m, "checks")
	client.Note = ExtractJSONField(m, "note")
	client.AdSource = ExtractJSONField(m, "ad_source")
	client.SignatureAllowedNumbers = ExtractJSONField(m, "signature_allowed_numbers")

	// Риск уровень из внешней системы
	client.ExternalRiskLevel = client.RiskLevel

	return client
}

// Вспомогательные функции
func ExtractString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return strings.TrimSpace(s)
		}
	}

	// Проверяем в person_info
	if pi, ok := m["person_info"].(map[string]any); ok {
		if v, ok := pi[key]; ok {
			if s, ok := v.(string); ok {
				return strings.TrimSpace(s)
			}
		}
	}
	return ""
}

func ExtractInt(m map[string]any, key string) int {
	if v, ok := m[key]; ok {
		switch t := v.(type) {
		case float64:
			return int(t)
		case int:
			return t
		case string:
			if i, err := strconv.Atoi(t); err == nil {
				return i
			}
		}
	}
	return 0
}

func ExtractIntPtr(m map[string]any, key string) *int {
	if v, ok := m[key]; ok && v != nil {
		switch t := v.(type) {
		case float64:
			i := int(t)
			return &i
		case int:
			return &t
		case string:
			if i, err := strconv.Atoi(t); err == nil {
				return &i
			}
		}
	}
	return nil
}

func ExtractBoolPtr(m map[string]any, key string) *bool {
	if v, ok := m[key]; ok && v != nil {
		switch t := v.(type) {
		case bool:
			return &t
		case string:
			if b, err := strconv.ParseBool(t); err == nil {
				return &b
			}
		}
	}
	return nil
}

func ExtractTimePtr(m map[string]any, key string) *time.Time {
	if str := ExtractString(m, key); str != "" {
		// Пробуем разные форматы
		formats := []string{
			"2006-01-02T15:04:05.000Z07:00",
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
		}

		for _, format := range formats {
			if t, err := time.Parse(format, str); err == nil {
				return &t
			}
		}
	}
	return nil
}

func ExtractJSONField(m map[string]any, key string) datatypes.JSON {
	if val, exists := m[key]; exists && val != nil {
		if jsonData, err := json.Marshal(val); err == nil {
			return datatypes.JSON(jsonData)
		}
	}
	return nil
}
