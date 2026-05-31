package grpcserver

import (
	"context"
	"net"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const realIPMetadataKey = "x-real-ip"

// TrustedSubnetInterceptor restricts requests to the configured trusted subnet.
func TrustedSubnetInterceptor(subnet *net.IPNet) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if subnet == nil {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "missing metadata")
		}

		values := md.Get(realIPMetadataKey)
		if len(values) == 0 {
			return nil, status.Error(codes.PermissionDenied, "missing x-real-ip metadata")
		}

		ip := net.ParseIP(strings.TrimSpace(values[0]))
		if ip == nil || !subnet.Contains(ip) {
			return nil, status.Error(codes.PermissionDenied, "agent ip is not trusted")
		}

		return handler(ctx, req)
	}
}
