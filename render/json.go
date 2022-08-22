package render

import (
	"encoding/json"
	"net/http"
)

type JSON struct {
	Data any
}

func (j *JSON) Render(w http.ResponseWriter, statusCode int) error {
	j.WriteContentType(w)
	w.WriteHeader(statusCode)
	jsonData, err := json.Marshal(j.Data) // 数据编码转byte字符
	if err != nil {
		return err
	}
	_, err = w.Write(jsonData) // 调用Write后会自动调用一次WriteHeader(),而WriteHeader()不能重复调用,否则会报错
	return err
}

func (j *JSON) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, "application/json; charset=utf-8")
}
