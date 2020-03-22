package middleware

import (
	"context"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type ContextKey string

const ContextRequestID ContextKey = "requestID"
const ContextMSRequestID ContextKey = "xMsRequestId"
const ContextRemoteAddr ContextKey = "clientIP"
const ContextRequestURI ContextKey = "requestURI"

// Logger provides http middleware for logging additional information about the request
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		remoteAddr := r.RemoteAddr
		ctx = context.WithValue(ctx, ContextRemoteAddr, remoteAddr)

		requestURI := r.RequestURI
		ctx = context.WithValue(ctx, ContextRequestURI, requestURI)

		requestID := r.Header.Get("X-Request-Id")
		ctx = context.WithValue(ctx, ContextRequestID, requestID)

		msRequestID := r.Header.Get("X-Ms-Request-Id")
		ctx = context.WithValue(ctx, ContextMSRequestID, msRequestID)

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)

	})
}

// LogWithContext is used as a helper for logging http request information
func LogWithContext(ctx context.Context) *log.Entry {
	entry := log.NewEntry(log.StandardLogger())

	requestIDKey := "requestID"
	if requestID := ctx.Value(ContextRequestID); requestID != nil {
		entry = entry.WithField(requestIDKey, requestID)
	} else {
		log.Debug("No X-Request-Id header found")
		entry = entry.WithField(requestIDKey, "-")
	}

	msRequestIDKey := "msRequestID"
	if msRequestID := ctx.Value(ContextMSRequestID); msRequestID != nil {
		entry = entry.WithField(msRequestIDKey, msRequestID)
	} else {
		log.Debug("No X-Request-Id header found")
		entry = entry.WithField(msRequestIDKey, "-")
	}

	remoteAddrKey := "remoteAddr"
	if remoteAddr := ctx.Value(ContextRemoteAddr); remoteAddr != nil {
		entry = entry.WithField(remoteAddrKey, remoteAddr)
	} else {
		entry = entry.WithField(remoteAddrKey, "-")
	}

	requestURIKey := "requestURI"
	if requestURI := ctx.Value(ContextRequestURI); requestURI != nil {
		entry = entry.WithField(requestURIKey, requestURI)
	} else {
		entry = entry.WithField(requestURIKey, "-")
	}

	return entry
}
