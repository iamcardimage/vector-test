package db

import "gorm.io/gorm"

// RecalcNeedsSecondPart: due_at просрочен или 2-я часть не к текущей версии клиента
func RecalcNeedsSecondPart(gdb *gorm.DB) (int64, error) {
	res := gdb.Exec(`
		UPDATE core.clients_versions AS c
		SET needs_second_part = true
		FROM core.second_part_versions AS sp
		WHERE c.is_current = true
		  AND sp.is_current = true
		  AND sp.client_id = c.client_id
		  AND (
			 (sp.due_at IS NOT NULL AND sp.due_at <= NOW())
		   OR (sp.client_version <> c.version)
		  )
		  AND c.needs_second_part = false
	`)
	return res.RowsAffected, res.Error
}

// RecalcPassportExpiry: по дате рождения клиент переходит порог 20 или 45 лет → нужна 2-я часть
// День рождения парсим из c.raw (jsonb): raw->>'birthday' или raw->'person_info'->>'birthday' в формате DD.MM.YYYY
func RecalcPassportExpiry(gdb *gorm.DB) (int64, error) {
	res := gdb.Exec(`
		WITH birthdays AS (
		  SELECT
		    c.client_id,
		    c.is_current,
		    c.needs_second_part,
		    COALESCE(
		      NULLIF(c.raw->>'birthday',''),
		      NULLIF(c.raw->'person_info'->>'birthday','')
		    ) AS bday_str
		  FROM core.clients_versions c
		  WHERE c.is_current = true
		)
		UPDATE core.clients_versions AS c
		SET needs_second_part = true
		FROM birthdays b
		WHERE c.client_id = b.client_id
		  AND c.is_current = true
		  AND c.needs_second_part = false
		  AND b.bday_str IS NOT NULL
		  AND (
		    NOW() >= (to_date(b.bday_str, 'DD.MM.YYYY') + INTERVAL '20 years')
		    OR
		    NOW() >= (to_date(b.bday_str, 'DD.MM.YYYY') + INTERVAL '45 years')
		  )
	`)
	return res.RowsAffected, res.Error
}
