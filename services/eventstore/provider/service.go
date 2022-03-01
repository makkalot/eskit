package provider

import (
	"context"
	"errors"
	"github.com/makkalot/eskit/generated/grpc/go/common"
	store "github.com/makkalot/eskit/generated/grpc/go/eventstore"
	eskitcommon "github.com/makkalot/eskit/services/lib/common"
	"github.com/makkalot/eskit/services/lib/eventstore"
	"github.com/satori/go.uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"strconv"
	"time"
)

type EventStoreProvider struct {
	estore eventstore.Store
}

func NewEventStoreApiProvider(estore eventstore.Store) (store.EventstoreServiceServer, error) {
	return &EventStoreProvider{
		estore: estore,
	}, nil
}

func (svc *EventStoreProvider) Healtz(ctx context.Context, request *store.HealthRequest) (*store.HealthResponse, error) {
	_, err := svc.estore.Get(&common.Originator{Id: uuid.Must(uuid.NewV4()).String()}, false)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "connecting to store failed : %v", err)
	}
	return &store.HealthResponse{
		Message: "OK",
	}, nil
}

func (svc *EventStoreProvider) Append(ctx context.Context, request *store.AppendEventRequest) (*store.AppendEventResponse, error) {
	if request.Event == nil {
		return nil, status.Errorf(codes.InvalidArgument, "missing event")
	}

	event := request.Event
	if event.GetOriginator() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "missing originator")
	}

	if event.GetEventType() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "missing event type")
	}

	if err := svc.estore.Append(event); err != nil {
		log.Println("append failed : ", err)
		if errors.Is(err, eventstore.ErrDuplicate) {
			return nil, status.Errorf(codes.AlreadyExists, "append : %v", err)
		}
		return nil, status.Errorf(codes.Internal, "append : %v", err)
	}

	return &store.AppendEventResponse{}, nil
}

func (svc *EventStoreProvider) GetEvents(ctx context.Context, request *store.GetEventsRequest) (*store.GetEventsResponse, error) {
	if request.GetOriginator() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "missing originator")
	}

	events, err := svc.estore.Get(request.Originator, false)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fetch : %v", err)
	}

	return &store.GetEventsResponse{Events: events}, nil
}

func (svc *EventStoreProvider) Logs(ctx context.Context, request *store.AppLogRequest) (*store.AppLogResponse, error) {
	fromID := request.FromId
	if fromID == "" {
		fromID = "0"
	}

	fromIDInt, err := strconv.ParseUint(fromID, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid fromID : %v", err)
	}

	size := request.Size
	if size == 0 {
		size = 20
	}

	if request.Selector == "" {
		results, err := svc.estore.Logs(fromIDInt, size, request.PipelineId)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "fetch logs : %v", err)
		}

		return &store.AppLogResponse{
			Results: results,
		}, nil
	}

	var finalResults []*store.AppLogEntry
	results, err := svc.estore.Logs(fromIDInt, size, request.PipelineId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fetch logs : %v", err)
	}

	for results != nil && len(results) > 0 && len(results) < int(size) {
		for _, r := range results {
			if svc.isEntryCompliant(r.Event, request.Selector) {
				finalResults = append(finalResults, r)
			}
		}

		//log.Println("Results : ", spew.Sdump(results))
		nextID := results[len(results)-1].Id
		nextIDInt, err := strconv.ParseUint(nextID, 10, 64)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "invalid fromID : %v", err)
		}

		nextIDInt++
		results, err = svc.estore.Logs(nextIDInt, size, request.PipelineId)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "fetch logs : %v", err)
		}

	}

	return &store.AppLogResponse{Results: finalResults}, nil
}

func (svc *EventStoreProvider) LogsPoll(request *store.AppLogRequest, stream store.EventstoreService_LogsPollServer) error {
	fromID := request.FromId
	if fromID == "" {
		fromID = "0"
	}

	fromIDInt, err := strconv.ParseUint(fromID, 10, 64)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid fromID : %v", err)
	}

	size := request.Size
	if size == 0 {
		size = 20
	}

	lastIDInt := fromIDInt
	for {
		results, err := svc.estore.Logs(lastIDInt, size, request.PipelineId)
		if err != nil {
			return status.Errorf(codes.Internal, "fetch logs : %v", err)
		}

		if results == nil || len(results) == 0 {
			time.Sleep(time.Millisecond * 100)
			continue
		}

		for _, r := range results {
			if svc.isEntryCompliant(r.Event, request.Selector) {
				if err := stream.Send(r); err != nil {
					return status.Errorf(codes.Internal, "stream send : %v", err)
				}
			}

			nextID := results[len(results)-1].Id
			nextIDInt, err := strconv.ParseUint(nextID, 10, 64)
			if err != nil {
				return status.Errorf(codes.Internal, "invalid fromID : %v", err)
			}

			nextIDInt++
			lastIDInt = nextIDInt
		}
	}
}

func (svc *EventStoreProvider) isEntryCompliant(event *store.Event, selector string) bool {
	if selector == "" || selector == "*" {
		return true
	}

	selectorEntityType := eskitcommon.ExtractEntityTypeFromStr(selector)
	selectorEventType := eskitcommon.ExtractEventTypeFromStr(selector)

	entityType := eskitcommon.ExtractEntityType(event)
	eventName := eskitcommon.ExtractEventType(event)

	if selectorEntityType != "*" && selectorEntityType != entityType {
		return false
	}

	if selectorEventType != "*" && selectorEventType != eventName {
		return false
	}

	return true
}
