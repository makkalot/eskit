package crudstore

import (
	"context"
	"errors"
	"fmt"
	"github.com/makkalot/eskit/lib/common"
	"github.com/makkalot/eskit/lib/eventstore"
	"github.com/makkalot/eskit/lib/types"
	"gopkg.in/evanphx/json-patch.v3"
	"strconv"
	"strings"
	"time"
)

var (
	RecordNotFound = errors.New("not found")
	RecordDeleted  = errors.New("deleted")
)

func IsErrNotFound(err error) bool {
	return errors.Is(err, RecordNotFound)
}

func IsErrDeleted(err error) bool {
	return errors.Is(err, RecordDeleted)
}

func IsDuplicate(err error) bool {
	return errors.Is(err, eventstore.ErrDuplicate)
}

type CrudStore interface {
	Create(entityType string, originator *types.Originator, payload string) error
	Update(entityType string, originator *types.Originator, payload string) (*types.Originator, error)
	Get(originator *types.Originator, deleted bool) (string, *types.Originator, error)
	Delete(entityType string, originator *types.Originator) (*types.Originator, error)
	List(entityType, fromID string, size int) ([]*types.Originator, string, error)
}

type CrudStoreProvider struct {
	ctx    context.Context
	estore eventstore.Store
}

func NewCrudStoreProvider(ctx context.Context, estore eventstore.Store) (CrudStore, error) {
	return &CrudStoreProvider{
		ctx:    ctx,
		estore: estore,
	}, nil
}

func (crud *CrudStoreProvider) Create(entityType string, originator *types.Originator, payload string) error {
	if originator == nil {
		return fmt.Errorf("empty originator")
	}

	if originator.Version == "" {
		originator.Version = "1"
	}

	event := &types.Event{
		Originator: originator,
		EventType:  fmt.Sprintf("%s.Created", entityType),
		Payload:    payload,
		OccurredOn: time.Now().UTC(),
	}

	//log.Printf("Appending Create Event : %s", spew.Sdump(event))
	return crud.estore.Append(event)
}

func (crud *CrudStoreProvider) Update(entityType string, originator *types.Originator, payload string) (*types.Originator, error) {
	if originator.Version == "" {
		return nil, fmt.Errorf("misisng version")
	}

	newOriginator, err := common.IncrOriginator(originator)
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

	event := &types.Event{
		Originator: newOriginator,
		EventType:  fmt.Sprintf("%s.Updated", entityType),
		Payload:    string(patch),
		OccurredOn: time.Now().UTC(),
	}

	err = crud.estore.Append(event)
	if err != nil {
		return nil, err
	}
	return newOriginator, nil
}

func (crud *CrudStoreProvider) Get(originator *types.Originator, deleted bool) (string, *types.Originator, error) {
	events, err := crud.estore.Get(originator, false)
	if err != nil {
		return "", nil, err
	}

	if events == nil || len(events) == 0 {
		return "", nil, fmt.Errorf("%w", RecordNotFound)
	}

	latestEvent := events[len(events)-1]
	if crud.isEventDeleted(latestEvent) && !deleted {
		return "", nil, fmt.Errorf("%w", RecordDeleted)
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
			return "", nil, fmt.Errorf("%w", RecordNotFound)
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

func (crud *CrudStoreProvider) List(entityType, fromID string, size int) ([]*types.Originator, string, error) {
	if fromID == "" {
		fromID = "0"
	}

	var eventSize int
	if size == 0 {
		size = 10
	}

	// list is really for small objects with fewer version or debugging purposes
	eventSize = size * 20
	fromIDInt, err := strconv.ParseUint(fromID, 10, 64)
	if err != nil {
		return nil, "", fmt.Errorf("invalid fromID : %v", err)
	}

	logs, err := crud.estore.Logs(fromIDInt,
		uint32(eventSize),
		entityType)

	if err != nil {
		return nil, "", err
	}

	if logs == nil || len(logs) == 0 {
		return nil, "", nil
	}

	found := map[string]bool{}
	var preResults []*types.Originator
	var results []*types.Originator
	var lastID string

	for _, entry := range logs {
		originatorID := entry.Event.Originator.ID
		if _, ok := found[originatorID]; ok {
			if crud.isEventDeleted(entry.Event) {
				delete(found, originatorID)
			}
		} else {
			if !crud.isEventDeleted(entry.Event) {
				found[originatorID] = true
				preResults = append(preResults, &types.Originator{
					ID: originatorID,
				})
			}
		}

		lastID = entry.ID
		if len(found) >= size {
			break
		}
	}

	// prepare the last result
	for _, r := range preResults {
		if _, ok := found[r.ID]; !ok {
			continue
		}
		results = append(results, r)
	}

	lastID, err = common.IncrStringInt(lastID)
	if err != nil {
		return nil, "", err
	}

	return results, lastID, nil
}

func (crud *CrudStoreProvider) isEventDeleted(event *types.Event) bool {
	eventType := common.ExtractEventType(event)
	return strings.ToLower(eventType) == "deleted"
}

func (crud *CrudStoreProvider) isCrudEvent(event *types.Event) bool {
	eventType := common.ExtractEventType(event)
	switch strings.ToLower(eventType) {
	case "created", "updated", "deleted":
		return true
	default:
		return false
	}
}

func (crud *CrudStoreProvider) Delete(entityType string, originator *types.Originator) (*types.Originator, error) {
	_, latestOriginator, err := crud.Get(originator, false)
	if err != nil {
		return nil, err
	}

	newOriginator, err := common.IncrOriginator(latestOriginator)
	if err != nil {
		return nil, err
	}

	event := &types.Event{
		Originator: newOriginator,
		EventType:  fmt.Sprintf("%s.Deleted", entityType),
		Payload:    "{}",
		OccurredOn: time.Now().UTC(),
	}

	err = crud.estore.Append(event)
	if err != nil {
		return nil, err
	}

	return newOriginator, nil

}
