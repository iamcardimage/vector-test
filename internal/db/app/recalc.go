package app

import (
	"fmt"

	"gorm.io/gorm"
)

func RecalcNeedsSecondPart(gdb *gorm.DB) (int64, error) {
	query := `
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
	`

	result := gdb.Exec(query)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to recalc needs_second_part: %w", result.Error)
	}

	return result.RowsAffected, nil
}

func RecalcPassportExpiry(gdb *gorm.DB) (int64, error) {
	query := `
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
	`

	result := gdb.Exec(query)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to recalc passport expiry: %w", result.Error)
	}

	return result.RowsAffected, nil
}

func RecalcAll(gdb *gorm.DB) error {
	n1, err := RecalcNeedsSecondPart(gdb)
	if err != nil {
		return fmt.Errorf("RecalcNeedsSecondPart failed: %w", err)
	}

	n2, err := RecalcPassportExpiry(gdb)
	if err != nil {
		return fmt.Errorf("RecalcPassportExpiry failed: %w", err)
	}

	if n1 > 0 || n2 > 0 {
		fmt.Printf("Recalc completed: needs_second_part=%d, passport_expiry=%d\n", n1, n2)
	}

	return nil
}
