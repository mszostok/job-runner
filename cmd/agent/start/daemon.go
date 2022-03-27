package start

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/mszostok/job-runner/internal/auth"
	"github.com/mszostok/job-runner/internal/daemon"
	"github.com/mszostok/job-runner/internal/shutdown"
	pb "github.com/mszostok/job-runner/pkg/api/grpc"
	"github.com/mszostok/job-runner/pkg/cgroup"
	"github.com/mszostok/job-runner/pkg/file"
	"github.com/mszostok/job-runner/pkg/job"
	"github.com/mszostok/job-runner/pkg/job/repo"
)

const (
	daemonCGroupPath = "LPR"
	caFlagName       = "client-ca-cert"
	certFlagName     = "server-cert"
	keyFlagName      = "server-key"
)

// DaemonOptions holds options for starting daemon process.
type DaemonOptions struct {
	GRPCAddr string
	TLS      TLSOptions
}

// TLSOptions holds mTLS related settings.
type TLSOptions struct {
	Client struct {
		CAFilePath string
	}
	Server struct {
		CertFilePath string
		KeyFilePath  string
	}
}

// NewDaemon returns a new cobra.Command for starting daemon process.
func NewDaemon() *cobra.Command {
	var opts DaemonOptions
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Starts a long living Agent process.",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if err := cgroup.CheckCgroupV2Enabled(); err != nil {
				return err
			}

			// setup library
			jobRepo := repo.NewInMemory()

			flog, err := file.NewLogger()
			if err != nil {
				return err
			}

			svc, err := job.NewService(jobRepo, flog)
			if err != nil {
				return err
			}

			err = cgroup.BootstrapParent(daemonCGroupPath, cgroup.MemoryController, cgroup.CPUController, cgroup.IOController, cgroup.CPUSetController)
			if err != nil {
				return err
			}

			// setup gRPC server
			var lc net.ListenConfig
			listener, err := lc.Listen(c.Context(), "tcp", opts.GRPCAddr)
			if err != nil {
				return err
			}

			tlsConfig, err := getTLSConfig(opts.TLS)
			if err != nil {
				return err
			}

			srv := grpc.NewServer(
				grpc.Creds(credentials.NewTLS(tlsConfig)),
				grpc.UnaryInterceptor(auth.GRPCUnaryInterceptor),
				grpc.StreamInterceptor(auth.GRPCStreamInterceptor),
			)
			pb.RegisterJobServiceServer(srv, daemon.NewHandler(svc, jobRepo))

			// setup shutdown
			shutdownManager := &shutdown.ParentService{}
			shutdownManager.Register(flog)
			shutdownManager.Register(svc)
			shutdownManager.Register(shutdown.Func(srv.GracefulStop))

			// setup parallel execution
			scheduleParallel, parallelCtx := errgroup.WithContext(c.Context())
			scheduleParallel.Go(func() error {
				log.Printf("Starting TCP server on %s\n", opts.GRPCAddr)
				return srv.Serve(listener)
			})
			scheduleParallel.Go(func() error {
				<-parallelCtx.Done() // it's canceled on OS signals and if function passed to 'Go' method returns a non-nil error
				log.Println("Stopping server gracefully")
				return shutdownManager.Shutdown()
			})

			return scheduleParallel.Wait()
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&opts.GRPCAddr, "grpc-addr", ":50051", "Specifies gRPC server address.")
	flags.StringVar(&opts.TLS.Client.CAFilePath, caFlagName, "", "Path on the local disk to CA certificate to verify the client's certificate.")
	flags.StringVar(&opts.TLS.Server.CertFilePath, certFlagName, "", "Path on the local disk to client certificate to use for auth to the client's requests.")
	flags.StringVar(&opts.TLS.Server.KeyFilePath, keyFlagName, "", "Path on the local disk to client private key to use for auth to the client's requests.")

	for _, name := range []string{caFlagName, certFlagName, keyFlagName} {
		_ = cmd.MarkFlagRequired(name)
		_ = cmd.MarkFlagFilename(name)
	}

	return cmd
}

func getTLSConfig(opts TLSOptions) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(opts.Server.CertFilePath, opts.Server.KeyFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "while loading key pair")
	}

	ca := x509.NewCertPool()
	caBytes, err := ioutil.ReadFile(opts.Client.CAFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "while reading CA cert")
	}
	if ok := ca.AppendCertsFromPEM(caBytes); !ok {
		return nil, errors.Wrap(err, "while paring CA")
	}

	return &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
		ClientCAs:    ca,
		MinVersion:   tls.VersionTLS12,
	}, nil
}
