package auth

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

// GRPCUnaryInterceptor extracts user's information from the client certificates and put it in context.
func GRPCUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	user, err := extractUserDetails(ctx)
	if err != nil {
		return nil, err
	}
	ctx = NewContext(ctx, user)
	return handler(ctx, req)
}

// GRPCStreamInterceptor extracts user's information from the client certificates and put it in context.
func GRPCStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	user, err := extractUserDetails(ss.Context())
	if err != nil {
		return err
	}

	return handler(srv, &contextAwareStream{
		ServerStream: ss,
		ctx:          NewContext(ss.Context(), user),
	})
}

func extractUserDetails(ctx context.Context) (*User, error) {
	pInfo, ok := peer.FromContext(ctx)
	if !ok || pInfo == nil || pInfo.AuthInfo == nil {
		return nil, NewGRPCMissingCertError()
	}

	tlsInfo, ok := pInfo.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return nil, NewGRPCMissingCertError()
	}

	user := userFromCert(tlsInfo)
	if err := user.Validate(); err != nil {
		return nil, NewGRPCInvalidCertError(err)
	}

	return user, nil
}

func userFromCert(tlsInfo credentials.TLSInfo) *User {
	for _, chains := range tlsInfo.State.VerifiedChains {
		for _, chain := range chains {
			return NewUser(chain.Subject.CommonName, chain.Subject.Organization)
		}
	}
	return nil
}

// contextAwareStream wraps around the embedded grpc.ServerStream, and returns given ctx.
type contextAwareStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *contextAwareStream) Context() context.Context {
	return w.ctx
}
