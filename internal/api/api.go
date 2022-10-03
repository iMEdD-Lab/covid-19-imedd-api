package api

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"covid19-greece-api/internal/data"
	"covid19-greece-api/pkg/env"
	"covid19-greece-api/pkg/vartypes"
)

type Api struct {
	router *chi.Mux
	repo   data.Repo
}

// todo add rate limiter
// todo add cache
// todo add authentication

func NewApi(repo data.Repo) *Api {
	api := Api{
		repo: repo,
	}
	api.instantiateRouter()

	return &api
}

func (a *Api) Serve() error {
	listenAddr := env.EnvOrDefault("PORT", ":12122")
	log.Printf("Covid 19 GR API started. Listening on %s\n", listenAddr)

	return http.ListenAndServe(listenAddr, a.router)
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

	// jwt protected routes
	r.Group(func(r chi.Router) {
		r.Use(a.authenticationMw)

		r.Get("/check-auth", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("hello friend, you are authenticated!"))
			w.WriteHeader(http.StatusOK)
		})

		r.Get("/geo-info", func(w http.ResponseWriter, r *http.Request) {
			info, err := a.repo.GetGeoInfo(r.Context())
			if err != nil {
				renderJson(w, r, http.StatusInternalServerError, nil)
				return
			}
			renderJson(w, r, http.StatusOK, info)
		})

		r.Get("/cases", func(w http.ResponseWriter, r *http.Request) {
			filter := casesFilter(r.URL.Query())
			cases, err := a.repo.GetCases(r.Context(), filter)
			if err != nil {
				renderJson(w, r, http.StatusInternalServerError, nil)
				return
			}
			renderJson(w, r, http.StatusOK, cases)
		})
	})

	a.router = r
}

func casesFilter(values url.Values) data.CasesFilter {
	f := data.CasesFilter{}
	for k, v := range values {
		switch k {
		case "geo_id":
			f.GeoId = vartypes.StringToInt(v[0])
		case "end_date":
			f.EndDate, _ = time.Parse("2006-01-02", v[0])
		case "start_date":
			f.StartDate, _ = time.Parse("2006-01-02", v[0])
		}
	}

	return f
}

func (a *Api) authenticationMw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := ExtractTokenFromRequest(r)
		if err := a.Authenticate(tokenStr); err != nil {
			log.Println(err)
			renderJson(w, r, http.StatusUnauthorized, ErrorResp{"unauthorized"})
			return
		}
		next.ServeHTTP(w, r.WithContext(r.Context()))
	})
}

func (a *Api) Authenticate(token string) error {
	// todo fill out authentication method
	return nil
}

func ExtractTokenFromRequest(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

func renderJson(w http.ResponseWriter, r *http.Request, statusCode int, content interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(content); err != nil {
		log.Println("failed to marshal ErrorResp:", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

type ErrorResp struct {
	Msg string `json:"message"`
}
