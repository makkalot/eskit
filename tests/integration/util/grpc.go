package util

import (
	"fmt"
	"github.com/makkalot/eskit/lib/common"
	. "github.com/onsi/ginkgo/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"reflect"

	"github.com/davecgh/go-spew/spew"
)

func IsGrpcCodeError(err error, code codes.Code) bool {
	current := status.Code(err)
	return current == code
}

func AssertGrpcCodeErr(err error, code codes.Code) error {
	if current := status.Code(err); current != code {
		return fmt.Errorf("expected : %s:%d status code got %s:%d (%s)", code.String(), code, current.String(), current, err.Error())
	}
	return nil
}

func AssertGrpcCode(err error, code codes.Code) {
	if current := status.Code(err); current != code {
		Fail(AssertGrpcCodeErr(err, code).Error())
	}
}

func AssertContainsEventLogEntry(event *common.Event, logEntries []*common.AppLogEntry) {
	for _, entry := range logEntries {
		if reflect.DeepEqual(entry.Event, event) {
			return
		}
	}

	Fail(fmt.Sprintf("event : %s is not in %s", spew.Sdump(event), spew.Sdump(logEntries)))
}

func AssertContainsEvent(event *common.Event, events []*common.Event) {
	for _, e := range events {
		if reflect.DeepEqual(e, event) {
			return
		}
	}

	Fail(fmt.Sprintf("event : %s is not in %s", spew.Sdump(event), spew.Sdump(events)))
}
