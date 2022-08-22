package binding

import (
	"encoding/xml"
	"errors"
	"io"
	"net/http"
)

type xmlBinding struct {
}

func (xmlBinding) Name() string {
	return "xml"
}

func (b xmlBinding) Bind(r *http.Request, obj any) error {
	if r.Body == nil {
		return errors.New("invalid request")
	}
	return decodeXML(r.Body, obj)
}

func decodeXML(r io.Reader, obj any) error {
	decoder := xml.NewDecoder(r)
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return validate(obj)
}
