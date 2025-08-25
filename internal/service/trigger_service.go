package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type TriggerService struct{}

func NewTriggerService() *TriggerService {
	return &TriggerService{}
}

// Поля, изменение которых должно инициировать вторую часть
var triggerFields = []string{
	"main_phone", "name", "surname", "qualified_investor", "birthday",
	"contact_email", "patronymic", "male", "birth_place", "inn", "snils",
	"legal_capacity", "pass_series", "pass_number", "pass_issue_date",
	"pass_issuer", "pass_issuer_code", "is_rf_taxpayer", "pifs_portfolio_code",
	"actuality_updated_at", "tax_status", "is_american_national",
	"country", "city", "street", "house", "corps", "flat", "region", "district",
	"residential_country", "residential_index", "residential_city", "residential_street",
	"residential_house", "residential_corps", "residential_flat", "residential_region",
	"residential_district", "for_corresp_country", "for_corresp_index", "for_corresp_city",
	"for_corresp_street", "for_corresp_house", "for_corresp_corps", "for_corresp_flat",
	"for_corresp_region", "for_corresp_district",
}

// Пути к вложенным полям из блока addresses.*
var nestedAddressPaths = map[string][]string{
	"residential_country":  {"addresses", "residential", "country"},
	"residential_index":    {"addresses", "residential", "index"},
	"residential_city":     {"addresses", "residential", "city"},
	"residential_street":   {"addresses", "residential", "street"},
	"residential_house":    {"addresses", "residential", "house"},
	"residential_corps":    {"addresses", "residential", "corps"},
	"residential_flat":     {"addresses", "residential", "flat"},
	"residential_region":   {"addresses", "residential", "region"},
	"residential_district": {"addresses", "residential", "district"},
	"for_corresp_country":  {"addresses", "for_corresp", "country"},
	"for_corresp_index":    {"addresses", "for_corresp", "index"},
	"for_corresp_city":     {"addresses", "for_corresp", "city"},
	"for_corresp_street":   {"addresses", "for_corresp", "street"},
	"for_corresp_house":    {"addresses", "for_corresp", "house"},
	"for_corresp_corps":    {"addresses", "for_corresp", "corps"},
	"for_corresp_flat":     {"addresses", "for_corresp", "flat"},
	"for_corresp_region":   {"addresses", "for_corresp", "region"},
	"for_corresp_district": {"addresses", "for_corresp", "district"},
}

func (s *TriggerService) ComputeSecondPartTriggerHash(raw []byte) (string, error) {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return "", err
	}

	var b strings.Builder
	for _, key := range triggerFields {
		var (
			val any
			ok  bool
		)

		// 1) Вложенные адреса по карте путей
		if path, has := nestedAddressPaths[key]; has {
			if v, okp := getByPath(m, path); okp {
				val, ok = v, true
			}
		}

		// 2) Плоское поле верхнего уровня
		if !ok {
			if v, okTop := m[key]; okTop {
				val, ok = v, true
			}
		}

		// 3) Дубль внутри person_info.{key}
		if !ok {
			if v, okPI := getByPath(m, []string{"person_info", key}); okPI {
				val, ok = v, true
			}
		}

		// Нормализуем к строке
		s := ""
		if ok {
			s = toString(val)
		}

		// Конкатенация с ключом для стабильности
		b.WriteString(key)
		b.WriteString("=")
		b.WriteString(s)
		b.WriteString("|")
	}

	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:]), nil
}

func (s *TriggerService) ExtractExternalRiskLevel(raw []byte) string {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return ""
	}
	// обычно в корне
	if v, ok := m["risk_level"]; ok {
		return toString(v)
	}
	// запасной вариант: вдруг вложили куда-то
	if v, ok := getByPath(m, []string{"person_info", "risk_level"}); ok {
		return toString(v)
	}
	return ""
}

// Вспомогательные функции
func toString(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(t)
	case bool:
		if t {
			return "true"
		}
		return "false"
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

func getByPath(m map[string]any, path []string) (any, bool) {
	var cur any = m
	for _, p := range path {
		asMap, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = asMap[p]
		if !ok {
			return nil, false
		}
	}
	return cur, true
}
