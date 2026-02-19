package storage

import "database/sql"

func (r *Repository) SQLDB() *sql.DB {
	if r == nil {
		return nil
	}
	return r.db
}
