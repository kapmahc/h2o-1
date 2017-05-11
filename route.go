package h2o

type route struct {
	path     string
	methods  []string
	handlers []HandlerFunc
}
