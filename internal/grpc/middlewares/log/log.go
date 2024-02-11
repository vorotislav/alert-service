package log

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func NewServerLoggerInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		startedAt := time.Now()

		resp, err := handler(ctx, req)

		fields := []zap.Field{
			zap.String("code", status.Code(err).String()),
			zap.Duration("duration", time.Since(startedAt)),
		}

		if err != nil {
			fields = append(fields, zap.Error(err))

			logger.Error("call to grpc server handled with an error", fields...)
		} else {
			logger.Info("call to grpc server handled successfully", fields...)
		}

		return resp, err
	}
}

func NewClientLoggerInterceptor(logger *zap.Logger) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		callOpts ...grpc.CallOption,
	) error {
		startedAt := time.Now()

		err := invoker(ctx, method, req, reply, cc, callOpts...)

		fields := []zap.Field{
			zap.String("code", status.Code(err).String()),
			zap.String("method_client", method),
			zap.Duration("duration", time.Since(startedAt)),
		}

		if err != nil {
			fields = append(fields, zap.Error(err))

			logger.Error("grpc client call return an error", fields...)
		} else {
			logger.Info("grpc client call was successful", fields...)
		}

		return err
	}
}
