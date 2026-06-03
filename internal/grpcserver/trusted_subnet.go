package grpcserver

import (
	"context"
	"errors"
	"net"

	"github.com/ilyinon/go-musthave-metrics/internal/realip"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

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

		values := md.Get(realip.MetadataKey)
		if len(values) == 0 {
			return nil, status.Errorf(codes.PermissionDenied, "missing %s metadata", realip.MetadataKey)
		}

		if err := realip.CheckTrustedSubnet(subnet, values[0]); err != nil {
			if errors.Is(err, realip.ErrMissing) {
				return nil, status.Errorf(codes.PermissionDenied, "missing %s metadata", realip.MetadataKey)
			}
			return nil, status.Error(codes.PermissionDenied, "agent ip is not trusted")
		}

		return handler(ctx, req)
	}
}
