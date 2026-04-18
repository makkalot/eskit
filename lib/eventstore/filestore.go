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
	lastByteOffset   int64
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
	LineNumber int64
	// byte offset of the start of this line in the file
	ByteOffset int64
}

func NewFileMemoryStore(storePath string) (*FileMemoryStore, error) {
	basePath := filepath.Dir(storePath)

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		if err := os.MkdirAll(basePath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create store path: %v", err)
		}
	}

	file, err := os.OpenFile(storePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open store file: %v", err)
	}

	read_file, err := os.OpenFile(storePath, os.O_RDONLY, 0644)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to open store file for reading: %v", err)
	}

	s := &FileMemoryStore{
		storePath: storePath,
		file:      file,
		read_file: read_file,
	}

	if err := s.loadFromFile(); err != nil {
		file.Close()
		read_file.Close()
		return nil, fmt.Errorf("failed to load existing events from file: %v", err)
	}

	return s, nil
}

// loadFromFile scans the existing file and reconstructs storedLogEntries,
// lastEvent, lastEventLine, and lastByteOffset so that the store can resume
// correctly after being reopened.
func (s *FileMemoryStore) loadFromFile() error {
	// Rewind to start for reading.
	if _, err := s.read_file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek file: %w", err)
	}

	scanner := bufio.NewScanner(s.read_file)
	var lineNumber int64
	var byteOffset int64

	for scanner.Scan() {
		lineBytes := scanner.Bytes()
		lineLen := int64(len(lineBytes)) + 1 // +1 for newline

		var fileEvent storedFileEvent
		if err := json.Unmarshal(lineBytes, &fileEvent); err != nil {
			return fmt.Errorf("failed to unmarshal event on line %d: %w", lineNumber+1, err)
		}

		lineNumber++
		entityType := common.ExtractEntityTypeFromStr(fileEvent.EventType)
		logEntry := &storedLogEntry{
			ID:                     uint64(lineNumber),
			ApplicationID:          "default",
			PartitionID:            entityType,
			EventOriginatorId:      fileEvent.OriginatorID,
			EventOriginatorVersion: fileEvent.OriginatorVersion,
			CreatedAt:              fileEvent.CreatedAt,
			LineNumber:             lineNumber,
			ByteOffset:             byteOffset,
		}

		s.storedLogEntries = append(s.storedLogEntries, logEntry)
		s.lastEvent = &fileEvent
		s.lastEventLine = lineNumber
		s.lastByteOffset += lineLen
		byteOffset += lineLen
	}

	return scanner.Err()
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
	var qualifying []*storedLogEntry

	for _, logEntry := range s.storedLogEntries {
		if logEntry.EventOriginatorId != originator.ID {
			continue
		}
		if originator.Version != 0 {
			if fromVersion {
				// return events with version >= originator.Version
				if logEntry.EventOriginatorVersion < originator.Version {
					continue
				}
			} else {
				// return events with version <= originator.Version
				if logEntry.EventOriginatorVersion > originator.Version {
					continue
				}
			}
		}
		qualifying = append(qualifying, logEntry)
	}

	if len(qualifying) == 0 {
		return events, nil
	}

	// Build byte-offset lookup for qualifying entries.
	offsetMap := make(map[int64]*storedLogEntry, len(qualifying))
	var firstOffset, lastOffset int64
	for i, q := range qualifying {
		offsetMap[q.ByteOffset] = q
		if i == 0 || q.ByteOffset < firstOffset {
			firstOffset = q.ByteOffset
		}
		if i == 0 || q.ByteOffset > lastOffset {
			lastOffset = q.ByteOffset
		}
	}

	// Seek to the first qualifying entry by byte offset.
	if _, err := s.read_file.Seek(firstOffset, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek file: %w", err)
	}

	scanner := bufio.NewScanner(s.read_file)
	currentOffset := firstOffset

	for scanner.Scan() {
		lineBytes := scanner.Bytes()
		lineLen := int64(len(lineBytes)) + 1 // +1 for newline

		if _, ok := offsetMap[currentOffset]; ok {
			var fileEvent storedFileEvent
			if err := json.Unmarshal(lineBytes, &fileEvent); err != nil {
				return nil, fmt.Errorf("failed to unmarshal event at offset %d: %w", currentOffset, err)
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

		currentOffset += lineLen
		if currentOffset > lastOffset {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan file: %w", err)
	}

	return events, nil
}

func (s *FileMemoryStore) Logs(fromID uint64, size uint32, pipelineID string) ([]*types.AppLogEntry, error) {
	if size == 0 {
		size = 20
	}

	// Collect qualifying log entries in order.
	var qualified []*storedLogEntry
	for _, logEntry := range s.storedLogEntries {
		if logEntry.ID < fromID {
			continue
		}
		if pipelineID != "" && logEntry.PartitionID != pipelineID {
			continue
		}
		qualified = append(qualified, logEntry)
		if uint32(len(qualified)) >= size {
			break
		}
	}

	if len(qualified) == 0 {
		return nil, nil
	}

	// Build byte-offset lookup.
	offsetMap := make(map[int64]*storedLogEntry, len(qualified))
	var firstOffset, lastOffset int64
	for i, q := range qualified {
		offsetMap[q.ByteOffset] = q
		if i == 0 || q.ByteOffset < firstOffset {
			firstOffset = q.ByteOffset
		}
		if i == 0 || q.ByteOffset > lastOffset {
			lastOffset = q.ByteOffset
		}
	}

	// Seek to first qualifying entry.
	if _, err := s.read_file.Seek(firstOffset, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek file: %w", err)
	}

	var logs []*types.AppLogEntry
	scanner := bufio.NewScanner(s.read_file)
	currentOffset := firstOffset

	for scanner.Scan() {
		lineBytes := scanner.Bytes()
		lineLen := int64(len(lineBytes)) + 1 // +1 for newline

		if logEntry, ok := offsetMap[currentOffset]; ok {
			var fileEvent storedFileEvent
			if err := json.Unmarshal(lineBytes, &fileEvent); err != nil {
				return nil, fmt.Errorf("failed to unmarshal event for log entry %d: %w", logEntry.ID, err)
			}

			payload, _ := fileEvent.Payload.(string)
			event := &types.Event{
				Originator: &types.Originator{
					ID:      fileEvent.OriginatorID,
					Version: fileEvent.OriginatorVersion,
				},
				EventType:  fileEvent.EventType,
				Payload:    payload,
				OccurredOn: time.Unix(fileEvent.CreatedAt, 0).UTC(),
			}

			logs = append(logs, &types.AppLogEntry{
				ID:    logEntry.ID,
				Event: event,
			})

			if uint32(len(logs)) >= size {
				break
			}
		}

		currentOffset += lineLen
		if currentOffset > lastOffset {
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

	lineByteOffset := s.lastByteOffset
	lineBytes := append(eventData, '\n')
	if _, err := s.file.Write(lineBytes); err != nil {
		return fmt.Errorf("failed to write event to file: %v", err)
	}

	s.file.Sync()

	s.lastEvent = fileEvent
	s.lastEventLine++
	s.lastByteOffset += int64(len(lineBytes))

	entityType := common.ExtractEntityType(event)
	logEntry := &storedLogEntry{
		ID:                     uint64(len(s.storedLogEntries) + 1),
		ApplicationID:          "default",
		PartitionID:            entityType,
		EventOriginatorId:      event.Originator.ID,
		EventOriginatorVersion: event.Originator.Version,
		CreatedAt:              time.Now().Unix(),
		LineNumber:             s.lastEventLine,
		ByteOffset:             lineByteOffset,
	}

	s.storedLogEntries = append(s.storedLogEntries, logEntry)

	return nil
}
