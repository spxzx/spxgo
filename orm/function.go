package orm

import (
	"fmt"
	"strings"
)

func (s *SpxSession) Aggregate(funcName, field string) (int64, error) {
	var fieldSb strings.Builder
	if field == "" {
		field = "*"
	}
	fieldSb.WriteString(funcName + "(" + field + ")")
	query := fmt.Sprintf("select %s from %s", fieldSb.String(), s.tableName)
	var sb strings.Builder
	sb.WriteString(query)
	sb.WriteString(s.conditionParam.String())
	query = sb.String()
	s.db.logger.Info(query)
	stmt, err := s.db.db.Prepare(query)
	if err != nil {
		return -1, err
	}
	row := stmt.QueryRow(s.conditionValues...)
	if row.Err() != nil {
		return -1, row.Err()
	}
	var res int64
	err = row.Scan(&res)
	if err != nil {
		return -1, err
	}
	return res, nil
}

func (s *SpxSession) Average(field string) (int64, error) {
	if field == "" {
		panic("avg must act on a column")
	}
	return s.Aggregate("avg", field)
}

func (s *SpxSession) Count(field string) (int64, error) {
	return s.Aggregate("count", field)
}

func (s *SpxSession) Max(field string) (int64, error) {
	if field == "" {
		panic("max must act on a column")
	}
	return s.Aggregate("max", field)
}

func (s *SpxSession) Min(field string) (int64, error) {
	if field == "" {
		panic("min must act on a column")
	}
	return s.Aggregate("min", field)
}

func (s *SpxSession) Sum(field string) (int64, error) {
	if field == "" {
		panic("sum must act on a column")
	}
	return s.Aggregate("sum", field)
}
