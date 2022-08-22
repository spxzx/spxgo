package orm

import (
	"errors"
	"fmt"
	"strings"
)

func (s *SpxSession) Insert(data any) (int64, int64, error) {
	s.filling(Insert, data)
	query := fmt.Sprintf("insert into %s (%s) values (%s)",
		s.tableName, strings.Join(s.fieldName, ","), strings.Join(s.placeHolder, ","))
	s.db.logger.Info(query)
	return s.execute(Insert, query)
}

func (s *SpxSession) InsertBatch(data []any) (int64, int64, error) {
	if len(data) == 0 {
		return -1, -1, errors.New("no data insert")
	}
	s.filling(Insert, data[0])
	query := fmt.Sprintf("insert into %s (%s) values ",
		s.tableName, strings.Join(s.fieldName, ","))
	var sb strings.Builder
	sb.WriteString(query)
	for i := range data {
		sb.WriteString("(")
		sb.WriteString(strings.Join(s.placeHolder, ","))
		sb.WriteString(")")
		if i < len(data)-1 {
			sb.WriteString(",")
		}
	}
	s.batchFilling(data)
	query = sb.String()
	s.db.logger.Info(query)
	return s.execute(Insert, query)
}
