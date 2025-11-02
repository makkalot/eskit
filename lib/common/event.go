package common

import (
	"fmt"
	"github.com/makkalot/eskit/lib/types"
	"strings"
)

// IncrOriginator creates a new Originator with an incremented version
func IncrOriginator(originator *types.Originator) (*types.Originator, error) {
	if originator.Version == 0 {
		return nil, fmt.Errorf("missing version")
	}

	return &types.Originator{
		ID:      originator.ID,
		Version: originator.Version + 1,
	}, nil
}

func ExtractEntityType(event *types.Event) string {
	return ExtractEntityTypeFromStr(event.EventType)
}

func ExtractEntityTypeFromStr(eventStr string) string {
	parts := strings.Split(eventStr, ".")
	entityType := strings.Join(parts[:len(parts)-1], ".")
	return entityType
}

func ExtractEventType(event *types.Event) string {
	return ExtractEventTypeFromStr(event.EventType)
}

func ExtractEventTypeFromStr(eventStr string) string {
	parts := strings.Split(eventStr, ".")
	return parts[len(parts)-1]
}

func IsEventCompliant(event *types.Event, selector string) bool {
	if selector == "" || selector == "*" {
		return true
	}

	selectorEntityType := ExtractEntityTypeFromStr(selector)
	selectorEventType := ExtractEventTypeFromStr(selector)

	entityType := ExtractEntityType(event)
	eventName := ExtractEventType(event)

	if selectorEntityType != "*" && selectorEntityType != entityType {
		return false
	}

	if selectorEventType != "*" && selectorEventType != eventName {
		return false
	}

	return true
}
