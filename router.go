package h2o

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
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
func (p *Router) Run(port int, grace bool, options cors.Options) error {
	addr := fmt.Sprintf(":%d", port)
	log.Infof(
		"application starting on http://localhost:%d",
		port,
	)
	// --------------
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
	// ----------------
	hnd := cors.New(options).Handler(rt)
	// ----------------
	if grace {
		srv := &http.Server{Addr: addr, Handler: hnd}
		go func() {
			// service connections
			if err := srv.ListenAndServe(); err != nil {
				log.Error(err)
			}
		}()

		// Wait for interrupt signal to gracefully shutdown the server with
		// a timeout of 5 seconds.
		quit := make(chan os.Signal)
		signal.Notify(quit, os.Interrupt)
		<-quit
		log.Warningf("shutdown server ...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			return err
		}
		log.Info("server exist")
		return nil
	}
	// ----------------
	return http.ListenAndServe(addr, hnd)
}
