package httpapi

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
)

// Config defines server configuration
type Config struct {
	Log       *logrus.Entry
	BasicAuth *BasicAuth
}

// BasicAuth defines basic auth configuration
type BasicAuth struct {
	Username string
	Password string
}

// Server holds the configuration, router and http server
type Server struct {
	Config Config
	Router *chi.Mux
}

// NewServer creates a new HTTP server
func NewServer(config Config) Server {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/health"))
	r.Use(basicAuthMiddleware(config))

	s := Server{
		Config: config,
		Router: r,
	}

	return s
}

// RunAndBlock starts the HTTP server and blocks
func (s Server) RunAndBlock() error {
	auth := "none"
	if s.Config.BasicAuth != nil {
		auth = "basic"
	}

	addr := ":8113"

	l := s.Config.Log.WithFields(logrus.Fields{
		"component": "server",
		"address":   addr,
		"auth":      auth,
	})

	if auth != "basic" {
		l.Warnf("listening on %s", addr)
	} else {
		l.Infof("listening on %s", addr)
	}

	return http.ListenAndServe(addr, s.Router)
}

func basicAuthMiddleware(config Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ba := config.BasicAuth
			if ba == nil {
				next.ServeHTTP(w, r)
				return
			}

			user, pass, _ := r.BasicAuth()

			if (*ba).Username != user || (*ba).Password != pass {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized.", http.StatusUnauthorized)
				config.Log.WithFields(logrus.Fields{
					"service":    "HTTP-Server",
					"middleware": "basic-auth",
				}).Debugf("authentication failed")
				return
			}

			config.Log.WithFields(logrus.Fields{
				"service":    "HTTP-Server",
				"middleware": "basic-auth",
			}).Debugf("successfully authenticated")

			next.ServeHTTP(w, r)
		})
	}
}
