package orm

import "strings"

func (s *SpxSession) writeConditionParam(field string, value any) *SpxSession {
	s.conditionParam.WriteString(field + " = ?")
	s.conditionValues = append(s.conditionValues, value)
	return s
}

func (s *SpxSession) Where(field string, value any) *SpxSession {
	s.conditionParam.WriteString(" where ")
	return s.writeConditionParam(field, value)
}

func (s *SpxSession) And(field string, value any) *SpxSession {
	s.conditionParam.WriteString(" and ")
	return s.writeConditionParam(field, value)
}

func (s *SpxSession) Or(field string, value any) *SpxSession {
	s.conditionParam.WriteString(" or ")
	return s.writeConditionParam(field, value)
}

func (s *SpxSession) Like(field string, value any) *SpxSession {
	if s.conditionParam.String() == "" {
		s.conditionParam.WriteString(" where ")
	}
	s.conditionParam.WriteString(field + " like ?")
	s.conditionValues = append(s.conditionValues, "%"+value.(string)+"%")
	return s
}

func (s *SpxSession) LikeRight(field string, value any) *SpxSession {
	if s.conditionParam.String() == "" {
		s.conditionParam.WriteString(" where ")
	}
	s.conditionParam.WriteString(field + " like ?")
	s.conditionValues = append(s.conditionValues, value.(string)+"%")
	return s
}

func (s *SpxSession) LikeLeft(field string, value any) *SpxSession {
	if s.conditionParam.String() == "" {
		s.conditionParam.WriteString(" where ")
	}
	s.conditionParam.WriteString(field + " like ?")
	s.conditionValues = append(s.conditionValues, "%"+value.(string))
	return s
}

func (s *SpxSession) Group(field ...string) *SpxSession {
	s.conditionParam.WriteString(" group by ")
	s.conditionParam.WriteString(strings.Join(field, ","))
	return s
}

func (s *SpxSession) OrderDesc(field ...string) *SpxSession {
	s.conditionParam.WriteString(" order by ")
	s.conditionParam.WriteString(strings.Join(field, ","))
	s.conditionParam.WriteString(" desc")
	return s
}

func (s *SpxSession) OrderAsc(field ...string) *SpxSession {
	s.conditionParam.WriteString(" order by ")
	s.conditionParam.WriteString(strings.Join(field, ","))
	s.conditionParam.WriteString(" asc")
	return s
}

// Order 参数: desc 必须在 asc前面
func (s *SpxSession) Order(field ...string) *SpxSession {
	if len(field)%2 != 0 {
		panic("field num not true")
	}
	s.conditionParam.WriteString(" order by ")
	for i, v := range field {
		s.conditionParam.WriteString(" " + v)
		if i%2 != 0 && i < len(field)-1 {
			s.conditionParam.WriteString(",")
		}
	}
	return s
}
