package h2o

import (
	"net/http"
	"reflect"
	"runtime"
	"time"

	log "github.com/Sirupsen/logrus"
)

type route struct {
	path     string
	method   string
	handlers []HandlerFunc
}

func (p *route) call(c *Context) error {
	for _, h := range p.handlers {
		log.Debugf("call %s", runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name())
		if e := h(c); e != nil {
			return e
		}
	}
	return nil
}

func (p *route) handle(f func(http.ResponseWriter, *http.Request) *Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		begin := time.Now()
		ctx := f(w, r)
		log.Infof("%s %s %s %s", r.Proto, r.Method, r.RequestURI, ctx.ClientIP())
		if e := p.call(ctx); e != nil {
			log.Error(e)
			s := http.StatusInternalServerError
			if he, ok := e.(*HTTPError); ok {
				s = he.Status
			}
			http.Error(w, e.Error(), s)
		}
		log.Infof("done %s", time.Now().Sub(begin))
	}
}
