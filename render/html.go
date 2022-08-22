package render

import (
	"gitbuh.com/spxzx/spxgo/internal/bytesconv"
	"html/template"
	"net/http"
)

type HTML struct {
	Name       string
	Data       any
	Template   *template.Template
	IsTemplate bool
}

func (h *HTML) Render(w http.ResponseWriter, statusCode int) error {
	h.WriteContentType(w)
	w.WriteHeader(statusCode)
	if h.IsTemplate {
		return h.Template.ExecuteTemplate(w, h.Name, h.Data)
	}
	_, err := w.Write(bytesconv.StringToBytes(h.Data.(string)))
	return err
}

func (h *HTML) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, "text/html; charset=utf-8")
}
