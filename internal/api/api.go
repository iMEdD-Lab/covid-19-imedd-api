package api

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"covid19-greece-api/pkg/env"
)

type Api struct {
	Router *chi.Mux
}

func NewApi() *Api {
	api := Api{}
	api.instantiateRouter()

	return &api
}

func (a *Api) Serve() error {
	listenAddr := env.EnvOrDefault("PORT", ":12122")
	log.Printf("Covid 19 GR API started. Listening on %s\n", listenAddr)

	return http.ListenAndServe(listenAddr, a.Router)
}

func (a *Api) instantiateRouter() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// health status
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello friend."))
		w.WriteHeader(http.StatusOK)
	})

	a.Router = r
}
