package eventstore

import (
	"fmt"
	"github.com/makkalot/eskit/generated/grpc/go/common"
	store "github.com/makkalot/eskit/generated/grpc/go/eventstore"
	eskitcommon "github.com/makkalot/eskit/services/lib/common"
	"strconv"
)

type InMemoryStore struct {
	eventStore map[string][]*store.Event
	logs       []*store.AppLogEntry
}

func NewInMemoryStore() StoreWithCleanup {
	return &InMemoryStore{
		eventStore: map[string][]*store.Event{},
		logs:       []*store.AppLogEntry{},
	}
}

// Cleanup resets the db
func (s *InMemoryStore) Cleanup() error {
	s.eventStore = map[string][]*store.Event{}
	s.logs = []*store.AppLogEntry{}
	return nil
}

func (s *InMemoryStore) Append(event *store.Event) error {
	if s.eventStore[event.Originator.Id] == nil {
		s.eventStore[event.Originator.Id] = []*store.Event{event}
		return s.appendLog(event)
	}

	events := s.eventStore[event.Originator.Id]
	latestEvent := events[len(events)-1]
	latestVersion := latestEvent.Originator.Version
	latestVersionInt, err := strconv.ParseInt(latestVersion, 10, 64)
	if err != nil {
		return err
	}

	newVersion := event.Originator.Version
	newVersionInt, err := strconv.ParseInt(newVersion, 10, 64)
	if err != nil {
		return err
	}

	if newVersionInt <= latestVersionInt {
		//log.Println("current store is like : ", spew.Sdump(s.eventStore))
		return fmt.Errorf("you apply version : %d, db version is : %d for %s: %w", newVersionInt, latestVersionInt, event.Originator.Id, ErrDuplicate)
	}

	s.eventStore[event.Originator.Id] = append(s.eventStore[event.Originator.Id], event)

	return s.appendLog(event)
}

func (s *InMemoryStore) appendLog(event *store.Event) error {
	var latestID string
	if len(s.logs) == 0 {
		latestID = "1"
	} else {
		latestLog := s.logs[len(s.logs)-1]
		latestIDInt, err := strconv.ParseInt(latestLog.Id, 10, 64)
		if err != nil {
			return err
		}
		latestIDInt++
		latestID = strconv.Itoa(int(latestIDInt))
	}

	s.logs = append(s.logs, &store.AppLogEntry{
		Id:    latestID,
		Event: event,
	})

	return nil
}

func (s *InMemoryStore) Get(originator *common.Originator, fromVersion bool) ([]*store.Event, error) {
	//log.Println("event store : ", spew.Sdump(s.eventStore))

	events := s.eventStore[originator.Id]
	if events == nil || len(events) == 0 {
		return []*store.Event{}, nil
	}

	eventVersion := originator.Version
	if eventVersion == "" {
		return events, nil
	}

	eventVersionInt, err := strconv.ParseInt(eventVersion, 10, 64)
	if err != nil {
		return nil, err
	}

	var results []*store.Event
	for _, e := range events {
		currentOriginator := e.Originator
		currentVersion := currentOriginator.Version
		currentVersionInt, err := strconv.ParseInt(currentVersion, 10, 64)
		if err != nil {
			return nil, err
		}

		if !fromVersion {
			if currentVersionInt <= eventVersionInt {
				results = append(results, e)
			}
		} else {
			if currentVersionInt >= eventVersionInt {
				results = append(results, e)
			}
		}
	}

	return results, nil
}

func (s *InMemoryStore) Logs(fromID uint64, size uint32, pipelineID string) ([]*store.AppLogEntry, error) {
	if s.logs == nil || len(s.logs) == 0 {
		return []*store.AppLogEntry{}, nil
	}

	if len(s.logs)-int(fromID) < 0 {
		return []*store.AppLogEntry{}, nil
	}

	//log.Println("fetching : ")
	if fromID > 0 {
		fromID--
	}

	var results []*store.AppLogEntry
	//log.Println("logs : ", spew.Sdump(s.logs))
	if len(s.logs)-int(fromID) > int(size) {
		//log.Printf("fetching : %d:%d\n", int(fromID), int(fromID)+int(size))
		results = s.logs[fromID : int(fromID)+int(size)]
	} else {
		results = s.logs[fromID:]
	}

	if pipelineID == "" {
		return results, nil
	}

	var finalResults []*store.AppLogEntry
	for _, r := range results {
		if eskitcommon.ExtractEntityType(r.Event) == pipelineID {
			finalResults = append(finalResults, r)
		}
	}

	return finalResults, nil
}
