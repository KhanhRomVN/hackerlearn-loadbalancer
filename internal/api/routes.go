package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"hackerlearn-loadbalancer/internal/config"
	"hackerlearn-loadbalancer/internal/logger"
	"hackerlearn-loadbalancer/internal/service"
)

type Server struct {
	router  *mux.Router
	config  *config.Config
	handler *Handlers
}

func NewServer(cfg *config.Config, lb *service.LoadBalancer, logger *logger.Logger) *Server {
	handler := NewHandlers(lb, logger)
	router := mux.NewRouter()

	// Register routes
	router.HandleFunc("/loadbalancer", handler.HandleLoadBalancer).Methods("POST", "OPTIONS")
	router.HandleFunc("/api-check", handler.HandleAPICheck).Methods("GET")
	router.HandleFunc("/ping", handler.HandlePing).Methods("GET")

	return &Server{
		router:  router,
		config:  cfg,
		handler: handler,
	}
}

func (s *Server) Start() error {
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Content-Length", "Authorization", "accessToken"},
		AllowCredentials: true,
	})

	srv := &http.Server{
		Handler:      corsHandler.Handler(s.router),
		Addr:         s.config.ServerPort,
		WriteTimeout: s.config.WriteTimeout,
		ReadTimeout:  s.config.ReadTimeout,
	}

	return srv.ListenAndServe()
}
