package daemon

import (
	"context"
	"io"

	"github.com/cockroachdb/errors"
	"google.golang.org/grpc/status"

	"github.com/mszostok/job-runner/internal/auth"
	"github.com/mszostok/job-runner/pkg/api/grpc"
	"github.com/mszostok/job-runner/pkg/job"
	"github.com/mszostok/job-runner/pkg/job/repo"
)

var _ grpc.JobServiceServer = &Handler{}

// JobService provides full functionality to manged Jobs.
//go:generate mockery --name=JobService --output=automock --outpkg=automock --case=underscore --with-expecter
type JobService interface {
	Run(context.Context, job.RunInput) (*job.RunOutput, error)
	Get(context.Context, job.GetInput) (*job.GetOutput, error)
	Stop(context.Context, job.StopInput) (*job.StopOutput, error)
	StreamLogs(context.Context, job.StreamLogsInput) (*job.StreamLogsOutput, error)
}

// TenantGetter provides functionality to get Job's tenant information.
//go:generate mockery --name=TenantGetter --output=automock --outpkg=automock --case=underscore --with-expecter
type TenantGetter interface {
	GetJobTenant(in repo.GetJobTenantInput) (repo.GetJobTenantOutput, error)
}

// Handler handles incoming requests to the Daemon gRPC server.
type Handler struct {
	grpc.UnimplementedJobServiceServer

	svc          JobService
	tenantGetter TenantGetter
}

// NewHandler returns new Handler.
func NewHandler(svc JobService, getter TenantGetter) *Handler {
	return &Handler{
		tenantGetter: getter,
		svc:          svc,
	}
}

func (h *Handler) Run(ctx context.Context, req *grpc.RunRequest) (*grpc.RunResponse, error) {
	if req == nil {
		return nil, NilRequestInputError
	}

	user, err := auth.FromContext(ctx)
	if err != nil {
		return nil, TranslateError(err)
	}

	_, err = h.svc.Run(ctx, job.RunInput{
		Tenant:  user.Name,
		Name:    req.Name,
		Command: req.Command,
		Args:    req.Args,
		Env:     req.Env,
	})
	if err != nil {
		return nil, TranslateError(err)
	}

	return &grpc.RunResponse{}, nil
}

func (h *Handler) Get(ctx context.Context, req *grpc.GetRequest) (*grpc.GetResponse, error) {
	if req == nil {
		return nil, NilRequestInputError
	}

	if err := h.checkAuthorized(ctx, req.Name); err != nil {
		return nil, TranslateError(err)
	}

	out, err := h.svc.Get(ctx, job.GetInput{
		Name: req.Name,
	})
	if err != nil {
		return nil, TranslateError(err)
	}

	return &grpc.GetResponse{
		CreatedBy: out.CreatedBy,
		Status:    mapToGRPCStatus(out.Status),
		ExitCode:  int32(out.ExitCode),
	}, nil
}

func (h *Handler) Stop(ctx context.Context, req *grpc.StopRequest) (*grpc.StopResponse, error) {
	if req == nil {
		return nil, NilRequestInputError
	}

	if err := h.checkAuthorized(ctx, req.Name); err != nil {
		return nil, TranslateError(err)
	}

	stop := job.StopInput{
		Name: req.Name,
	}
	if req.GracePeriod != nil {
		stop.GracePeriod = *req.GracePeriod
	}
	out, err := h.svc.Stop(ctx, stop)
	if err != nil {
		return nil, TranslateError(err)
	}

	return &grpc.StopResponse{
		Status:   mapToGRPCStatus(out.Status),
		ExitCode: int32(out.ExitCode),
	}, nil
}

func (h *Handler) StreamLogs(req *grpc.StreamLogsRequest, gstream grpc.JobService_StreamLogsServer) error {
	if req == nil {
		return NilRequestInputError
	}

	ctx, jobName := gstream.Context(), req.Name

	if err := h.checkAuthorized(ctx, jobName); err != nil {
		return TranslateError(err)
	}

	// It's up to the 'StreamLogs' method to close the returned channels as it sends the data to it.
	// We can only use 'ctx' to cancel streaming and release associated resources.
	// TODO(simplification): In the future, change the returned channels to io.ReadCloser to make more readable API.
	stream, err := h.svc.StreamLogs(ctx, job.StreamLogsInput{Name: jobName})
	if err != nil {
		return TranslateError(err)
	}

	for {
		select {
		case <-ctx.Done():
			return status.FromContextError(ctx.Err()).Err()
		case out, ok := <-stream.Output:
			if !ok {
				return nil // output closed, no more chunk logs
			}

			err := gstream.Send(&grpc.StreamLogsResponse{
				Output: out,
			})
			if err != nil {
				return TranslateError(err)
			}
		case err := <-stream.Error:
			if errors.Is(err, io.EOF) {
				// no more chunk logs
				return nil
			}
			return TranslateError(err)
		}
	}
}

func (*Handler) Ping(_ context.Context, req *grpc.PingRequest) (*grpc.PingResponse, error) {
	return &grpc.PingResponse{
		Message: req.Message,
	}, nil
}

// TODO: it will be good to handle that directly in one place (e.g. interceptor)
// but it a bit hard to get the req.Name without playing with type casting
// for all known handler. It will be better to have a global "namespaces" per
// tenant.
func (h *Handler) checkAuthorized(ctx context.Context, jobName string) error {
	user, err := auth.FromContext(ctx)
	if err != nil {
		return err
	}

	out, err := h.tenantGetter.GetJobTenant(repo.GetJobTenantInput{
		Name: jobName,
	})
	if err != nil {
		return errors.Wrap(err, "while resoling Job's tenant")
	}

	err = user.CheckAuthorized(out.Tenant)
	if err != nil {
		return err
	}

	return nil
}

func mapToGRPCStatus(in job.Status) grpc.Status {
	return grpc.Status(grpc.Status_value[string(in)]) // TODO: rethink
}
