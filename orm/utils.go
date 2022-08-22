package orm

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

const (
	Insert = iota
	InsertBatch
	Update
	Delete
	Select
	SelectOne
)

func IsAutoId(id any) bool {
	t := reflect.TypeOf(id)
	switch t.Kind() {
	case reflect.Int64:
		if (id.(int64)) <= 0 {
			return true
		}
	case reflect.Int32:
		if (id.(int32)) <= 0 {
			return true
		}
	case reflect.Int:
		if (id.(int)) <= 0 {
			return true
		}
	default:
		return false
	}
	return false
}

func Name(name string) string {
	names := name[:]
	lastIndex := 0
	var sb strings.Builder
	for i, v := range names {
		if v >= 65 && v <= 90 {
			if i == 0 {
				continue
			}
			sb.WriteString(name[lastIndex:i])
			sb.WriteString("_")
			lastIndex = i
		}
	}
	sb.WriteString(name[lastIndex:])
	return sb.String()
}

func (s *SpxSession) execute(option int, query string) (int64, int64, error) {
	var stmt *sql.Stmt
	var err error
	if s.beginTx {
		stmt, err = s.tx.Prepare(query)
	} else {
		stmt, err = s.db.db.Prepare(query)
	}
	if err != nil {
		return -1, -1, err
	}
	var r sql.Result
	switch option {
	case Insert:
		r, err = stmt.Exec(s.values...)
	case Update:
		s.values = append(s.values, s.conditionValues...)
		r, err = stmt.Exec(s.values...)
	case Delete:
		r, err = stmt.Exec(s.conditionValues...)
	}
	if err != nil {
		return -1, -1, err
	}
	id, err := r.LastInsertId() // 这个只有在字段为自增时才能有效返回对应id
	if err != nil {
		return -1, -1, err
	}
	affected, err := r.RowsAffected()
	if err != nil {
		return -1, -1, err
	}
	return id, affected, nil
}

func (s *SpxSession) filling(option int, data any) {
	// reflect
	t := reflect.TypeOf(data)
	v := reflect.ValueOf(data)
	if t.Kind() != reflect.Pointer {
		panic(errors.New("data must be pointer"))
	}
	tVal := t.Elem()
	vVal := v.Elem()
	for i := 0; i < tVal.NumField(); i++ {
		sqlTag := tVal.Field(i).Tag.Get("spxorm")
		if sqlTag == "" {
			sqlTag = strings.ToLower(Name(tVal.Field(i).Name))
		} else {
			if strings.Contains(sqlTag, "auto_increment") && option != Update {
				continue
			}
			if strings.Contains(sqlTag, ",") && option != InsertBatch {
				sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
			}
		}
		id := vVal.Field(i).Interface()
		if strings.ToLower(sqlTag) == "id" && IsAutoId(id) {
			continue
		}
		switch option {
		case Insert:
			s.fieldName = append(s.fieldName, sqlTag)
			s.placeHolder = append(s.placeHolder, "?")
		case InsertBatch:

		case Update:
			if s.updateParam.String() != "" {
				s.updateParam.WriteString(",")
			}
			s.updateParam.WriteString(sqlTag)
			s.updateParam.WriteString("=?")
		}
		s.values = append(s.values, vVal.Field(i).Interface())
	}
}

func (s *SpxSession) batchFilling(data []any) {
	s.values = make([]any, 0)
	for _, value := range data {
		s.filling(InsertBatch, value)
	}
}

func (s *SpxSession) selectComponent(data any, fields ...string) (*sql.Rows, []string, error) {
	t := reflect.TypeOf(data)
	if t.Kind() != reflect.Pointer {
		panic(errors.New("data type must be pointer"))
	}
	fieldStr := "*"
	if len(fields) > 0 {
		fieldStr = strings.Join(fields, ",")
	}
	query := fmt.Sprintf("select %s from %s",
		fieldStr, s.tableName)
	var sb strings.Builder
	sb.WriteString(query)
	sb.WriteString(s.conditionParam.String())
	query = sb.String()
	s.db.logger.Info(query)
	stmt, err := s.db.db.Prepare(query)
	if err != nil {
		return nil, nil, err
	}
	rows, err := stmt.Query(s.conditionValues...)
	if err != nil {
		return nil, nil, err
	}
	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}
	return rows, cols, nil
}

// 类型转换
func (s *SpxSession) convertType(i, j int, values []any, tVal reflect.Type) reflect.Value {
	target := values[j]
	targetValue := reflect.ValueOf(target)
	fieldType := tVal.Field(i).Type
	res := reflect.ValueOf(targetValue.Interface()).Convert(fieldType)
	return res
}

func (s *SpxSession) nextFill(option int, data any, t reflect.Type, rows *sql.Rows, cols []string) (result any, err error) {
	if option == Select {
		// 对于多行查询
		// 由于传进来的是一个指针地址 会导致 result 里面的值都一样
		// 所以每次查询的时候要让 data 重新换一个地址
		data = reflect.New(t.Elem()).Interface()
	}
	values := make([]any, len(cols))
	fieldScan := make([]any, len(cols))
	for i := range fieldScan {
		fieldScan[i] = &values[i]
	}
	err = rows.Scan(fieldScan...)
	if err != nil {
		return nil, err
	}
	tVal := t.Elem()
	vVal := reflect.ValueOf(data).Elem()
	for i := 0; i < tVal.NumField(); i++ {
		sqlTag := tVal.Field(i).Tag.Get("spxorm")
		if sqlTag == "" {
			sqlTag = strings.ToLower(Name(tVal.Field(i).Name))
		} else {
			if strings.Contains(sqlTag, ",") {
				sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
			}
		}
		for j, colName := range cols {
			if sqlTag == colName {
				if vVal.Field(i).CanSet() {
					vVal.Field(i).Set(s.convertType(i, j, values, tVal))
				}
			}
		}
	}
	return data, nil
}
