package main

import (
	"crypto/subtle"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"

	"go.e13.dev/playground/ecp-metrics-server/cd"
)

func authReq(w http.ResponseWriter, r *http.Request, expectedToken string) error {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		w.Header().Set("WWW-Authenticate", `Bearer realm="Restricted"`)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "Unauthorized: Missing or invalid Authorization header")
		return fmt.Errorf("missing bearer token in Authorization")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) != 1 {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "Unauthorized: Invalid token")
		return fmt.Errorf("invalid token")
	}

	return nil
}

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
	var contentMux sync.RWMutex
	var components cd.Components

	dbHost := os.Getenv("CD_DB_HOST")
	dbName := os.Getenv("CD_DB_NAME")
	dbUser := os.Getenv("CD_DB_USER")
	dbPass := os.Getenv("CD_DB_PASS")
	authToken := os.Getenv("AUTH_TOKEN")

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

	go func() {
		timer := time.NewTimer(0)
		var prevErr bool

		for {
			select {
			case <-timer.C:

				cdContent, err := cd.GetCDContent(dbUser, dbPass, dbHost, dbName)
				if err != nil {
					log.Error("failed reading CD content from database", "error", err)
					prevErr = true
					timer.Reset(3 * time.Second)
					continue
				}
				componentsCur, err := cd.ParseCDContent(cdContent)
				if err != nil {
					log.Error("failed parsing CD content", "error", err)
					prevErr = true
					timer.Reset(3 * time.Second)
					continue
				}

				logF := log.Debug
				if prevErr {
					prevErr = false
					logF = log.Info
				}
				logF("Parsed CD content")

				contentMux.Lock()
				components = componentsCur
				contentMux.Unlock()

				timer.Reset(10 * time.Second)
			}
		}
	}()

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		if err := authReq(w, r, authToken); err != nil {
			return
		}
		for _, broker := range components.ComponentList.Brokers {
			fmt.Fprintf(w, "ecp_component_version{code=\"%s\",org=\"%s\",type=\"%s\",version=\"%s\"} 1\n", broker.Code, broker.Organization, broker.Type, broker.MADESImplementation.Version)
		}

		for _, ep := range components.ComponentList.Endpoints {
			fmt.Fprintf(w, "ecp_component_version{code=\"%s\",org=\"%s\",type=\"%s\",version=\"%s\"} 1\n", ep.Code, ep.Organization, ep.Type, ep.MADESImplementation.Version)
		}

		for _, cd := range components.ComponentList.ComponentDirectories {
			fmt.Fprintf(w, "ecp_component_version{code=\"%s\",org=\"%s\",type=\"%s\",version=\"%s\"} 1\n", cd.Code, cd.Organization, cd.Type, cd.MADESImplementation.Version)
		}
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Error("HTTP listener failed")
		os.Exit(1)
	}
}
