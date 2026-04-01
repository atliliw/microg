package serverinterceptors

import (
	"context"
	"github.com/atliliw/microg/pkg/errors"
	"strconv"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UnaryErrorInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	resp, err := handler(ctx, req)
	if err == nil {
		return resp, nil
	}

	coder := errors.ParseCoder(err)
	if coder == nil {
		return nil, err
	}

	st := status.New(toGRPCCode(coder.HTTPStatus()), coder.String())

	errorInfo := &errdetails.ErrorInfo{
		Reason: coder.String(),
		Domain: "microg",
		Metadata: map[string]string{
			"code":      strconv.Itoa(coder.Code()),
			"http_code": strconv.Itoa(coder.HTTPStatus()),
			"reference": coder.Reference(),
		},
	}

	stWithDetails, derr := st.WithDetails(errorInfo)
	if derr != nil {
		return nil, st.Err()
	}

	return nil, stWithDetails.Err()
}

func toGRPCCode(httpCode int) codes.Code {
	switch {
	case httpCode >= 200 && httpCode < 300:
		return codes.OK
	case httpCode == 400:
		return codes.InvalidArgument
	case httpCode == 401:
		return codes.Unauthenticated
	case httpCode == 403:
		return codes.PermissionDenied
	case httpCode == 404:
		return codes.NotFound
	case httpCode == 409:
		return codes.AlreadyExists
	case httpCode == 429:
		return codes.ResourceExhausted
	case httpCode >= 500:
		return codes.Internal
	default:
		return codes.Unknown
	}
}
