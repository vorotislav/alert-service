package headers

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func NewServerCheckIP(cidr string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		if cidr == "" {
			return handler(ctx, req)
		}

		if md, ok := metadata.FromIncomingContext(ctx); ok {
			values := md.Get("X-Real-IP")
			if len(values) > 0 {
				ip := values[0]
				contains, err := containsIP(ip, cidr)

				if err != nil {
					return nil, status.Errorf(codes.Internal, err.Error())
				}

				if !contains {
					return nil, status.Errorf(codes.Unavailable, "sender ip is not allowed")
				}
			} else {
				return nil, status.Errorf(codes.Internal, "header X-Real-IP is not found")
			}
		}

		return handler(ctx, req)
	}
}

func containsIP(ip, cidr string) (bool, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false, fmt.Errorf("cannot parse cidr: %w", err)
	}

	if ipNet.Contains(net.ParseIP(ip)) {
		return true, nil
	}

	return false, nil
}
