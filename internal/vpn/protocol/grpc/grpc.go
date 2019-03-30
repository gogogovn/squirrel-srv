package grpc

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"hub.ahiho.com/ahiho/squirrel-srv/pkg/api/v1"
	"hub.ahiho.com/ahiho/squirrel-srv/pkg/auth"
	"hub.ahiho.com/ahiho/squirrel-srv/pkg/logger"
	"net"
	"os"
	"os/signal"
)

// codeToLevel redirects OK to DEBUG level logging instead of INFO
// This is example how you can log several gRPC code results
func codeToLevel(code codes.Code) zapcore.Level {
	if code == codes.OK {
		// It is DEBUG
		return zap.DebugLevel
	}
	return grpc_zap.DefaultCodeToLevel(code)
}

// RunServer runs gRPC service to publish User service
func RunServer(ctx context.Context, v1API v1.ServiceServer, port string, creds credentials.TransportCredentials) error {
	listen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	// gRPC server statup options
	var opts []grpc.ServerOption

	if creds != nil {
		opts = append(opts, grpc.Creds(creds))
	}

	// Shared options for the logger, with a custom gRPC code to log level function.
	o := []grpc_zap.Option{
		grpc_zap.WithLevels(codeToLevel),
	}
	// Make sure that log statements internal to gRPC library are logged using the zapLogger as well.
	grpc_zap.ReplaceGrpcLogger(logger.Log)

	// Add unary interceptor
	opts = append(opts, grpc_middleware.WithUnaryServerChain(
		grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_zap.UnaryServerInterceptor(logger.Log, o...),
		grpc_auth.UnaryServerInterceptor(auth.VerifyClientKey),
	))

	// Add stream interceptor (added as an example here)
	opts = append(opts, grpc_middleware.WithStreamServerChain(
		grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_zap.StreamServerInterceptor(logger.Log, o...),
		grpc_auth.StreamServerInterceptor(auth.VerifyClientKey),
	))

	// register service
	server := grpc.NewServer(opts...)
	v1.RegisterServiceServer(server, v1API)

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			// sig is a ^C, handle it
			logger.Log.Warn("shutting down gRPC server...")

			server.GracefulStop()

			<-ctx.Done()
		}
	}()

	// start gRPC server
	logger.Log.Info("starting gRPC server...")
	return server.Serve(listen)
}
