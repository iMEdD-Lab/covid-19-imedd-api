package api

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gitea.com/go-chi/cache"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"

	"covid19-greece-api/internal/data"
	"covid19-greece-api/pkg/env"
	"covid19-greece-api/pkg/vartypes"
)

type Api struct {
	router *chi.Mux
	repo   data.Repo
	cache  cache.Cache
}

func NewApi(repo data.Repo) *Api {
	api := Api{
		repo:  repo,
		cache: cache.NewMemoryCacher(),
	}
	api.initRouter()

	return &api
}

func (a *Api) Serve() error {
	listenAddr := env.EnvOrDefault("PORT", ":8080")
	log.Printf("Covid 19 GR API started. Listening on %s\n", listenAddr)

	return http.ListenAndServe(listenAddr, a.router)
}

func (a *Api) initRouter() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	r.Use(a.cacheMw)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET"},
	}))

	// health status
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello friend."))
		w.WriteHeader(http.StatusOK)
	})

	// jwt protected routes
	r.Group(func(r chi.Router) {
		r.Use(a.authMw)

		r.Get("/check_auth", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("hello friend, you are authenticated!"))
			w.WriteHeader(http.StatusOK)
		})

		r.Get("/geo_info", func(w http.ResponseWriter, r *http.Request) {
			info, err := a.repo.GetGeoInfo(r.Context())
			if err != nil {
				log.Println(err)
				a.respondError(w, r, http.StatusInternalServerError, nil)
				return
			}
			a.respond200(w, r, info, false)
		})

		r.Get("/cases", func(w http.ResponseWriter, r *http.Request) {
			filter := casesFilter(r.URL.Query())
			cases, err := a.repo.GetCases(r.Context(), filter)
			if err != nil {
				log.Println(err)
				a.respondError(w, r, http.StatusInternalServerError, nil)
				return
			}
			a.respond200(w, r, cases, false)
		})

		// helper endpoint
		r.Get("/timeline_fields", func(w http.ResponseWriter, r *http.Request) {
			tlFields := []string{
				"cases",
				"total_reinfections",
				"deaths",
				"deaths_cum",
				"recovered",
				"beds_occupancy",
				"icu_occupancy",
				"intubated",
				"intubated_vac",
				"intubated_unvac",
				"hospital_admissions",
				"hospital_discharges",
				"estimated_new_rtpcr_tests",
				"estimated_new_rapid_tests",
				"estimated_new_total_tests",
			}
			a.respond200(w, r, tlFields, false)
		})

		r.Get("/timeline", func(w http.ResponseWriter, r *http.Request) {
			tlf := getTimelineFilter(r.URL.Query())
			info, err := a.repo.GetFromTimeline(r.Context(), tlf.DatesFilter)
			if err != nil {
				log.Println(err)
				a.respondError(w, r, http.StatusInternalServerError, nil)
				return
			}
			if len(tlf.Fields) > 0 {
				a.respond200(w, r, keepFields(tlf.Fields, info), false)
				return
			}
			a.respond200(w, r, info, false)
		})

		r.Get("/{field}", func(w http.ResponseWriter, r *http.Request) {
			field := chi.URLParam(r, "field")
			info, err := a.repo.GetFromTimeline(r.Context(), datesFilter(r.URL.Query()))
			if err != nil {
				log.Println(err)
				a.respondError(w, r, http.StatusInternalServerError, nil)
				return
			}
			a.respond200(w, r, keepFields([]string{field}, info), false)
		})

	})

	a.router = r
}

func getTimelineFilter(values url.Values) TimelineFilter {
	var fields []string
	if len(values["fields"]) > 0 {
		fields = strings.Split(values["fields"][0], ",")
	}
	return TimelineFilter{
		DatesFilter: datesFilter(values),
		Fields:      fields,
	}
}

type TimelineFilter struct {
	data.DatesFilter
	Fields []string
}

func keepFields(fields []string, fullInfos []data.FullInfo) []map[string]interface{} {
	var res []map[string]interface{}
	for _, fi := range fullInfos {
		r := make(map[string]interface{})
		r["date"] = fi.Date
		for _, f := range fields {
			switch f {
			case "cases":
				r[f] = fi.Cases
			case "total_reinfections":
				r[f] = fi.TotalReinfections
			case "deaths":
				r[f] = fi.Deaths
			case "deaths_cum":
				r[f] = fi.DeathsCum
			case "recovered":
				r[f] = fi.Recovered
			case "beds_occupancy":
				r[f] = fi.BedsOccupancy
			case "icu_occupancy":
				r[f] = fi.IcuOccupancy
			case "intubated":
				r[f] = fi.Intubated
			case "intubated_vac":
				r[f] = fi.IntubatedVac
			case "intubated_unvac":
				r[f] = fi.IntubatedUnvac
			case "hospital_admissions":
				r[f] = fi.HospitalAdmissions
			case "hospital_discharges":
				r[f] = fi.HospitalDischarges
			case "estimated_new_rtpcr_tests":
				r[f] = fi.EstimatedNewRtpcrTests
			case "estimated_new_rapid_tests":
				r[f] = fi.EstimatedNewRapidTests
			case "estimated_new_total_tests":
				r[f] = fi.EstimatedNewTotalTests
			default:
				// do nothing
			}
		}
		res = append(res, r)
	}

	return res
}

func datesFilter(values url.Values) data.DatesFilter {
	f := data.DatesFilter{}
	for k, v := range values {
		switch k {
		case "end_date":
			f.EndDate, _ = time.Parse("2006-01-02", v[0])
		case "start_date":
			f.StartDate, _ = time.Parse("2006-01-02", v[0])
		}
	}

	return f
}

func casesFilter(values url.Values) data.CasesFilter {
	// todo order by?
	f := data.CasesFilter{}
	f.DatesFilter = datesFilter(values)
	for k, v := range values {
		switch k {
		case "geo_id":
			f.GeoId = vartypes.StringToInt(v[0])
		}
	}

	return f
}

func (a *Api) authMw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := ExtractTokenFromRequest(r)
		if err := a.Authenticate(tokenStr); err != nil {
			log.Println(err)
			a.respondError(w, r, http.StatusUnauthorized, ErrorResp{"unauthorized"})
			return
		}
		next.ServeHTTP(w, r.WithContext(r.Context()))
	})
}

func ExtractTokenFromRequest(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

func (a *Api) Authenticate(token string) error {
	// todo fill out authentication method
	return nil
}

func (a *Api) cacheMw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && a.cache.IsExist(r.URL.RequestURI()) {
			content := a.cache.Get(r.URL.RequestURI())
			a.respond200(w, r, content, true)
			return
		}
		next.ServeHTTP(w, r.WithContext(r.Context()))
	})
}

func (a *Api) respondError(w http.ResponseWriter, r *http.Request, statusCode int, content interface{}) {
	w.WriteHeader(statusCode)
	bytes, err := json.Marshal(content)
	if err != nil {
		log.Println("failed to marshal ErrorResp:", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(bytes)
}

func (a *Api) respond200(w http.ResponseWriter, r *http.Request, content interface{}, fromCache bool) {
	if !fromCache {
		a.cache.Put(r.URL.RequestURI(), content, 60*60*24)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	bytes, err := json.Marshal(content)
	if err != nil {
		log.Println("failed to marshal ErrorResp:", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(bytes)))
	w.Write(bytes)
}

type ErrorResp struct {
	Msg string `json:"message"`
}
