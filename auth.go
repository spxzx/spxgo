package spxgo

import (
	"encoding/base64"
	"net/http"
)

type Accounts struct {
	UnAuthHandler func(c *Context)
	Users         map[string]string
}

func (a *Accounts) BasicAuth(next HandlerFunc) HandlerFunc {
	return func(c *Context) {
		username, password, ok := c.R.BasicAuth()
		if !ok {
			a.unAuthHandler(c)
			return
		}
		pwd, exist := a.Users[username]
		if !exist || pwd != password {
			a.unAuthHandler(c)
			return
		}
		c.Set("user", username)
		next(c)
	}
}

func (a *Accounts) unAuthHandler(c *Context) {
	if a.UnAuthHandler != nil {
		a.UnAuthHandler(c)
	} else {
		if err := c.String(http.StatusUnauthorized, "user is unauthorized"); err != nil {
			return
		}
	}
}

func BasicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
