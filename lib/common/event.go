package common

import (
	"fmt"
	"strconv"
	"strings"
)

// IncrStringInt increments a string representing an integer
func IncrStringInt(s string) (string, error) {
	versionInt, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return "", fmt.Errorf("parse : %v", err)
	}

	versionInt++
	return strconv.Itoa(int(versionInt)), nil
}

// MustIncrStringInt increments a string representing an integer and panics on error
func MustIncrStringInt(s string) string {
	newVersion, err := IncrStringInt(s)
	if err != nil {
		panic(fmt.Sprintf("incrementing version failed : %v", err))
	}
	return newVersion
}

// IncrOriginator increments the version of an originator
func IncrOriginator(originator *Originator) (*Originator, error) {
	if originator.Version == "" {
		return nil, fmt.Errorf("missing version")
	}

	newVersion, err := IncrStringInt(originator.Version)
	if err != nil {
		return nil, err
	}

	return &Originator{
		Id:      originator.Id,
		Version: newVersion,
	}, nil
}

func ExtractEntityType(event *Event) string {
	return ExtractEntityTypeFromStr(event.EventType)
}

func ExtractEntityTypeFromStr(eventStr string) string {
	parts := strings.Split(eventStr, ".")
	entityType := strings.Join(parts[:len(parts)-1], ".")
	return entityType
}

func ExtractEventType(event *Event) string {
	return ExtractEventTypeFromStr(event.EventType)
}

func ExtractEventTypeFromStr(eventStr string) string {
	parts := strings.Split(eventStr, ".")
	return parts[len(parts)-1]
}

func IsEventCompliant(event *Event, selector string) bool {
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
