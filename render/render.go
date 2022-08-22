package render

import "net/http"

type Render interface {
	Render(w http.ResponseWriter, statusCode int) error
	WriteContentType(w http.ResponseWriter)
}

func writeContentType(w http.ResponseWriter, value string) {
	// type Header map[string][]string
	w.Header().Set("Content-Type", value)
}
