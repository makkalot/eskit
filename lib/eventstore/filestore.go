package eventstore

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/makkalot/eskit/lib/common"
	"github.com/makkalot/eskit/lib/types"
)

type FileMemoryStore struct {
	storePath        string
	file             *os.File
	read_file        *os.File
	lastEvent        *storedFileEvent
	lastEventLine    int64
	storedLogEntries []*storedLogEntry
}

type storedFileEvent struct {
	OriginatorID      string      `json:"originator_id"`
	OriginatorVersion uint64      `json:"originator_version"`
	EventType         string      `json:"event_type"`
	Payload           interface{} `json:"payload"`
	CreatedAt         int64       `json:"created_at"`
}

type storedLogEntry struct {
	ID                     uint64
	ApplicationID          string
	PartitionID            string
	EventOriginatorId      string
	EventOriginatorVersion uint64
	CreatedAt              int64
	EventPayload           string
	// the line number of the log entry in the file, used for pagination
	// we can use this line number to seek the file for pagination, and we can also use it to track the last read position of the log entries
	LineNumber int64
}

func NewFileMemoryStore(storePath string) *FileMemoryStore {
	basePath := filepath.Dir(storePath)

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		if err := os.MkdirAll(basePath, 0755); err != nil {
			panic(fmt.Sprintf("failed to create store path: %v", err))
		}
	}

	file, err := os.OpenFile(storePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Sprintf("failed to open store file: %v", err))
	}

	read_file, err := os.OpenFile(storePath, os.O_RDONLY, 0644)
	if err != nil {
		panic(fmt.Sprintf("failed to open store file for reading: %v", err))
	}

	return &FileMemoryStore{
		storePath: storePath,
		file:      file,
		read_file: read_file,
	}
}

func (s *FileMemoryStore) Cleanup() error {
	s.file.Close()
	s.read_file.Close()

	if err := os.Remove(s.storePath); err != nil {
		return fmt.Errorf("failed to cleanup store path: %v", err)
	}
	return nil
}

func (s *FileMemoryStore) Append(event *types.Event) error {
	latestEvent := s.lastEvent
	if latestEvent == nil {
		return s.appendFileEvent(event)
	}
	latestVersion := latestEvent.OriginatorVersion
	newVersion := event.Originator.Version

	if newVersion <= latestVersion {
		//log.Println("current store is like : ", spew.Sdump(s.eventStore))
		return fmt.Errorf("you apply version : %d, db version is : %d for %s: %w", newVersion, latestVersion, event.Originator.ID, ErrDuplicate)
	}

	return s.appendFileEvent(event)
}

func (s *FileMemoryStore) Get(originator *types.Originator, fromVersion bool) ([]*types.Event, error) {
	var events []*types.Event
	var eventLines []int64

	for _, logEntry := range s.storedLogEntries {
		if logEntry.EventOriginatorId == originator.ID {
			if fromVersion {
				if logEntry.EventOriginatorVersion >= originator.Version {
					eventLines = append(eventLines, logEntry.LineNumber)
				}
			} else {
				eventLines = append(eventLines, logEntry.LineNumber)
			}
		}
	}

	if len(eventLines) == 0 {
		return events, nil
	}

	targetLines := make(map[int64]struct{}, len(eventLines))
	var firstLine, lastLine int64
	for i, lineNo := range eventLines {
		targetLines[lineNo] = struct{}{}
		if i == 0 || lineNo < firstLine {
			firstLine = lineNo
		}
		if i == 0 || lineNo > lastLine {
			lastLine = lineNo
		}
	}

	// Seek directly to the first relevant line (1-based).
	if err := s.seekToLine(int(firstLine)); err != nil {
		return nil, fmt.Errorf("failed to seek file: %w", err)
	}

	// We are already positioned at firstLine, so track line numbers accordingly.
	currentLine := firstLine - 1
	remaining := len(targetLines)
	scanner := bufio.NewScanner(s.read_file)

	for scanner.Scan() {
		currentLine++
		if currentLine > lastLine || remaining == 0 {
			break
		}

		if _, ok := targetLines[currentLine]; !ok {
			continue
		}

		var fileEvent storedFileEvent
		if err := json.Unmarshal(scanner.Bytes(), &fileEvent); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event on line %d: %w", currentLine, err)
		}

		delete(targetLines, currentLine)
		remaining--

		if fileEvent.OriginatorID != originator.ID {
			continue
		}

		if fromVersion && fileEvent.OriginatorVersion < originator.Version {
			continue
		}

		payload, _ := fileEvent.Payload.(string)

		events = append(events, &types.Event{
			Originator: &types.Originator{
				ID:      fileEvent.OriginatorID,
				Version: fileEvent.OriginatorVersion,
			},
			EventType:  fileEvent.EventType,
			Payload:    payload,
			OccurredOn: time.Unix(fileEvent.CreatedAt, 0).UTC(),
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan file: %w", err)
	}

	return events, nil
}

func (s *FileMemoryStore) Logs(fromID uint64, size uint32, pipelineID string) ([]*types.AppLogEntry, error) {
	var logs []*types.AppLogEntry

	start_index := 0
	start_line := 0

	for i, logEntry := range s.storedLogEntries {
		if logEntry.ID < fromID {
			continue
		}
		if pipelineID != "" && logEntry.PartitionID != pipelineID {
			continue
		}

		start_index = i
		start_line = int(logEntry.LineNumber)
		break
	}

	s.seekToLine(start_line)
	scanner := bufio.NewScanner(s.read_file)

	currentIndex := start_index
	for scanner.Scan() {
		line := scanner.Text()

		logEntry := s.storedLogEntries[currentIndex]
		if pipelineID != "" && logEntry.PartitionID != pipelineID {
			continue
		}

		event := &types.Event{}
		if err := json.Unmarshal([]byte(line), event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event for log entry %d: %w", logEntry.ID, err)
		}

		logs = append(logs, &types.AppLogEntry{
			ID:    logEntry.ID,
			Event: event,
		})

		currentIndex++
		if size > 0 && uint32(len(logs)) >= size {
			break
		}
	}

	return logs, nil
}

// seekToLine moves the file cursor to the start of the given 1-based line.
// It rewinds to the beginning and then consumes (line-1) newline-delimited lines.
func (s *FileMemoryStore) seekToLine(line int) error {
	if line <= 1 {
		// Start from the very first line
		if _, err := s.read_file.Seek(0, io.SeekStart); err != nil {
			return fmt.Errorf("failed to seek file: %w", err)
		}
		return nil
	}

	// Rewind to start
	if _, err := s.read_file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek file: %w", err)
	}

	// Skip (line-1) lines
	r := bufio.NewReader(s.read_file)
	for i := 1; i < line; i++ {
		if _, err := r.ReadString('\n'); err != nil {
			if err == io.EOF {
				return fmt.Errorf("requested line %d exceeds file length", line)
			}
			return fmt.Errorf("failed while skipping to line %d: %w", line, err)
		}
	}

	// The next read starts at the beginning of the requested line.
	return nil
}

func (s *FileMemoryStore) appendFileEvent(event *types.Event) error {
	// create a storedFileEvent
	fileEvent := &storedFileEvent{
		OriginatorID:      event.Originator.ID,
		OriginatorVersion: uint64(event.Originator.Version),
		EventType:         event.EventType,
		Payload:           event.Payload,
		CreatedAt:         time.Now().Unix(),
	}

	// write the fileEvent to a file
	eventData, err := json.Marshal(fileEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %v", err)
	}

	if _, err := s.file.Write(append(eventData, '\n')); err != nil {
		return fmt.Errorf("failed to write event to file: %v", err)
	}

	s.file.Sync()

	s.lastEvent = fileEvent
	s.lastEventLine++

	entityType := common.ExtractEntityType(event)
	logEntry := &storedLogEntry{
		ID:            uint64(len(s.storedLogEntries) + 1),
		ApplicationID: "default",
		PartitionID:   entityType,
		// keep payload empty for now, we can read the event from the file when we read the log entries
		// it's kind of lazy loading, we can read the event from the file when we read the log entries, and we can also use the line number to seek the file for pagination
		EventOriginatorId:      event.Originator.ID,
		EventOriginatorVersion: event.Originator.Version,
		CreatedAt:              time.Now().Unix(),
		LineNumber:             s.lastEventLine,
	}

	s.storedLogEntries = append(s.storedLogEntries, logEntry)

	return nil
}
