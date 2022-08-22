package orm

import (
	"errors"
	"fmt"
	"strings"
)

func (s *SpxSession) Update(data ...any) (int64, int64, error) {
	// support Update("field", any) / Update(data)
	if len(data) == 0 || len(data) > 2 {
		return -1, -1, errors.New("param not valid")
	}
	single := true
	if len(data) == 2 {
		single = false
	}
	if !single {
		s.updateParam.WriteString(data[0].(string))
		s.updateParam.WriteString("=?")
		s.values = append(s.values, data[1])
	} else {
		s.filling(Update, data[0])
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("update %s set %s", s.tableName, s.updateParam.String()))
	sb.WriteString(s.conditionParam.String())
	query := sb.String()
	s.db.logger.Info(query)
	return s.execute(Update, query)
}
