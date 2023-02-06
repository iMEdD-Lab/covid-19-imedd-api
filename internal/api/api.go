package api

import (
	"context"
	"encoding/json"
	"fmt"
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
	"covid19-greece-api/pkg/vartypes"
)

const (
	perPageDefault = 100
)

var tlFields = []string{
	"daily_cases",
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
	"cases_cum",
	"waste_highest_place",
	"waste_highest_percent",
}

type Api struct {
	Router  *chi.Mux
	repo    data.Repo
	cache   cache.Cache
	dataSrv *data.Service
	secret  string
}

// NewApi initiates and API struct
func NewApi(
	repo data.Repo,
	dataSrv *data.Service,
	secret string,
) *Api {
	api := Api{
		repo:    repo,
		cache:   cache.NewMemoryCacher(),
		dataSrv: dataSrv,
		secret:  secret,
	}
	api.initRouter()

	return &api
}

func (a *Api) initRouter() {
	// initiate API Router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// use rate limit of 100 requests per minute
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	// use cache middleware
	r.Use(a.cacheMw)

	// be open to CORS requests, allow only GET
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:     []string{"*"},
		AllowOriginFunc:    func(r *http.Request, origin string) bool { return true },
		AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:     []string{"Link"},
		AllowCredentials:   true,
		OptionsPassthrough: true,
		MaxAge:             3599, // Maximum value not ignored by any of major browsers
	}))

	// health status
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello friend."))
		w.WriteHeader(http.StatusOK)
	})

	// cached routes
	r.Group(func(r chi.Router) {
		r.Use(a.cacheMw)

		// helper endpoint
		r.Get("/regional_units", func(w http.ResponseWriter, r *http.Request) {
			rus, err := a.repo.GetRegionalUnits(r.Context())
			if err != nil {
				log.Println(err)
				a.respondError(w, r, http.StatusInternalServerError, nil)
				return
			}
			a.respond200(w, r, rus, false)
		})

		// helper endpoint
		r.Get("/municipalities", func(w http.ResponseWriter, r *http.Request) {
			municipalities, err := a.repo.GetMunicipalities(r.Context())
			if err != nil {
				log.Println(err)
				a.respondError(w, r, http.StatusInternalServerError, nil)
				return
			}
			p := getPagination(r.URL.Query(), len(municipalities))
			a.respond200(w, r, municipalities[p.start:p.end], false)
		})

		// COVID-19 deaths per Greek municipality
		r.Get("/deaths_per_municipality", func(w http.ResponseWriter, r *http.Request) {
			f := deathsFilter(r.URL.Query())
			municipalities, err := a.repo.GetDeathsPerMunicipality(r.Context(), f)
			if err != nil {
				log.Println(err)
				a.respondError(w, r, http.StatusInternalServerError, nil)
				return
			}
			p := getPagination(r.URL.Query(), len(municipalities))
			a.respond200(w, r, municipalities[p.start:p.end], false)
		})

		// COVID-19 deaths per Greek prefecture
		r.Get("/cases", func(w http.ResponseWriter, r *http.Request) {
			filter := casesFilter(r.URL.Query())
			cases, err := a.repo.GetCases(r.Context(), filter)
			if err != nil {
				log.Println(err)
				a.respondError(w, r, http.StatusInternalServerError, nil)
				return
			}
			p := getPagination(r.URL.Query(), len(cases))
			a.respond200(w, r, cases[p.start:p.end], false)
		})

		// helper endpoint
		r.Get("/timeline_fields", func(w http.ResponseWriter, r *http.Request) {
			a.respond200(w, r, tlFields, false)
		})

		// returns full COVID-19 info for every date of a specific period
		r.Get("/timeline", func(w http.ResponseWriter, r *http.Request) {
			tlf := timelineFilter(r.URL.Query())
			info, err := a.repo.GetFromTimeline(r.Context(), tlf.DatesFilter)
			if err != nil {
				log.Println(err)
				a.respondError(w, r, http.StatusInternalServerError, nil)
				return
			}
			p := getPagination(r.URL.Query(), len(info))
			if len(tlf.Fields) > 0 {
				a.respond200(w, r, keepFields(tlf.Fields, info)[p.start:p.end], false)
				return
			}
			a.respond200(w, r, info[p.start:p.end], false)
		})

		// same as /timeline, but for a specific field (for example, "total_reinfections")
		r.Get("/{field}", func(w http.ResponseWriter, r *http.Request) {
			field := chi.URLParam(r, "field")
			info, err := a.repo.GetFromTimeline(r.Context(), datesFilter(r.URL.Query()))
			if err != nil {
				log.Println(err)
				a.respondError(w, r, http.StatusInternalServerError, nil)
				return
			}
			p := getPagination(r.URL.Query(), len(info))
			a.respond200(w, r, keepFields([]string{field}, info)[p.start:p.end], false)
		})

		// returns COVID19 demographics info by date
		r.Get("/demographics", func(w http.ResponseWriter, r *http.Request) {
			info, err := a.repo.GetDemographicInfo(r.Context(), demographicsFilter(r.URL.Query()))
			if err != nil {
				log.Println(err)
				a.respondError(w, r, http.StatusInternalServerError, nil)
				return
			}
			p := getPagination(r.URL.Query(), len(info))
			a.respond200(w, r, info[p.start:p.end], false)
		})

	})

	// authentication protected routes
	r.Group(func(r chi.Router) {
		r.Use(a.authMw)

		// same as health, but only for authenticated users
		r.Get("/check_auth", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("hello friend, you are authenticated!"))
			w.WriteHeader(http.StatusOK)
		})

		// same as health, but only for authenticated users
		r.Get("/refresh", func(w http.ResponseWriter, r *http.Request) {
			go func() {
				if err := a.dataSrv.PopulateEverything(context.Background()); err != nil {
					log.Printf("data refresh failed: %v", err)
				}
			}()
			if err := a.cache.Flush(); err != nil {
				log.Printf("cache could not be flushed: %v", err)
			}
			w.WriteHeader(http.StatusOK)
		})
	})

	a.Router = r
}

// demographicsFilter initializes filter for demographics query
func demographicsFilter(values url.Values) data.DemographicFilter {
	f := data.DemographicFilter{}
	for k, v := range values {
		switch k {
		case "category":
			f.Category = v[0]
		}
	}
	f.DatesFilter = datesFilter(values)

	return f
}

// deathsFilter initializes filter for deaths query
func deathsFilter(values url.Values) data.DeathsFilter {
	f := data.DeathsFilter{}
	for k, v := range values {
		switch k {
		case "year":
			f.Year = vartypes.StringToInt(v[0])
		case "municipality_id":
			f.MunId = vartypes.StringToInt(v[0])
		}
	}

	return f
}

type TimelineFilter struct {
	data.DatesFilter
	Fields []string
}

// timelineFilter initializes filter for timeline
func timelineFilter(values url.Values) TimelineFilter {
	var fields []string
	if len(values["fields"]) > 0 {
		fields = strings.Split(values["fields"][0], ",")
	}
	return TimelineFilter{
		DatesFilter: datesFilter(values),
		Fields:      fields,
	}
}

// datesFilter initializes filter for start and end date of a time period
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

// casesFilter initializes filter for cases per regional unit
func casesFilter(values url.Values) data.CasesFilter {
	f := data.CasesFilter{}
	f.DatesFilter = datesFilter(values)
	for k, v := range values {
		switch k {
		case "regional_unit_id":
			f.RegionalUnitId = vartypes.StringToInt(v[0])
		}
	}

	return f
}

// keepFields is a helper function for returning specific fields of timeline full info.
func keepFields(fields []string, fullInfos []data.FullInfo) []map[string]interface{} {
	var res []map[string]interface{}
	for _, fi := range fullInfos {
		r := make(map[string]interface{})
		r["date"] = fi.Date
		for _, f := range fields {
			switch f {
			case "daily_cases":
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
			case "cases_cum":
				r[f] = fi.CasesCum
			case "waste_highest_place":
				r[f] = fi.WasteHighestPlace
			case "waste_highest_place_en":
				r[f] = fi.WasteHighestPlaceEn
			case "waste_highest_percent":
				r[f] = fi.WasteHighestPercent
			default:
				// do nothing
			}
		}
		res = append(res, r)
	}

	return res
}

// authMw is the authentication middleware function. Currently a bit useless as we don't have authentication
func (a *Api) authMw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := ExtractTokenFromRequest(r)
		if err := a.Authenticate(tokenStr); err != nil {
			a.respondError(w, r, http.StatusUnauthorized, ErrorResp{"unauthorized"})
			return
		}
		next.ServeHTTP(w, r.WithContext(r.Context()))
	})
}

func ExtractTokenFromRequest(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearToken, "Bearer ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

// Authenticate authenticates requests
func (a *Api) Authenticate(token string) error {
	if token == a.secret {
		return nil
	}
	return fmt.Errorf("authentication failed")
}

// cacheMw is the middleware function for caching our responses
func (a *Api) cacheMw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && a.cache.IsExist(r.URL.RequestURI()) {
			content := a.cache.Get(r.URL.RequestURI())
			b, _ := json.Marshal(content)
			w.Header().Set("Content-Length", strconv.Itoa(len(b)))
			w.Header().Set("Access-Control-Allow-Origin", "*")
			a.respond200(w, r, content, true)
			return
		}
		next.ServeHTTP(w, r.WithContext(r.Context()))
	})
}

// respondError helper function for erroneous API responses
func (a *Api) respondError(w http.ResponseWriter, r *http.Request, statusCode int, content interface{}) {
	w.WriteHeader(statusCode)
	bytes, err := json.Marshal(content)
	if err != nil {
		log.Println("failed to marshal ErrorResp:", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(bytes)
}

// respondError helper function for successful API responses
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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(bytes)
}

type ErrorResp struct {
	Msg string `json:"message"`
}

type pagination struct {
	page    int
	perPage int
	start   int
	end     int
}

// getPagination returns pagination values extracted from the request
func getPagination(values url.Values, count int) pagination {
	perPage := perPageDefault
	page := 1

	if pp, ok := values["per_page"]; ok {
		perPage = vartypes.StringToInt(pp[0])
		if perPage <= 0 {
			perPage = perPageDefault
		}
	}

	if p, ok := values["page"]; ok {
		page = vartypes.StringToInt(p[0])
		if page <= 0 {
			page = 1
		}
	}

	start := (page - 1) * perPage
	if start >= count {
		return pagination{
			page:    page,
			perPage: perPage,
			start:   0,
			end:     0,
		}
	}
	end := start + perPage
	if end > count {
		end = count
	}

	return pagination{
		page:    page,
		perPage: perPage,
		start:   start,
		end:     end,
	}
}
