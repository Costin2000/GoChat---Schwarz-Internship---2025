package errchecks

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"
)

// Note: This cannot be an runnable example, because it is intended for tests only.
func Test_ErrChecks(t *testing.T) {
	createErrorFromReason := func(reason string) error {
		st := grpcStatus.New(codes.InvalidArgument, "invalid filter values")
		st, err := st.WithDetails(&errdetails.ErrorInfo{Reason: reason})
		if err != nil {
			return err
		}
		return st.Err()
	}
	createErrorInfoFromReasonAndMetaData := func(reason string, metaData map[string]string) error {
		st := grpcStatus.New(codes.Internal, "invalid filter values")
		st, err := st.WithDetails(&errdetails.ErrorInfo{Metadata: metaData, Reason: reason})
		if err != nil {
			return err
		}
		return st.Err()
	}
	tests := []struct {
		name            string
		givenRequest    *Request
		givenCheckErr   error
		givenFetchError error
		wantError       Check
	}{
		{
			name:         "state filter = 0",
			givenRequest: &Request{FilterStatus: 0, FilterDescription: "filter"},
			wantError: IsInvalidArgument([]*errdetails.BadRequest_FieldViolation{
				// Note: we prefer to check the equality by calling the validation function again.
				// However, we do not want to have this dependency here, so we compare strings.
				{Field: "filter_status", Description: "filter status must be greater than zero"},
			}),
		},
		{
			name:         "description filter is empty",
			givenRequest: &Request{FilterStatus: 2, FilterDescription: ""},
			wantError: IsInvalidArgument([]*errdetails.BadRequest_FieldViolation{
				// Note: we prefer to check the equality by calling the validation function again.
				// However, we do not want to have this dependency here, so we compare strings.
				{Field: "filter_description", Description: "filter description must be non-empty"},
			}),
		},
		{
			name:          "precondition check failed",
			givenRequest:  &Request{FilterStatus: 2, FilterDescription: "filter"},
			givenCheckErr: fmt.Errorf("something is wrong"),
			wantError: IsFailedPrecondition([]*errdetails.PreconditionFailure_Violation{
				{Type: "ready", Subject: "my_service", Description: "service not ready"},
			}),
		},
		{
			name:            "fetching failed",
			givenRequest:    &Request{FilterStatus: 2, FilterDescription: "filter"},
			givenFetchError: fmt.Errorf("some random issue"),
			wantError: All(
				HasMsgPrefix("fetching failed"),  // prefix is the preferred check if you want to ensure that a specific code path is covered
				MsgContains("some random issue"), // contains is the preferred check if you want to ensure that a specific error is included
				HasStatusCode(codes.Unknown),
			),
		},
		{
			name:            "fetching failed with specific reason",
			givenRequest:    &Request{FilterStatus: 2, FilterDescription: "filter"},
			givenFetchError: createErrorFromReason("a very specific reason"),
			wantError:       HasReason("a very specific reason"),
		},
		{
			name:         "fetching failed with specific errorInfo",
			givenRequest: &Request{FilterStatus: 2, FilterDescription: "filter"},
			givenFetchError: createErrorInfoFromReasonAndMetaData("failed", map[string]string{
				"filterA": "failed with this reason",
				"filterB": "failed with this reason",
			}),
			wantError: HasErrorInfo(&errdetails.ErrorInfo{
				Reason: "failed",
				Metadata: map[string]string{
					"filterA": "failed with this reason",
					"filterB": "failed with this reason",
				},
			}),
		},
		{
			name:            "fetching failed with context cancelled",
			givenRequest:    &Request{FilterStatus: 2, FilterDescription: "filter"},
			givenFetchError: context.Canceled,
			wantError:       Is(context.Canceled),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &server{
				client: client{
					fetch: func(ctx context.Context, filterStatus int64, filterDescription string) (*Response, error) {
						return nil, tt.givenFetchError
					},
					check: func(ctx context.Context) error {
						return tt.givenCheckErr
					},
				},
			}
			// we only test the error here, because this is what the package is about.
			_, err := srv.handler(context.Background(), tt.givenRequest)
			Assert(t, err, tt.wantError)
		})
	}
}

type Request struct {
	FilterStatus      int64
	FilterDescription string
}

type Response struct {
	Status int64
	Text   string
}

type client struct {
	check func(ctx context.Context) error
	fetch func(ctx context.Context, filterStatus int64, filterDescription string) (*Response, error)
}

type server struct {
	client client
}

// This is a small grpc handler function we want to test
// Note that this function should use package functionality from pkg/status & pkg/validations
// However, to reduce dependencies we do not do this here.
func (s *server) handler(ctx context.Context, req *Request) (*Response, error) {
	// If this check fails we see this as some precondition not met.
	if s.client.check(ctx) != nil {
		// Normally you would create the error details using pkg/status
		st := grpcStatus.New(codes.FailedPrecondition, "service not ready")
		st, err := st.WithDetails(&errdetails.PreconditionFailure{Violations: []*errdetails.PreconditionFailure_Violation{
			{Type: "ready", Subject: "my_service", Description: "service not ready"},
		}})
		if err != nil {
			return nil, err
		}
		return nil, st.Err()
	}

	// Validate both input fields. Normally you would do this using the pkg/validation
	var violations []*errdetails.BadRequest_FieldViolation
	if req.FilterStatus <= 0 {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{Field: "filter_status", Description: "filter status must be greater than zero"})
	}
	if req.FilterDescription == "" {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{Field: "filter_description", Description: "filter description must be non-empty"})
	}
	if len(violations) > 0 {
		// Normally you would create the error details using pkg/status
		st := grpcStatus.New(codes.InvalidArgument, "invalid filter values")
		st, err := st.WithDetails(&errdetails.BadRequest{FieldViolations: violations})
		if err != nil {
			return nil, err
		}
		return nil, st.Err()
	}

	resp, err := s.client.fetch(ctx, req.FilterStatus, req.FilterDescription)
	if err != nil {
		if errors.Is(err, context.Canceled) { // some specific error handling, others: sql.ErrNoRows, io.EOF
			return nil, err
		}
		// wrap the error to verify in our test that we actually test this path
		return nil, fmt.Errorf("fetching failed: %w", err)
	}
	return resp, nil
}
