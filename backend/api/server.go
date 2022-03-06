package api

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"moneropot/util"
	"net"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

// Server defines how api request is handled
type (
	Server struct {
		router   *mux.Router
		validate *validator.Validate
	}

	validationError struct {
		Params map[string]string `json:"params"`
	}

	alertError struct {
		Title   string `json:"title"`
		Message string `json:"message"`
	}

	rpcType struct {
		contentType string
		body        []byte
	}
)

var (
	errAuth      = fmt.Errorf("not logged in")
	errForbidden = fmt.Errorf("forbidden")
	errNotFound  = fmt.Errorf("not found")
	errRateLimit = fmt.Errorf("rate limit")

	StaticFS embed.FS
)

// Error validation error
func (ve validationError) Error() string {
	return fmt.Sprintf("Validation Error: %v", ve.Params)
}

// Error validation error
func (ae alertError) Error() string {
	return fmt.Sprintf("Alert Error: %s - %s", ae.Title, ae.Message)
}

// NewServer returns the instance of api server that implements the Handler interface
func NewServer() *Server {
	r := mux.NewRouter()

	srv := &Server{
		router:   r,
		validate: validator.New(),
	}
	vAlias := map[string]string{
		"invalid": "min=95,max=95",
		"rfnn":    "omitempty,min=1",
	}
	for k, v := range vAlias {
		srv.validate.RegisterAlias(k, v)
	}
	validUsername := regexp.MustCompile(`^[A-Za-z][A-Za-z0-9-_]{2,25}$`)
	// register function to get tag name from json tags.
	srv.validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	srv.validate.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		return validUsername.Match([]byte(fl.Field().String()))
	})
	r.Use(srv.limits)
	sr := r.PathPrefix("/api").Subrouter()

	sr.HandleFunc("/accounts", srv.handlePostAccount()).Methods(http.MethodPost)
	sr.HandleFunc("/info", srv.handleGetInfo()).Methods(http.MethodGet)
	sr.HandleFunc("/entries", srv.handleGetEntries()).Methods(http.MethodGet)
	sr.HandleFunc("/events", util.HandleEvents).Methods(http.MethodGet)

	// internal is subject to changes without notice
	sr.HandleFunc("/internal/{method}", func(w http.ResponseWriter, r *http.Request) {
		method := mux.Vars(r)["method"]
		m := reflect.ValueOf(srv).MethodByName(method)
		if m.IsValid() {
			t := reflect.TypeOf(m.Interface())
			if t.NumIn() == 2 {
				m.Call([]reflect.Value{reflect.ValueOf(w), reflect.ValueOf(r)})
				return
			}
			rr := m.Call([]reflect.Value{reflect.ValueOf(r)})
			result := rr[0].Interface()
			if result != nil {
				if err, ok := result.(error); ok {
					srv.writeError(w, err)
					return
				} else if rt, ok := result.(rpcType); ok {
					w.Header().Set("Content-Type", rt.contentType)
					w.Write(rt.body)
					return
				}
			}
			srv.writeJSON(w, result)
			return
		}
		srv.writeError(w, errNotFound)
	}).Methods(http.MethodPost, http.MethodGet)

	// catch all
	r.PathPrefix("/").HandlerFunc(srv.handleCatchAll()).Methods(http.MethodGet)
	return srv
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	st := util.UtcNow()
	if !util.Config.Production {
		log.Println(r.Method, r.URL.Path)
	}
	s.router.ServeHTTP(w, r)
	if tt := time.Since(st).Milliseconds(); tt > 1000 {
		log.Println("Slow request:", r.Method, r.RequestURI, "took", tt, "ms")
	}
}

func (s *Server) isAdmin(r *http.Request) bool {
	cKey := "isAdmin:" + s.RealIP(r)
	_, ok := util.Cache.Get(cKey)
	if ok {
		return false
	}

	valid := r.Header.Get("X-Key") == util.Config.AdminKey
	if !valid {
		// brute force protection
		util.Cache.Set(cKey, "1", time.Hour*1)
	}

	return valid
}

func (s *Server) handler(f func(r *http.Request) interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result := f(r)
		if result != nil {
			if err, ok := result.(error); ok {
				s.writeError(w, err)
				return
			}
		}
		s.writeJSON(w, result)
	}
}

func (s *Server) writeError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	var msg string
	code := http.StatusInternalServerError

	if errMsg := err.Error(); errMsg != "" {
		if strings.Contains(errMsg, "http: request body too large") {
			code = http.StatusRequestEntityTooLarge
		} else if strings.Contains(errMsg, "JSON input") {
			code = http.StatusBadRequest
		}
	}
	resp := map[string]interface{}{}

	if e, ok := err.(validationError); ok {
		code = http.StatusBadRequest
		msg = "validation"
		resp["params"] = e.Params
	} else if e, ok := err.(alertError); ok {
		code = http.StatusBadRequest
		msg = e.Message
		resp["title"] = e.Title
	} else if _, ok := err.(*json.UnmarshalTypeError); ok {
		log.Printf("UnmarshalTypeError: %v", err)
		code = http.StatusBadRequest
	} else if _, ok := err.(*json.SyntaxError); ok {
		log.Printf("JsonSyntaxError: %v", err)
		code = http.StatusBadRequest
	} else if err == errNotFound {
		code = http.StatusNotFound
	} else if err == errAuth {
		code = http.StatusUnauthorized
	} else if err == errForbidden {
		code = http.StatusForbidden
	} else if err == errRateLimit {
		code = http.StatusServiceUnavailable
	} else {
		log.Printf("InternalError: %v", err)
	}

	if msg == "" && code > 0 {
		msg = http.StatusText(code)
	}
	resp["error"] = msg
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		panic(err)
	}
	return true
}

func (s *Server) writeJSON(w http.ResponseWriter, resp interface{}) {
	if resp == nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		panic(err)
	}
}

func (s *Server) readJSON(r *http.Request, out interface{}) error {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		return err
	}
	if err := r.Body.Close(); err != nil {
		return err
	}
	if err := json.Unmarshal(body, out); err != nil {
		return err
	}
	return nil
}

func (s *Server) bind(r *http.Request, out interface{}) error {
	if err := s.readJSON(r, out); err != nil {
		return err
	}
	if err := s.validationError(out); err != nil {
		return err
	}
	return nil
}

func (s *Server) validationError(src interface{}) error {
	err := s.validate.Struct(src)
	if err != nil {
		vr := validationError{Params: make(map[string]string)}
		for _, err := range err.(validator.ValidationErrors) {
			ss := strings.Split(err.Namespace(), ".")
			ss = ss[1:]
			vr.Params[strings.Join(ss, ".")] = err.Tag()
		}
		return vr
	}
	return nil
}

func (s *Server) limits(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1mb for any regular api requests but 10 mb for uploads
		var maxBodyLimit int64 = 1 << 20
		if r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/api/blobs") {
			maxBodyLimit = 10 << 20
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyLimit)
		next.ServeHTTP(w, r)
	})
}

// QueryParam returns query parameter by name
func (s *Server) QueryParam(r *http.Request, name string) string {
	if v, ok := r.URL.Query()[name]; ok && len(v) > 0 {
		return v[0]
	}
	return ""
}

func newValidationErr(params ...string) validationError {
	p := make(map[string]string)
	for i := 0; i < len(params); i += 2 {
		if i+1 < len(params) {
			p[params[i]] = params[i+1]
		}
	}
	return validationError{Params: p}
}

func (s *Server) handleCatchAll() http.HandlerFunc {
	contentStatic, err := fs.Sub(StaticFS, "dist")
	if err != nil {
		log.Fatalf("handleCatchAll error %v", err)
	}
	handler := http.FileServer(http.FS(contentStatic))

	return func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if p == "/" || strings.Contains(p, "/css/") || strings.Contains(p, "/js/") ||
			strings.Contains(p, "/img/") || strings.Contains(p, "favicon.ico") ||
			strings.Contains(p, "index.html") {
			handler.ServeHTTP(w, r)
		} else {
			r = r.Clone(r.Context())
			r.URL.Path = "/"
			handler.ServeHTTP(w, r)
		}
	}
}

func (s *Server) RealIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ", ")[0]
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	ra, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ra
}
