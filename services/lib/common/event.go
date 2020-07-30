package common

import (
	"fmt"
	"github.com/makkalot/eskit/generated/grpc/go/common"
	store "github.com/makkalot/eskit/generated/grpc/go/eventstore"
	"strconv"
	"strings"
)

func IncrStringInt(s string) (string, error) {
	versionInt, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return "", fmt.Errorf("parse : %v", err)
	}

	versionInt++
	return strconv.Itoa(int(versionInt)), nil
}

func MustIncrStringInt(s string) string {
	newVersion, err := IncrStringInt(s)
	if err != nil {
		panic(fmt.Sprintf("incrementing version failed : %v", err))
	}
	return newVersion
}

func IncrOriginator(originator *common.Originator) (*common.Originator, error) {
	if originator.Version == "" {
		return nil, fmt.Errorf("missing version")
	}

	newVersion, err := IncrStringInt(originator.Version)
	if err != nil {
		return nil, err
	}

	return &common.Originator{
		Id:      originator.Id,
		Version: newVersion,
	}, nil
}

func ExtractEntityType(event *store.Event) string {
	return ExtractEntityTypeFromStr(event.EventType)
}

func ExtractEntityTypeFromStr(eventStr string) string {
	parts := strings.Split(eventStr, ".")
	entityType := strings.Join(parts[:len(parts)-1], ".")
	return entityType
}

func ExtractEventType(event *store.Event) string {
	return ExtractEventTypeFromStr(event.EventType)
}

func ExtractEventTypeFromStr(eventStr string) string {
	parts := strings.Split(eventStr, ".")
	return parts[len(parts)-1]
}
