package server

import (
	"context"
	"crypto/subtle"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.e13.dev/playground/ecp-metrics-server/cd"
	"go.e13.dev/playground/ecp-metrics-server/cd/types"
)

type Server struct {
	authToken    string
	host         string
	port         string
	cdComponents types.Components
	compMux      sync.RWMutex
	log          *slog.Logger
	shutdownCh   chan struct{}
}

var _ cd.ComponentsSetter = &Server{}

func (s *Server) authReq(w http.ResponseWriter, r *http.Request) error {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		w.Header().Set("WWW-Authenticate", `Bearer realm="Restricted"`)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "Unauthorized: Missing or invalid Authorization header")
		return fmt.Errorf("missing bearer token in Authorization")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if subtle.ConstantTimeCompare([]byte(token), []byte(s.authToken)) != 1 {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "Unauthorized: Invalid token")
		return fmt.Errorf("invalid token")
	}

	return nil
}

func New(authToken string, host, port string, log *slog.Logger) Server {
	return Server{
		authToken:  authToken,
		log:        log,
		host:       host,
		port:       port,
		shutdownCh: make(chan struct{}),
	}
}

func (s *Server) SetComponents(c types.Components) {
	s.compMux.Lock()
	defer s.compMux.Unlock()
	s.cdComponents = c
}

func (s *Server) ShutdownCh() <-chan struct{} {
	return s.shutdownCh
}

func (s *Server) Start(ctx context.Context) {
	srvMux := http.NewServeMux()
	srvMux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		if err := s.authReq(w, r); err != nil {
			return
		}
		s.compMux.RLock()
		defer s.compMux.RUnlock()
		for _, broker := range s.cdComponents.ComponentList.Brokers {
			fmt.Fprintf(w, "ecp_component_version{code=\"%s\",org=\"%s\",type=\"%s\",version=\"%s\"} 1\n", broker.Code, broker.Organization, broker.Type, broker.MADESImplementation.Version)
		}

		for _, ep := range s.cdComponents.ComponentList.Endpoints {
			fmt.Fprintf(w, "ecp_component_version{code=\"%s\",org=\"%s\",type=\"%s\",version=\"%s\"} 1\n", ep.Code, ep.Organization, ep.Type, ep.MADESImplementation.Version)
		}

		for _, cd := range s.cdComponents.ComponentList.ComponentDirectories {
			fmt.Fprintf(w, "ecp_component_version{code=\"%s\",org=\"%s\",type=\"%s\",version=\"%s\"} 1\n", cd.Code, cd.Organization, cd.Type, cd.MADESImplementation.Version)
		}
	})

	srv := http.Server{
		Handler: srvMux,
		Addr:    s.host + ":" + s.port,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
		s.log.Info("HTTP server shut down")
		close(s.shutdownCh)
	}()

	s.log.Info("starting server", "addr", srv.Addr)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.log.Error("HTTP listener failed", "error", err)
	}
}
