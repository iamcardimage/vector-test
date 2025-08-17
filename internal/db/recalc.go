package db

import "gorm.io/gorm"

// RecalcNeedsSecondPart помечает текущие версии клиентов, для которых:
// - текущая 2-я часть имеет due_at <= now, ИЛИ
// - текущая 2-я часть относится к старой версии клиента (sp.client_version <> c.version)
// Возвращает кол-во обновлённых строк.
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
