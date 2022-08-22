package token

import (
	"errors"
	"gitbuh.com/spxzx/spxgo"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"time"
)

const JWTToken = "spxgo_token"

type JwtHandler struct {
	Algorithm      string
	TimeOut        time.Duration
	TimeFunc       func() time.Time
	RefreshTimeOut time.Duration
	RefreshKey     string
	Key            []byte
	PrivateKey     string // 私钥
	SendCookie     bool
	Authenticator  func(c *spxgo.Context) (map[string]any, error)
	CookieName     string
	CookieMaxAge   int
	CookieDomain   string
	SecureCookie   bool
	CookieHTTPOnly bool
	Header         string
	AuthHandler    func(c *spxgo.Context, err error)
}

type JwtResponse struct {
	Token        string
	RefreshToken string
}

func (j *JwtHandler) LoginHandler(c *spxgo.Context) (*JwtResponse, error) {
	data, err := j.Authenticator(c)
	if err != nil {
		return nil, err
	}
	if j.Algorithm == "" {
		j.Algorithm = "HS256"
	}
	// A 部分
	signingMethod := jwt.GetSigningMethod(j.Algorithm)
	token := jwt.New(signingMethod)
	// B 部分
	claims := token.Claims.(jwt.MapClaims)
	if data != nil {
		for k, v := range data {
			claims[k] = v
		}
	}
	if j.TimeFunc == nil {
		j.TimeFunc = func() time.Time {
			return time.Now()
		}
	}
	expire := j.TimeFunc().Add(j.TimeOut)
	// 过期时间
	claims["exp"] = expire.Unix()
	// 发布时间
	claims["iat"] = j.TimeFunc().Unix()
	// C 部分
	var tokenString string
	var tokenErr error
	if j.usingPublicKeyAlgo() {
		tokenString, tokenErr = token.SignedString(j.PrivateKey)
	} else {
		tokenString, tokenErr = token.SignedString(j.Key)
	}
	if tokenErr != nil {
		return nil, tokenErr
	}
	refreshToken, err := j.refreshToken(token)
	if err != nil {
		return nil, err
	}
	if j.SendCookie {
		if j.CookieName == "" {
			j.CookieName = JWTToken
		}
		if j.CookieMaxAge == 0 {
			j.CookieMaxAge = int(expire.Unix() - j.TimeFunc().Unix())
		}
		c.SetCookie(j.CookieName, tokenString, j.CookieMaxAge, "/",
			j.CookieDomain, j.SecureCookie, j.CookieHTTPOnly)
	}
	return &JwtResponse{
		Token:        tokenString,
		RefreshToken: refreshToken,
	}, nil
}

func (j *JwtHandler) usingPublicKeyAlgo() bool {
	switch j.Algorithm {
	case "RS256", "RS512", "RS384":
		return true
	}
	return false
}

func (j *JwtHandler) refreshToken(token *jwt.Token) (string, error) {
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = j.TimeFunc().Add(j.RefreshTimeOut).Unix()
	var tokenString string
	var tokenErr error
	if j.usingPublicKeyAlgo() {
		tokenString, tokenErr = token.SignedString(j.PrivateKey)
	} else {
		tokenString, tokenErr = token.SignedString(j.Key)
	}
	if tokenErr != nil {
		return "", tokenErr
	}
	return tokenString, nil
}

func (j *JwtHandler) LogoutHandler(c *spxgo.Context) {
	if j.SecureCookie {
		if j.CookieName == "" {
			j.CookieName = JWTToken
		}
		c.SetCookie(j.CookieName, "", -1, "/",
			j.CookieDomain, j.SecureCookie, j.CookieHTTPOnly)
	}
}

func (j *JwtHandler) tokenErrorHandler(c *spxgo.Context, err error) {
	if j.AuthHandler == nil {
		c.W.WriteHeader(http.StatusUnauthorized)
	} else {
		j.AuthHandler(c, err)
	}
}

func (j *JwtHandler) RefreshHandler(c *spxgo.Context) (*JwtResponse, error) {
	rToken, ok := c.Get(j.RefreshKey)
	if !ok {
		return nil, errors.New("refresh token is null")
	}
	if j.Algorithm == "" {
		j.Algorithm = "HS256"
	}
	// 解析token
	data, err := jwt.Parse(rToken.(string), func(token *jwt.Token) (interface{}, error) {
		if j.usingPublicKeyAlgo() {
			return []byte(j.PrivateKey), nil
		} else {
			return j.Key, nil
		}
	})
	if err != nil {
		j.tokenErrorHandler(c, err)
		return nil, err
	}
	// B 部分
	claims := data.Claims.(jwt.MapClaims)
	if j.TimeFunc == nil {
		j.TimeFunc = func() time.Time {
			return time.Now()
		}
	}
	expire := j.TimeFunc().Add(j.TimeOut)
	// 过期时间
	claims["exp"] = expire.Unix()
	// 发布时间
	claims["iat"] = j.TimeFunc().Unix()
	// C 部分
	var tokenString string
	var tokenErr error
	if j.usingPublicKeyAlgo() {
		tokenString, tokenErr = data.SignedString(j.PrivateKey)
	} else {
		tokenString, tokenErr = data.SignedString(j.Key)
	}
	if tokenErr != nil {
		return nil, tokenErr
	}
	refreshToken, err := j.refreshToken(data)
	if err != nil {
		return nil, err
	}
	if j.SendCookie {
		if j.CookieName == "" {
			j.CookieName = JWTToken
		}
		if j.CookieMaxAge == 0 {
			j.CookieMaxAge = int(expire.Unix() - j.TimeFunc().Unix())
		}
		c.SetCookie(j.CookieName, tokenString, j.CookieMaxAge, "/",
			j.CookieDomain, j.SecureCookie, j.CookieHTTPOnly)
	}
	return &JwtResponse{
		Token:        tokenString,
		RefreshToken: refreshToken,
	}, nil
}

// JWT 登陆中间件

func (j *JwtHandler) AuthInterceptor(next spxgo.HandlerFunc) spxgo.HandlerFunc {
	return func(c *spxgo.Context) {
		if j.Header == "" {
			j.Header = "Authorization"
		}
		token := c.R.Header.Get(j.Header)
		if token == "" {
			if j.SendCookie {
				cookie, err := c.R.Cookie(j.CookieName)
				if err != nil {
					j.tokenErrorHandler(c, err)
					return
				}
				token = cookie.String()
			}
		}
		if token == "" {
			j.tokenErrorHandler(c, nil)
			return
		}
		data, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			if j.usingPublicKeyAlgo() {
				return []byte(j.PrivateKey), nil
			} else {
				return j.Key, nil
			}
		})
		if err != nil {
			j.tokenErrorHandler(c, nil)
			return
		}
		claims := data.Claims.(jwt.MapClaims)
		c.Set("jwt_claims", claims)
		next(c)
	}
}
