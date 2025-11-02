package eventstore

import (
	"fmt"
	"github.com/makkalot/eskit/lib/common"
	"github.com/makkalot/eskit/lib/types"
)

type InMemoryStore struct {
	eventStore map[string][]*types.Event
	logs       []*types.AppLogEntry
}

func NewInMemoryStore() StoreWithCleanup {
	return &InMemoryStore{
		eventStore: map[string][]*types.Event{},
		logs:       []*types.AppLogEntry{},
	}
}

// Cleanup resets the db
func (s *InMemoryStore) Cleanup() error {
	s.eventStore = map[string][]*types.Event{}
	s.logs = []*types.AppLogEntry{}
	return nil
}

func (s *InMemoryStore) Append(event *types.Event) error {
	if s.eventStore[event.Originator.ID] == nil {
		s.eventStore[event.Originator.ID] = []*types.Event{event}
		return s.appendLog(event)
	}

	events := s.eventStore[event.Originator.ID]
	latestEvent := events[len(events)-1]
	latestVersion := latestEvent.Originator.Version
	newVersion := event.Originator.Version

	if newVersion <= latestVersion {
		//log.Println("current store is like : ", spew.Sdump(s.eventStore))
		return fmt.Errorf("you apply version : %d, db version is : %d for %s: %w", newVersion, latestVersion, event.Originator.ID, ErrDuplicate)
	}

	s.eventStore[event.Originator.ID] = append(s.eventStore[event.Originator.ID], event)

	return s.appendLog(event)
}

func (s *InMemoryStore) appendLog(event *types.Event) error {
	var latestID uint64
	if len(s.logs) == 0 {
		latestID = 1
	} else {
		latestLog := s.logs[len(s.logs)-1]
		latestID = latestLog.ID + 1
	}

	s.logs = append(s.logs, &types.AppLogEntry{
		ID:    latestID,
		Event: event,
	})

	return nil
}

func (s *InMemoryStore) Get(originator *types.Originator, fromVersion bool) ([]*types.Event, error) {
	//log.Println("event store : ", spew.Sdump(s.eventStore))

	events := s.eventStore[originator.ID]
	if events == nil || len(events) == 0 {
		return []*types.Event{}, nil
	}

	eventVersion := originator.Version
	if eventVersion == 0 {
		return events, nil
	}

	var results []*types.Event
	for _, e := range events {
		currentVersion := e.Originator.Version

		if !fromVersion {
			if currentVersion <= eventVersion {
				results = append(results, e)
			}
		} else {
			if currentVersion >= eventVersion {
				results = append(results, e)
			}
		}
	}

	return results, nil
}

func (s *InMemoryStore) Logs(fromID uint64, size uint32, pipelineID string) ([]*types.AppLogEntry, error) {
	if s.logs == nil || len(s.logs) == 0 {
		return []*types.AppLogEntry{}, nil
	}

	if len(s.logs)-int(fromID) < 0 {
		return []*types.AppLogEntry{}, nil
	}

	//log.Println("fetching : ")
	if fromID > 0 {
		fromID--
	}

	var results []*types.AppLogEntry
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

	var finalResults []*types.AppLogEntry
	for _, r := range results {
		if common.ExtractEntityType(r.Event) == pipelineID {
			finalResults = append(finalResults, r)
		}
	}

	return finalResults, nil
}
