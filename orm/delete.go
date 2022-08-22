package orm

import (
	"fmt"
	"strings"
)

func (s *SpxSession) Delete() (int64, error) {
	query := fmt.Sprintf("delete from %s", s.tableName)
	var sb strings.Builder
	sb.WriteString(query)
	sb.WriteString(s.conditionParam.String())
	query = sb.String()
	s.db.logger.Info(query)
	_, r, err := s.execute(Delete, query)
	return r, err
}
