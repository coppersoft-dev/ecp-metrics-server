package main

import (
	"crypto/subtle"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
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

func main() {

	var contentMux sync.RWMutex
	var components cd.Components

	dbHost := os.Getenv("CD_DB_HOST")
	dbName := os.Getenv("CD_DB_NAME")
	dbUser := os.Getenv("CD_DB_USER")
	dbPass := os.Getenv("CD_DB_PASS")
	authToken := os.Getenv("AUTH_TOKEN")

	slog.SetLogLoggerLevel(slog.LevelDebug)

	go func() {
		timer := time.NewTimer(0)
		var prevErr bool

		for {
			select {
			case <-timer.C:

				cdContent, err := cd.GetCDContent(dbUser, dbPass, dbHost, dbName)
				if err != nil {
					slog.Error("failed reading CD content from database", "error", err)
					prevErr = true
					timer.Reset(3 * time.Second)
					continue
				}
				componentsCur, err := cd.ParseCDContent(cdContent)
				if err != nil {
					slog.Error("failed parsing CD content", "error", err)
					prevErr = true
					timer.Reset(3 * time.Second)
					continue
				}

				logF := slog.Debug
				if prevErr {
					prevErr = false
					logF = slog.Info
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

	log.Fatal(http.ListenAndServe(":8080", nil))
}
