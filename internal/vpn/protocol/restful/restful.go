package restful

import (
	"context"
	"encoding/json"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net/http"
	"os"
	"os/signal"
	"squirrel-srv/internal/vpn/protocol/restful/middleware"
	"squirrel-srv/pkg/api/v1"
	"squirrel-srv/pkg/logger"
	"time"
)

type grpcError struct {
	Message string `json:"message,omitempty"`
	Code int `json:"code,omitempty"`
}

type errorBody struct {
	Err grpcError `json:"error,omitempty"`
}

// RunServer runs HTTP/REST gateway
func RunServer(ctx context.Context, grpcPort, httpPort string, creds credentials.TransportCredentials) error {
	runtime.HTTPError = customHTTPError

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	var opts []grpc.DialOption
	if creds != nil {
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	if err := v1.RegisterServiceHandlerFromEndpoint(ctx, mux, "127.0.0.1:"+grpcPort, opts); err != nil {
		logger.Log.Fatal("failed to start HTTP gateway", zap.String("reason", err.Error()))
	}

	srv := &http.Server{
		Addr: ":" + httpPort,
		// Add handler with middlware
		Handler: middleware.AddRequestID(
			middleware.AddLogger(logger.Log, mux)),
	}

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			// sig is a ^C, handle it
			logger.Log.Warn("stop HTTP/REST gateway...")
		}

		_, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		_ = srv.Shutdown(ctx)
	}()

	logger.Log.Info("starting HTTP/REST gateway...")
	return srv.ListenAndServe()
}

func customHTTPError(_ context.Context, _ *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, _ *http.Request, err error) {
	const fallback = `{"error": "failed to marshal error message"}`

	w.Header().Set("Content-type", marshaler.ContentType())
	w.WriteHeader(runtime.HTTPStatusFromCode(grpc.Code(err)))
	jErr := json.NewEncoder(w).Encode(errorBody{
		grpcError{
			grpc.ErrorDesc(err),
			runtime.HTTPStatusFromCode(grpc.Code(err)),
		},
	})

	if jErr != nil {
		_, _ = w.Write([]byte(fallback))
	}
}
