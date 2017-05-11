package h2o

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/go-playground/form"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/unrolled/render"
	validator "gopkg.in/go-playground/validator.v9"
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

// Use use middlewares
func (p *Router) Use(handlers ...HandlerFunc) {
	p.handlers = append(p.handlers, handlers...)
}

// Crud crud
func (p *Router) Crud(path string, list []HandlerFunc, create []HandlerFunc, read []HandlerFunc, update []HandlerFunc, delete []HandlerFunc) {
	p.GET(path, list...)
	p.POST(path, create...)
	child := path + "/{id}"
	p.GET(child, read...)
	p.POST(child, update...)
	p.DELETE(child, delete...)
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
func (p *Router) Run(port int, grace bool, cro cors.Options, rdo render.Options) error {
	addr := fmt.Sprintf(":%d", port)
	log.Infof(
		"application starting on http://localhost:%d",
		port,
	)
	// --------------
	rt := mux.NewRouter()
	va := validator.New()
	de := form.NewDecoder()
	rd := render.New(rdo)
	for _, r := range p.routes {
		rt.HandleFunc(
			r.path,
			r.handle(func(w http.ResponseWriter, r *http.Request) *Context {
				return &Context{
					Request: r,
					Writer:  w,
					vars:    mux.Vars(r),
					dec:     de,
					val:     va,
					rdr:     rd,
				}
			}),
		).Methods(r.methods...)
	}
	// ----------------
	hnd := cors.New(cro).Handler(rt)
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

// WalkFunc walk func
type WalkFunc func(methods []string, path string, handlers ...HandlerFunc) error

// Walk walk routes
func (p *Router) Walk(f WalkFunc) error {
	for _, r := range p.routes {
		if e := f(r.methods, r.path, r.handlers...); e != nil {
			return e
		}
	}
	return nil
}
