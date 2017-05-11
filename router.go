package h2o

import (
	"net/http"
	"reflect"
	"runtime"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

// HandlerFunc an adapter to allow the use of ordinary functions as HTTP handler
type HandlerFunc func(*Context) error

// New new blank Engine instance without any middleware attached.
func New() *Router {
	return &Router{
		handlers: make([]HandlerFunc, 0),
		routes:   make([]*route, 0),
	}
}

// Router associated with a prefix and an array of handlers
type Router struct {
	path     string
	handlers []HandlerFunc
	routes   []*route
}

// Group creates a new router group
func (p *Router) Group(group func(*Router), path string, handlers ...HandlerFunc) {
	rt := &Router{
		path:     p.path + path,
		handlers: append(p.handlers, handlers...),
		routes:   make([]*route, 0),
	}
	group(rt)
	p.routes = append(p.routes, rt.routes...)
}

// GET http get
func (p *Router) GET(path string, handlers ...HandlerFunc) {
	p.Handle([]string{http.MethodGet}, path, handlers...)
}

// POST http post
func (p *Router) POST(path string, handlers ...HandlerFunc) {
	p.Handle([]string{http.MethodPost}, path, handlers...)
}

// PUT http put
func (p *Router) PUT(path string, handlers ...HandlerFunc) {
	p.Handle([]string{http.MethodPut}, path, handlers...)
}

// PATCH http patch
func (p *Router) PATCH(path string, handlers ...HandlerFunc) {
	p.Handle([]string{http.MethodPatch}, path, handlers...)
}

// DELETE http delete
func (p *Router) DELETE(path string, handlers ...HandlerFunc) {
	p.Handle([]string{http.MethodDelete}, path, handlers...)
}

// Handle registers a new request handle and middleware with the given path and method.
func (p *Router) Handle(methods []string, path string, handlers ...HandlerFunc) {
	p.routes = append(
		p.routes,
		&route{
			methods:  methods,
			path:     p.path + path,
			handlers: append(p.handlers, handlers...),
		},
	)
}

// Run attaches the router to a http.Server and starts listening and serving HTTP requests.
func (p *Router) Run(port int, grace bool) error {
	rt := mux.NewRouter()
	for _, r := range p.routes {
		rt.HandleFunc(r.path, func(w http.ResponseWriter, r *http.Request) {
			begin := time.Now()
			ctx := Context{
				Request: r,
				Writer:  w,
				vars:    mux.Vars(r),
			}
			log.Infof("%s %s %s %s", r.Proto, r.Method, r.RequestURI, ctx.ClientIP())
			for _, h := range p.handlers {
				log.Debugf("call %s", runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name())
				if e := h(&ctx); e != nil {
					log.Error(e)
					s := http.StatusInternalServerError
					if he, ok := e.(*HTTPError); ok {
						s = he.Status
					}
					http.Error(w, e.Error(), s)
				}
			}
			log.Infof("done %s", time.Now().Sub(begin))
		}).Methods(r.methods...)
	}

	return nil
}
