package daemon_test

import (
	"context"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mszostok/job-runner/internal/auth"
	"github.com/mszostok/job-runner/internal/daemon"
	"github.com/mszostok/job-runner/internal/daemon/automock"
	"github.com/mszostok/job-runner/pkg/api/grpc"
	"github.com/mszostok/job-runner/pkg/job"
	"github.com/mszostok/job-runner/pkg/job/repo"
)

func TestHandler_Run_Success(t *testing.T) {
	// given
	serviceMock := &automock.JobService{}
	fetcherMock := &automock.TenantGetter{}
	handler := daemon.NewHandler(serviceMock, fetcherMock)

	user := auth.User{
		Name:  "Ricky",
		Roles: map[string]struct{}{"user": {}},
	}

	req := grpc.RunRequest{
		Name:    "test-name",
		Command: "sh",
		Args:    []string{"-c", "echo $MOTTO"},
		Env:     []string{"MOTTO=hakuna_matata"},
	}

	ctx := auth.NewContext(context.Background(), &user)

	serviceMock.EXPECT().Run(ctx, job.RunInput{
		Tenant:  user.Name,
		Name:    req.Name,
		Command: req.Command,
		Args:    req.Args,
		Env:     req.Env,
	}).Return(&job.RunOutput{}, nil).Once()

	// when
	out, err := handler.Run(ctx, &req)

	// then
	require.NoError(t, err)
	assert.NotNil(t, out)

	serviceMock.AssertExpectations(t)
	fetcherMock.AssertExpectations(t)
}

func TestHandler_Run_Failures(t *testing.T) {
	// globally given
	user := func() *auth.User {
		return &auth.User{
			Name:  "Ricky",
			Roles: map[string]struct{}{"user": {}},
		}
	}

	tests := []struct {
		name         string
		ctx          context.Context
		serviceError error

		expCode codes.Code
	}{
		{
			name:    "Should return unauthenticated error",
			ctx:     context.Background(),
			expCode: codes.Unauthenticated,
		},
		{
			name:         "Should return conflict error",
			ctx:          auth.NewContext(context.Background(), user()),
			serviceError: repo.NewIDCConflictError("test"),
			expCode:      codes.AlreadyExists,
		},
		{
			name:         "Should return internal error",
			ctx:          auth.NewContext(context.Background(), user()),
			serviceError: errors.New("internal error"),
			expCode:      codes.Internal,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			// given
			serviceMock := &automock.JobService{}
			fetcherMock := &automock.TenantGetter{}
			handler := daemon.NewHandler(serviceMock, fetcherMock)

			req := grpc.RunRequest{
				Name:    "test-name",
				Command: "sh",
			}

			serviceMock.EXPECT().
				Run(test.ctx, mock.AnythingOfType("job.RunInput")).
				Return(nil, test.serviceError).
				Maybe()

			// when
			out, err := handler.Run(test.ctx, &req)

			// then
			require.Error(t, err)
			gRPCErr := status.Convert(err)
			assert.Equal(t, test.expCode, gRPCErr.Code())

			assert.Nil(t, out)

			serviceMock.AssertExpectations(t)
			fetcherMock.AssertExpectations(t)
		})
	}
}

// TODO(simplification): test rest handlers
