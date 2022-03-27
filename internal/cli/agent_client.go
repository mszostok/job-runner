package cli

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"github.com/cockroachdb/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/mszostok/job-runner/internal/cli/config"
	pb "github.com/mszostok/job-runner/pkg/api/grpc"
)

// NewDefaultGRPCAgentClient returns gRPC Agent client created from default context.
func NewDefaultGRPCAgentClient() (pb.JobServiceClient, func() error, error) {
	cfg, err := config.GetAgentAuthDetails()
	if err != nil {
		return nil, nil, err
	}

	return NewGRPCAgentClient(cfg)
}

// NewGRPCAgentClient returns gRPC Agent client.
func NewGRPCAgentClient(cfg config.Agent) (pb.JobServiceClient, func() error, error) {
	cert, err := tls.LoadX509KeyPair(cfg.ClientAuth.CertFilePath, cfg.ClientAuth.KeyFilePath)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while loading client cert")
	}

	ca := x509.NewCertPool()
	caBytes, err := ioutil.ReadFile(cfg.AgentCAFilePath)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while reading server CA")
	}
	if ok := ca.AppendCertsFromPEM(caBytes); !ok {
		return nil, nil, errors.Wrap(err, "while parsing server CA")
	}

	tlsConfig := &tls.Config{

		ServerName:   "x.lpr.example.com",
		Certificates: []tls.Certificate{cert},
		RootCAs:      ca,
		MinVersion:   tls.VersionTLS12,
	}

	conn, err := grpc.Dial(cfg.ServerURL, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	if err != nil {
		return nil, nil, err
	}

	return pb.NewJobServiceClient(conn), conn.Close, nil
}
