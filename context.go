package h2o

import (
	"context"
	"net"
	"net/http"
	"strings"

	validator "gopkg.in/go-playground/validator.v9"

	"github.com/go-playground/form"
	"github.com/unrolled/render"
)

// K key type
type K string

// H hash
type H map[string]interface{}

// Context context
type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request

	vars map[string]string
	val  *validator.Validate
	dec  *form.Decoder
	rdr  *render.Render
}

// Set set
func (p *Context) Set(k string, v interface{}) {
	p.Request = p.Request.WithContext(
		context.WithValue(p.Request.Context(), K(k), v),
	)
}

// Get get
func (p *Context) Get(k string) interface{} {
	return p.Request.Context().Value(K(k))
}

// Redirect redirect
func (p *Context) Redirect(code int, url string) {
	http.Redirect(p.Writer, p.Request, url, code)
}

// SetHeader set header
func (p *Context) SetHeader(k, v string) {
	p.Request.Header.Set(k, v)
}

// GetHeader get header
func (p *Context) GetHeader(k string) string {
	return p.Request.Header.Get(k)
}

// Param the value of the URL param.
func (p *Context) Param(k string) string {
	return p.vars[k]
}

// ClientIP client ip
func (p *Context) ClientIP() string {
	// -------------
	if ip := strings.TrimSpace(p.GetHeader("X-Real-Ip")); ip != "" {
		return ip
	}
	// -------------
	ip := p.GetHeader("X-Forwarded-For")
	if idx := strings.IndexByte(ip, ','); idx >= 0 {
		ip = ip[0:idx]
	}
	ip = strings.TrimSpace(ip)
	if ip != "" {
		return ip
	}
	// -------------
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(p.Request.RemoteAddr)); err == nil {
		return ip
	}
	// -----------
	return ""
}

// JSON write json
func (p *Context) JSON(s int, v interface{}) error {
	return p.rdr.JSON(p.Writer, s, v)
}

// XML write xml
func (p *Context) XML(s int, v interface{}) error {
	return p.rdr.XML(p.Writer, s, v)
}

// Bind binds the passed struct pointer
func (p *Context) Bind(v interface{}) error {
	e := p.Request.ParseForm()
	if e == nil {
		e = p.dec.Decode(v, p.Request.Form)
	}
	if e == nil {
		e = p.val.Struct(v)
	}
	return e
}
