package db

import (
	"fmt"
	"log"
	"os"
	"time"
	"vector/internal/models"

	"github.com/joho/godotenv"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

func Connect() (*gorm.DB, error) {
	_ = godotenv.Load()

	host := getenv("DB_HOST", "localhost")
	port := getenv("DB_PORT", "5432")
	user := getenv("POSTGRES_USER", "app")
	pass := getenv("POSTGRES_PASSWORD", "app")
	name := getenv("POSTGRES_DB", "vector")
	ssl := getenv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		host, port, user, pass, name, ssl,
	)

	lg := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold:             time.Second,
		LogLevel:                  logger.Warn, // тише лог
		IgnoreRecordNotFoundError: true,        // скрыть "record not found"
		Colorful:                  true,
	})

	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: lg,
	})
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
func MigrateExternal(gdb *gorm.DB) error {
	return gdb.AutoMigrate(&models.ExternalUser{})
}

func UpsertExternalUsers(gdb *gorm.DB, items []models.ExternalUser) error {
	if len(items) == 0 {
		return nil
	}
	return gdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"surname", "name", "patronymic", "updated_at"}),
	}).Create(&items).Error
}

func MigrateStaging(gdb *gorm.DB) error {
	if err := gdb.Exec("CREATE SCHEMA IF NOT EXISTS staging").Error; err != nil {
		return err
	}
	return gdb.AutoMigrate(&models.StagingExternalUser{})
}

func MigrateCoreClients(gdb *gorm.DB) error {
	if err := gdb.Exec("CREATE SCHEMA IF NOT EXISTS core").Error; err != nil {
		return err
	}
	if err := gdb.AutoMigrate(&models.ClientVersion{}); err != nil {
		return err
	}
	// Один текущий срез на клиента (частичный уникальный индекс)
	return gdb.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_clients_current
		ON core.clients_versions (client_id)
		WHERE is_current = true
	`).Error
}

func UpsertStagingExternalUsers(gdb *gorm.DB, items []models.StagingExternalUser) error {
	if len(items) == 0 {
		return nil
	}
	return gdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"raw", "synced_at"}),
	}).Create(&items).Error
}

type ClientListItem struct {
	ClientID          int    `gorm:"column:client_id" json:"id"`
	Surname           string `json:"surname"`
	Name              string `json:"name"`
	Patronymic        string `json:"patronymic"`
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
			"external_risk_level",
			"needs_second_part",
			"second_part_created",
		}).
		Order("client_id ASC").
		Limit(perPage).
		Offset(offset).
		Scan(&items).Error

	return
}

func MigrateCoreSecondPart(gdb *gorm.DB) error {
	if err := gdb.Exec("CREATE SCHEMA IF NOT EXISTS core").Error; err != nil {
		return err
	}
	if err := gdb.AutoMigrate(&models.SecondPartVersion{}); err != nil {
		return err
	}
	// Один текущий на клиента
	if err := gdb.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_second_part_current
		ON core.second_part_versions (client_id)
		WHERE is_current = true
	`).Error; err != nil {
		return err
	}
	// Ускорить выборку “актуальная 2-я часть на текущей версии клиента”
	return gdb.Exec(`
		CREATE INDEX IF NOT EXISTS idx_sp_client_version_current
		ON core.second_part_versions (client_id, client_version)
		WHERE is_current = true
	`).Error
}

// Текущая версия клиента
func GetClientCurrent(gdb *gorm.DB, id int) (models.ClientVersion, error) {
	var v models.ClientVersion
	err := gdb.Where("client_id = ? AND is_current = true", id).Take(&v).Error
	return v, err
}

// Текущая 2-я часть
func GetSecondPartCurrent(gdb *gorm.DB, id int) (models.SecondPartVersion, error) {
	var sp models.SecondPartVersion
	err := gdb.Where("client_id = ? AND is_current = true", id).Take(&sp).Error
	return sp, err
}

// История 2-й части
func ListSecondPartHistory(gdb *gorm.DB, id int) ([]models.SecondPartVersion, error) {
	var vs []models.SecondPartVersion
	err := gdb.Where("client_id = ?", id).Order("version ASC").Find(&vs).Error
	return vs, err
}

// Создать draft 2-й части (SCD2) с префиллом и, опционально, override Data и risk
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
		// текущий клиент
		curClient, err := GetClientCurrent(tx, clientID)
		if err != nil {
			return err
		}

		// текущая 2-я часть
		var curSP models.SecondPartVersion
		err = tx.Where("client_id = ? AND is_current = true", clientID).Take(&curSP).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		// закрыть текущую 2-ю часть, если была
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

		// итоговый Data: override > префилл
		data := prefill
		if dataOverride != nil {
			data = *dataOverride
		}

		// расчет due_at по risk
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

		// пометить у клиента, что 2-я часть создана
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

// Создать новую версию 2-й части с новым статусом (SCD2), копируя данные из текущей.
// Если текущей 2-й части нет — сначала создаём draft.
func TransitionSecondPartStatus(
	gdb *gorm.DB,
	clientID int,
	newStatus string, // submitted|approved|rejected|doc_requested
	actorID *int,
	reason *string, // для rejected / doc_requested
) (models.SecondPartVersion, error) {
	now := time.Now().UTC()
	var out models.SecondPartVersion

	err := gdb.Transaction(func(tx *gorm.DB) error {
		// Текущий клиент (для привязки client_version)
		curClient, err := GetClientCurrent(tx, clientID)
		if err != nil {
			return err
		}

		// Текущая 2-я часть
		var curSP models.SecondPartVersion
		err = tx.Where("client_id = ? AND is_current = true", clientID).Take(&curSP).Error
		if err == gorm.ErrRecordNotFound {
			// Нет текущей 2-й части — создадим draft, затем будем переходить
			spDraft, err := CreateSecondPartDraft(tx, clientID, nil, actorID, nil)
			if err != nil {
				return err
			}
			curSP = spDraft
		} else if err != nil {
			return err
		}

		// Закрываем текущую
		if err := tx.Model(&models.SecondPartVersion{}).
			Where("client_id = ? AND is_current = true", clientID).
			Updates(map[string]any{
				"is_current": false,
				"valid_to":   now,
			}).Error; err != nil {
			return err
		}

		// Создаём новую версию со сменой статуса, копируя Data/Risk из текущей
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

		// Акторы и причина
		if actorID != nil {
			switch newStatus {
			case "approved":
				next.ApprovedByUserID = actorID
			default:
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

// Миграция таблицы пользователей
func MigrateCoreUsers(gdb *gorm.DB) error {
	if err := gdb.Exec("CREATE SCHEMA IF NOT EXISTS core").Error; err != nil {
		return err
	}
	return gdb.AutoMigrate(&models.AppUser{})
}

// Поиск пользователя по Bearer-токену
func GetUserByToken(gdb *gorm.DB, token string) (models.AppUser, error) {
	var u models.AppUser
	err := gdb.Where("token = ?", token).Take(&u).Error
	return u, err
}

// Сиды (dev): создаём примеров пользователей
func SeedAppUsers(gdb *gorm.DB) error {
	users := []models.AppUser{
		{Email: "admin@example.com", Role: "admin", Token: "admin-token"},
		{Email: "staff@example.com", Role: "staff", Token: "staff-token"},
		{Email: "viewer@example.com", Role: "viewer", Token: "viewer-token"},
	}
	for _, u := range users {
		var existing models.AppUser
		if err := gdb.Where("email = ?", u.Email).First(&existing).Error; err == gorm.ErrRecordNotFound {
			if err := gdb.Create(&u).Error; err != nil {
				return err
			}
		}
	}
	return nil
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

	// count
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
