package util

import (
	. "github.com/onsi/ginkgo"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
	"fmt"
	store "github.com/makkalot/eskit/generated/grpc/go/eventstore"
	"github.com/golang/protobuf/proto"

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

func AssertContainsEventLogEntry(event *store.Event, logEntries []*store.AppLogEntry) {
	for _, entry := range logEntries {
		if proto.Equal(entry.Event, event) {
			return
		}
	}

	Fail(fmt.Sprintf("event : %s is not in %s", spew.Sdump(event), spew.Sdump(logEntries)))
}

func AssertContainsEvent(event *store.Event, events []*store.Event) {
	for _, e := range events {
		if proto.Equal(e, event) {
			return
		}
	}

	Fail(fmt.Sprintf("event : %s is not in %s", spew.Sdump(event), spew.Sdump(events)))
}
