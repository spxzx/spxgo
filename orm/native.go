package orm

import (
	"reflect"
	"strings"
)

// 原生 SQL 支持

func (s *SpxSession) Exec(sql string, values ...any) (int64, error) {
	s.db.logger.Info(sql)
	stmt, err := s.db.db.Prepare(sql)
	if err != nil {
		return -1, err
	}
	r, err := stmt.Exec(values...)
	if err != nil {
		return -1, err
	}
	if strings.Contains(strings.ToLower(sql), "insert") {
		return r.LastInsertId()
	}
	return r.RowsAffected()
}

func (s *SpxSession) QueryRow(sql string, data any, queryValues ...any) error {
	s.db.logger.Info(sql)
	stmt, err := s.db.db.Prepare(sql)
	if err != nil {
		return err
	}
	rows, err := stmt.Query(queryValues...)
	if err != nil {
		return err
	}
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	if rows.Next() {
		_, err = s.nextFill(SelectOne, data, reflect.TypeOf(data), rows, cols)
		if err != nil {
			return err
		}
	}
	return nil
}
