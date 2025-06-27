package cd

import (
	"database/sql"
	"log/slog"

	"go.e13.dev/playground/ecp-metrics-server/cd/types"
)

type ComponentsSetter interface {
	SetComponents(c types.Components)
}

type Service struct {
	db         *sql.DB
	log        *slog.Logger
	cs         ComponentsSetter
	shutdownCh chan struct{}
}

func NewService(db *sql.DB, cs ComponentsSetter, log *slog.Logger) Service {
	return Service{
		db:         db,
		log:        log,
		cs:         cs,
		shutdownCh: make(chan struct{}),
	}
}

func (s *Service) ShutdownCh() <-chan struct{} {
	return s.shutdownCh
}
