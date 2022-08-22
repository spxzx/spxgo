package orm

import "reflect"

func (s *SpxSession) SelectOne(data any, fields ...string) error {
	rows, cols, err := s.selectComponent(data, fields...)
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

func (s *SpxSession) Select(data any, fields ...string) ([]any, error) {
	rows, cols, err := s.selectComponent(data, fields...)
	if err != nil {
		return nil, err
	}
	result := make([]any, 0)
	for rows.Next() {
		res, err := s.nextFill(Select, data, reflect.TypeOf(data), rows, cols)
		if err != nil {
			return nil, err
		}
		result = append(result, res)
	}
	return result, nil
}
