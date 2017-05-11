package main

import (
	"log"

	"github.com/kapmahc/h2o"
	"github.com/rs/cors"
	"github.com/unrolled/render"
)

func h1(c *h2o.Context) error {
	c.Writer.Write([]byte("h1 \n"))
	return nil
}
func h2(c *h2o.Context) error {
	c.Writer.Write([]byte("h2 \n"))
	return nil
}
func h3(c *h2o.Context) error {
	c.Writer.Write([]byte("h3 \n"))
	return nil
}

func main() {
	rt := h2o.New()
	rt.GET("/demo", h1, h2, h3)
	if err := rt.Run(8080, true, cors.Options{}, render.Options{}); err != nil {
		log.Fatal(err)
	}
}
