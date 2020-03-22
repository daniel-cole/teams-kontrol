package main

import (
	"context"
	"github.com/daniel-cole/teams-kontrol/command"
	"github.com/daniel-cole/teams-kontrol/healthz"
	"github.com/daniel-cole/teams-kontrol/k8s"
	"github.com/daniel-cole/teams-kontrol/middleware"
	"github.com/daniel-cole/teams-kontrol/teams"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	setupLogging()

	err := k8s.CreateClient()
	if err != nil {
		logrus.Fatalf("failed to create k8s client %v", err)
	}

	// initialise teams and command packages for loading in environment files
	teams.Init()
	command.Init()

	//  add handlers
	healthzHandler := http.HandlerFunc(healthz.Handler)
	http.Handle("/healthz", middleware.Logger(healthzHandler))
	http.Handle("/teams", middleware.Logger(teams.AuthHandler(http.HandlerFunc(teams.MessageHandler))))

	// only load /command endpoint if specified in environment variable
	// this handler is insecure and should not be loaded in production if you are exposing it externally
	if os.Getenv(command.KontrolInsecureCommandHandlerEnvKey) == "TRUE" {
		logrus.Warnf("Detected %s set to 'TRUE'. Loading insecure command endpoint on /command", command.KontrolInsecureCommandHandlerEnvKey)
		http.Handle("/command", middleware.Logger(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				command.Handler(k8s.Client, w, r)
			})))
	}

	listenAddr := "0.0.0.0:9000"

	server := &http.Server{
		Addr:         listenAddr,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	done := make(chan bool, 1)
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		graceTime := 30 * time.Second
		logrus.Infof("Server is shutting down... grace period: %s", graceTime.String())

		ctx, cancel := context.WithTimeout(context.Background(), graceTime)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			logrus.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	tlsCertFile := os.Getenv("TEAMS_KONTROL_TLS_CERT")
	tlsKeyFile := os.Getenv("TEAMS_KONTROL_TLS_KEY")

	if tlsCertFile != "" && tlsKeyFile != "" { // TLS enabled
		logrus.Info("Starting server with TLS enabled")
		logrus.Infof("TLS key: %s", tlsKeyFile)
		logrus.Infof("TLS cert: %s", tlsCertFile)

		if err := server.ListenAndServeTLS(tlsCertFile, tlsKeyFile); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Could not listen on %s: %v\n", "0.0.0.0:9000", err)
		}

	} else {
		logrus.Info("Starting server without TLS enabled")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Could not listen on %s: %v\n", "0.0.0.0:9000", err)
		}
	}

	<-done

	logrus.Info("Server stopped")
}

func setupLogging() {

	logLevel := os.Getenv("TEAMS_KONTROL_LOG_LEVEL")
	switch logLevel {
	case "": // default to INFO level logging
		logrus.SetLevel(logrus.InfoLevel)
	case "INFO":
		logrus.SetLevel(logrus.InfoLevel)
	case "WARN":
		logrus.SetLevel(logrus.WarnLevel)
	case "ERROR":
		logrus.SetLevel(logrus.ErrorLevel)
	case "DEBUG":
		logrus.SetLevel(logrus.DebugLevel)
	}
	logrus.Info("Log level set to: " + logLevel)
}
