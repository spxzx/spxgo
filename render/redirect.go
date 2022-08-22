package render

import (
	"errors"
	"fmt"
	"net/http"
)

type Redirect struct {
	Request  *http.Request
	Location string
}

func (r *Redirect) Render(w http.ResponseWriter, statusCode int) error {
	if (statusCode < http.StatusMultipleChoices || statusCode > http.StatusPermanentRedirect) && statusCode != http.StatusCreated {
		return errors.New(fmt.Sprintf("Cannot redirect with status code %d", statusCode))
	}
	http.Redirect(w, r.Request, r.Location, statusCode)
	return nil
}

func (r *Redirect) WriteContentType(w http.ResponseWriter) {
}
