// Package errchecks provides a unified way to check for errors in your unit test.
// It must not be used in your production code.
// For an example, on how to use it in unit tests please check Test_ErrChecks in errchecks_test.go
package errchecks

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
)

type Check = func(error) string

// Assert verifies that the passed check matches the given the error.
// If check == nil we handle it as IsNil.
// Reason: If you do not pass a check in your test, we assume that you have no error at all.
func Assert(t *testing.T, value error, check Check) {
	t.Helper()
	if check == nil {
		check = IsNil
	}
	msg := check(value)
	if msg != "" {
		t.Fatal(msg)
	}
}

func IsNil(err error) string {
	if err != nil {
		return fmt.Sprintf("error is not nil: %v", err)
	}
	return ""
}

// Use for some specific/constant errors only (see examples in test). Otherwise only compare the fields that are relevant for the api.
func Is(expectedErr error) Check {
	return func(err error) string {
		if errors.Is(err, expectedErr) {
			return ""
		}
		return fmt.Sprintf("unexpected error: want=%v got=%v", expectedErr, err)
	}
}

func HasStatusCode(code codes.Code) Check {
	return func(err error) string {
		st, _ := grpcStatus.FromError(err)
		if st.Code() == code {
			return ""
		}
		return fmt.Sprintf("unexpected error code: want=%s got=%s, err=%v", code.String(), st.Code().String(), err)
	}
}

func HasReason(reason string) Check {
	return func(err error) string {
		details := grpcStatus.Convert(err).Details()
		var errInfo *errdetails.ErrorInfo
		for _, detail := range details {
			if errDetail, ok := detail.(*errdetails.ErrorInfo); ok {
				errInfo = errDetail
				break
			}
		}
		if errInfo == nil {
			return fmt.Sprintf("reason expected but not found: want=%s", reason)
		}
		return cmp.Diff(reason, errInfo.Reason, protocmp.Transform())
	}
}

func IsFailedPrecondition(violations []*errdetails.PreconditionFailure_Violation) Check {
	return func(err error) string {
		if msg := HasStatusCode(codes.FailedPrecondition)(err); msg != "" {
			return msg
		}
		details := grpcStatus.Convert(err).Details()
		var foundViolations []*errdetails.PreconditionFailure_Violation
		for _, detail := range details {
			if preconditionFailed, ok := detail.(*errdetails.PreconditionFailure); ok {
				foundViolations = preconditionFailed.Violations
				break
			}
		}
		if len(foundViolations) == 0 && len(violations) == 0 {
			// ignore the difference between nil & []
			return ""
		}
		return cmp.Diff(violations, foundViolations, protocmp.Transform())
	}
}

// IsInvalidArgument verify a specific error code (codes.InvalidArgument) and that all violations exactly match.
// In order to make the validations as robust as possible please do not verify description by a static string but by
// comparing the result of the validation function only. Although the description in errdetails.BadRequest_FieldViolation is part
// of the API, we might want to migrate to a translatable description in the future.
// GOOD (check for the invalidity but not the message):
//
//	errchecks.IsInvalidArgument([]*errdetails.BadRequest_FieldViolation{
//		validations.ValidateNotNil("delivery_date", (*pbDate.Date)(nil)),
//	}),
//
// BAD (check vor a very specific message, which can't be changed anymore):
//
//	errchecks.IsInvalidArgument([]*errdetails.BadRequest_FieldViolation{
//		{Field: "delivery_date", Description: "unexpected nil value"},
//	}),
func IsInvalidArgument(violations []*errdetails.BadRequest_FieldViolation) Check {
	return func(err error) string {
		if msg := HasStatusCode(codes.InvalidArgument)(err); msg != "" {
			return msg
		}
		details := grpcStatus.Convert(err).Details()
		var foundViolations []*errdetails.BadRequest_FieldViolation
		for _, detail := range details {
			if badRequest, ok := detail.(*errdetails.BadRequest); ok {
				foundViolations = badRequest.FieldViolations
				break
			}
		}
		if len(foundViolations) == 0 && len(violations) == 0 {
			// ignore the difference between nil & []
			return ""
		}
		return cmp.Diff(violations, foundViolations, protocmp.Transform())
	}
}

// HasMsgPrefix is the preferred check if you want to ensure that a specific code path is covered.
// in the example
//
//	 err := pkgFn()
//	 if err != nil {
//			// wrap the error to verify in our test that we actually test this path
//			return nil, fmt.Errorf("a : %w", err)
//		}
//
// you can use HasMsgPrefix to verify that the error case covers an issue in pkgFn
func HasMsgPrefix(msg string) Check {
	return func(err error) string {
		if err == nil {
			return fmt.Sprintf("expected an error but got no error: wantMsg=%s", msg)
		}
		if strings.HasPrefix(err.Error(), msg) {
			return ""
		}
		return fmt.Sprintf("unexpected error msg: wantMsg=%s got=%v", msg, err)
	}
}

// MsgContains is the preferred check if you want to ensure that a specific error is the root cause
// However, often the root cause is within a pkg or external lib. In this case never prefer HasMsgPrefix.
// to not make any assumption on the pkg/lib.
func MsgContains(msg string) Check {
	return func(err error) string {
		if err == nil {
			return fmt.Sprintf("expected an error but got no error: wantMsg=%s", msg)
		}
		if strings.Contains(err.Error(), msg) {
			return ""
		}
		return fmt.Sprintf("unexpected error msg: wantMsg=%s got=%v", msg, err)
	}
}

// HasErrorInfo is the preferred check if you want to ensure that the specific ErrorInfo is matched
func HasErrorInfo(errorInfo *errdetails.ErrorInfo) Check {
	return func(err error) string {
		if err == nil {
			return "expected an error but got no error"
		}
		details := grpcStatus.Convert(err).Details()
		var foundErrInfo *errdetails.ErrorInfo
		for _, detail := range details {
			if errInfoInError, ok := detail.(*errdetails.ErrorInfo); ok {
				foundErrInfo = errInfoInError
				break
			}
		}
		if foundErrInfo == nil && errorInfo == nil {
			// ignore the difference between nil & nil
			return ""
		}
		return cmp.Diff(errorInfo, foundErrInfo, protocmp.Transform())
	}
}

func IsAlertable(val bool) Check {
	return func(err error) string {
		typedErr, ok := err.(interface {
			IsAlertable() bool
		})
		for !ok && err != nil {
			err = errors.Unwrap(err)
			typedErr, ok = err.(interface {
				IsAlertable() bool
			})
		}
		if !ok {
			return "interface IsAlertable is not implemented"
		}
		if typedErr.IsAlertable() != val {
			return "implements IsAlertable but returns wrong value"
		}
		return ""
	}
}

func IsRetryable(val bool) Check {
	return func(err error) string {
		typedErr, ok := err.(interface {
			IsRetryable() bool
		})
		for !ok && err != nil {
			err = errors.Unwrap(err)
			typedErr, ok = err.(interface {
				IsRetryable() bool
			})
		}
		if !ok {
			return "interface IsRetryable is not implemented"
		}
		if typedErr.IsRetryable() != val {
			return "implements IsRetryable but returns wrong value"
		}
		return ""
	}
}

// All combines many checks into a single one
// If any of these checks is nil, we ignore the check.
// If all of these checks are nil (or no checks are passed) we handle it as IsNil
func All(checks ...Check) Check {
	return func(err error) string {
		nonNilChecks := make([]Check, 0, len(checks))
		for _, check := range checks {
			if check != nil {
				nonNilChecks = append(nonNilChecks, check)
			}
		}
		if len(nonNilChecks) == 0 {
			return IsNil(err)
		}
		msgs := make([]string, 0, len(checks))
		for _, check := range nonNilChecks {
			msg := check(err)
			if msg != "" {
				msgs = append(msgs, msg)
			}
		}
		if len(msgs) != 0 {
			return fmt.Sprintf("failed test validations: [%s]", strings.Join(msgs, ", "))
		}
		return ""
	}
}
