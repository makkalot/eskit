package provider

import (
	"github.com/makkalot/eskit/generated/grpc/go/common"
	store "github.com/makkalot/eskit/generated/grpc/go/eventstore"
	"fmt"
	"time"
	"context"
	commonutil "github.com/makkalot/eskit/services/common"
	"gopkg.in/evanphx/json-patch.v3"
	"errors"
	"strings"
	"strconv"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
)

var (
	RecordNotFound  = errors.New("not found")
	RecordDeleted   = errors.New("deleted")
	RecordDuplicate = errors.New("duplicate")
)

func IsErrNotFound(err error) bool {
	return err == RecordNotFound
}

func IsErrDeleted(err error) bool {
	return err == RecordDeleted
}

func IsDuplicare(err error) bool {
	return err == RecordDuplicate
}

type CrudStore interface {
	Create(entityType string, originator *common.Originator, payload string) error
	Update(entityType string, originator *common.Originator, payload string) (*common.Originator, error)
	Get(originator *common.Originator, deleted bool) (string, *common.Originator, error)
	Delete(entityType string, originator *common.Originator) (*common.Originator, error)
	List(entityType, fromID string, size int) ([]*common.Originator, string, error)
}

type CrudStoreProvider struct {
	ctx    context.Context
	estore store.EventstoreServiceClient
}

func NewCrudStoreProvider(ctx context.Context, estore store.EventstoreServiceClient) (CrudStore, error) {
	return &CrudStoreProvider{
		ctx:    ctx,
		estore: estore,
	}, nil
}

func (crud *CrudStoreProvider) Create(entityType string, originator *common.Originator, payload string) error {
	if originator == nil {
		return fmt.Errorf("missing originator")
	}

	if originator.Version == "" {
		originator.Version = "1"
	}

	event := &store.Event{
		Originator: originator,
		EventType:  fmt.Sprintf("%s.Created", entityType),
		Payload:    payload,
		OccuredOn:  time.Now().UTC().Unix(),
	}

	//log.Printf("Appending Create Event : %s", spew.Sdump(event))
	_, err := crud.estore.Append(crud.ctx, &store.AppendEventRequest{Event: event})
	if status.Code(err) == codes.AlreadyExists {
		return RecordDuplicate
	}
	return err
}

func (crud *CrudStoreProvider) Update(entityType string, originator *common.Originator, payload string) (*common.Originator, error) {
	if originator.Version == "" {
		return nil, fmt.Errorf("misisng version")
	}

	newOriginator, err := commonutil.IncrOriginator(originator)
	if err != nil {
		return nil, err
	}

	latestObj, _, err := crud.Get(originator, false)
	if err != nil {
		return nil, err
	}

	patch, err := jsonpatch.CreateMergePatch([]byte(latestObj), []byte(payload))
	if err != nil {
		return nil, fmt.Errorf("patch creation failed : %v", err)
	}

	//log.Println("Patch : original : ", string(latestObj))
	//log.Println("Patch : payload : ", string(payload))
	//log.Println("Patch : patch : ", string(patch))

	event := &store.Event{
		Originator: newOriginator,
		EventType:  fmt.Sprintf("%s.Updated", entityType),
		Payload:    string(patch),
		OccuredOn:  time.Now().UTC().Unix(),
	}

	_, err = crud.estore.Append(crud.ctx, &store.AppendEventRequest{Event: event})
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return nil, RecordDuplicate
		}
		return nil, err
	}
	return newOriginator, nil
}

func (crud *CrudStoreProvider) Get(originator *common.Originator, deleted bool) (string, *common.Originator, error) {
	resp, err := crud.estore.GetEvents(crud.ctx, &store.GetEventsRequest{
		Originator: originator,
	})

	if err != nil {
		return "", nil, err
	}

	events := resp.GetEvents()
	if events == nil || len(events) == 0 {
		return "", nil, RecordNotFound
	}

	latestEvent := events[len(events)-1]
	if crud.isEventDeleted(latestEvent) && !deleted {
		return "", nil, RecordDeleted
	}

	currentPayload := []byte(events[0].Payload)
	currentOriginator := events[0].Originator

	originatorVersion := originator.Version
	if originatorVersion != "" {
		originatorVersionInt, err := strconv.ParseInt(originatorVersion, 10, 64)
		if err != nil {
			return "", nil, err
		}

		// the version we're looking for is not created yet
		if int(originatorVersionInt) > len(events) {
			return "", nil, RecordNotFound
		}
	}

	if len(events) == 1 {
		return string(currentPayload), currentOriginator, nil
	}

	for _, e := range events[1:] {

		// ignore non crud events
		if !crud.isCrudEvent(e) {
			continue
		}

		if crud.isEventDeleted(e) {
			continue
		}

		currentPayload, err = jsonpatch.MergePatch([]byte(currentPayload), []byte(e.Payload))
		if err != nil {
			return "", nil, fmt.Errorf("apply patch : %v", err)
		}

		currentOriginator = e.Originator
	}

	return string(currentPayload), currentOriginator, nil

}

func (crud *CrudStoreProvider) List(entityType, fromID string, size int) ([]*common.Originator, string, error) {
	if fromID == "" {
		fromID = "0"
	}

	var eventSize int
	if size == 0 {
		size = 10
	}

	eventSize = size * 20

	resp, err := crud.estore.Logs(crud.ctx, &store.AppLogRequest{
		FromId:     fromID,
		Size:       uint32(eventSize),
		PipelineId: entityType,
	})

	if err != nil {
		return nil, "", err
	}

	if len(resp.GetResults()) == 0 {
		return nil, "", nil
	}

	found := map[string]bool{}
	var preResults []*common.Originator
	var results []*common.Originator
	var lastID string

	for _, entry := range resp.Results {
		originatorID := entry.Event.Originator.Id
		if _, ok := found[originatorID]; ok {
			if crud.isEventDeleted(entry.Event) {
				delete(found, originatorID)
			}
		} else {
			if !crud.isEventDeleted(entry.Event) {
				found[originatorID] = true
				preResults = append(preResults, &common.Originator{
					Id: originatorID,
				})
			}
		}

		lastID = entry.Id
		if len(found) >= size {
			break
		}
	}

	// prepare the last result
	for _, r := range preResults {
		if _, ok := found[r.Id]; !ok {
			continue
		}
		results = append(results, r)
	}

	lastID, err = commonutil.IncrStringInt(lastID)
	if err != nil {
		return nil, "", err
	}

	return results, lastID, nil
}

func (crud *CrudStoreProvider) isEventDeleted(event *store.Event) bool {
	eventType := commonutil.ExtractEventType(event)
	return strings.ToLower(eventType) == "deleted"
}

func (crud *CrudStoreProvider) isCrudEvent(event *store.Event) bool {
	eventType := commonutil.ExtractEventType(event)
	switch strings.ToLower(eventType) {
	case "created", "updated", "deleted":
		return true
	default:
		return false
	}
}

func (crud *CrudStoreProvider) Delete(entityType string, originator *common.Originator) (*common.Originator, error) {
	_, latestOriginator, err := crud.Get(originator, false)
	if err != nil {
		return nil, err
	}

	newOriginator, err := commonutil.IncrOriginator(latestOriginator)
	if err != nil {
		return nil, err
	}

	event := &store.Event{
		Originator: newOriginator,
		EventType:  fmt.Sprintf("%s.Deleted", entityType),
		Payload:    "{}",
		OccuredOn:  time.Now().UTC().Unix(),
	}

	_, err = crud.estore.Append(crud.ctx, &store.AppendEventRequest{Event: event})
	if err != nil {
		return nil, err
	}

	return newOriginator, nil

}
