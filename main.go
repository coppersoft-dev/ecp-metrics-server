package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"go.e13.dev/playground/ecp-metrics-server/cd"
	"go.e13.dev/playground/ecp-metrics-server/server"
)

func parseVerbosity(v string) (*int, error) {
	if v == "" {
		return nil, nil
	}
	lvl, err := strconv.Atoi(v)
	if err != nil {
		return nil, fmt.Errorf("failed parsing verbosity level %q: %v", v, err)
	}
	lvl = -lvl
	return &lvl, nil
}

type LogHandlerType string

const (
	JSONHandler LogHandlerType = "json"
	TextHandler LogHandlerType = "text"
)

func newLogger(level *slog.LevelVar, hdlrType LogHandlerType) (*slog.Logger, error) {
	var hdlr slog.Handler
	opts := &slog.HandlerOptions{
		Level: level,
	}
	switch hdlrType {
	case JSONHandler:
		hdlr = slog.NewJSONHandler(os.Stdout, opts)
	case TextHandler:
		hdlr = slog.NewTextHandler(os.Stdout, opts)
	default:
		return nil, fmt.Errorf("logging handler %q unknown", hdlrType)
	}

	return slog.New(hdlr), nil
}

func main() {
	dbHost := os.Getenv("CD_DB_HOST")
	dbName := os.Getenv("CD_DB_NAME")
	dbUser := os.Getenv("CD_DB_USER")
	dbPass := os.Getenv("CD_DB_PASS")
	authToken := os.Getenv("AUTH_TOKEN")
	lHost := os.Getenv("LISTEN_HOST")
	if lHost == "" {
		lHost = "127.0.0.1" // safe fallback
	}
	lPort := os.Getenv("LISTEN_PORT")
	if lPort == "" {
		lPort = "8080"
	}

	logLevel := new(slog.LevelVar)
	logLevel.Set(slog.LevelInfo)

	var hdlrType LogHandlerType = LogHandlerType(os.Getenv("LOG_FORMAT"))
	if hdlrType == "" {
		hdlrType = TextHandler
	}
	log, err := newLogger(logLevel, LogHandlerType(hdlrType))
	if err != nil {
		panic(fmt.Sprintf("failed creating logger: %s", err))
	}

	customLogLevel, err := parseVerbosity(os.Getenv("VERBOSITY"))
	if err != nil {
		log.Error("failed parsing verbosity", "error", err)
		os.Exit(1)
	}
	if customLogLevel != nil {
		logLevel.Set(slog.Level(*customLogLevel))
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error)

	connStr := fmt.Sprintf("postgresql://%s:%s@%s:5432/%s?sslmode=require", dbUser, dbPass, dbHost, dbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Error("failed opening DB connection", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// set reasonable connection pooling limits. These limits
	// are rather low because the applcation currently really
	// only needs a single open connection.
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(3)

	go func(ctx context.Context) {
		ticker := time.NewTicker(30 * time.Second)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				log.Debug("db stats", "max_open_conns", db.Stats().MaxOpenConnections, "in_use", db.Stats().InUse, "idle", db.Stats().Idle)
			}
		}
	}(ctx)

	srv := server.New(authToken, lHost, lPort, log.With("component", "server"))
	go srv.Start(ctx)

	cdSvc := cd.NewService(db, &srv, log.With("component", "cd_service"))
	go cdSvc.StartPollLoop(ctx, errCh)

	ret := 0

	select {
	case sig := <-sigChan:
		log.Info("received signal, exiting", "signal", sig)
		cancel()
	case <-ctx.Done():
		log.Info("shutting down application")
	case <-errCh:
		ret = 1
		cancel()
	}

	timeout := time.NewTimer(5 * time.Second)
	// wait for server to shut down
	select {
	case <-timeout.C:
		log.Error("server took too long to shut down, continuing shutdown procedure")
		ret = 1
	case <-srv.ShutdownCh():
		log.Info("server shutdown complete")
		timeout.Stop()
	}

	timeout.Reset(5 * time.Second)
	// wait for CD service to shut down
	select {
	case <-timeout.C:
		log.Error("CD service took too long to shut down, continuing shutdown procedure")
		ret = 1
	case <-cdSvc.ShutdownCh():
		log.Info("CD service shutdown complete")
		timeout.Stop()
	}

	log.Info("exiting")
	os.Exit(ret)
}
