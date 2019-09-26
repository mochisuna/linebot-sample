package handler

import (
	"net/http"
	"time"

	"github.com/mochisuna/linebot-sample/domain/service"
	"github.com/unrolled/render"
	validator "gopkg.in/go-playground/validator.v9"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
)

var rendering = render.New(render.Options{})
var validate = validator.New()

// Services is grouping application services structure
type Services struct {
	CallbackService service.CallbackService
}

// Server HTTP server
type Server struct {
	*http.Server
	*Services
	*Line
}

// New inject to domain services
func New(addr string, services *Services, line *Line) *Server {
	return &Server{
		Server: &http.Server{
			Addr: addr,
		},
		Services: services,
		Line:     line,
	}
}

// ListenAndServe override http ListenAndServe
func (s *Server) ListenAndServe() error {
	r := chi.NewRouter()

	// cord option
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	r.Use(cors.Handler)

	// 公式提供のmiddleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// routings
	// 開発レベルでバージョン分けする可能性がゼロではないので一応バージョンをラベル切っておく
	// APIコールしようかなとも考えたが、callback内で解決した方が安全な気がしたので一旦他にルーティングしない
	r.Route("/v1", func(r chi.Router) {
		r.Post("/callback", s.callback)
	})
	r.Route("/health", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	s.Handler = r
	return s.Server.ListenAndServe()
}
