package common

/**
This file contains common data structures used by the eskit library
They were converted from protobuf to golang structs we used in the past

message Originator {
  string id = 1;
  string version = 2;
}

// Event is what you operate on with event store
// It's the smalled bit that the event store is aware of
message Event {
    // The object this event belongs to
    contracts.common.Originator originator = 1;
    // this is the event type that this is related to
    // event type should be in the format of `Entity.Created` so that store can infer the
    // partition this event belongs to
    string event_type = 2;
    // the data of the event is inside the payload
    string payload = 3;
    // utc unix timestamp of the event occurence
    int64 occured_on = 4;
}

message AppLogEntry {
    //  the id number in the stream
    string id = 1;
    Event event = 2;
}
**/

// Originator is the originator of an event
type Originator struct {
	Id      string `json:"id"`
	Version string `json:"version"`
}

// Event is the event that is stored in the event store
type Event struct {
	Originator *Originator `json:"originator"`
	EventType  string      `json:"event_type"`
	Payload    string      `json:"payload"`
	OccurredOn int64       `json:"occurred_on"`
}

// AppLogEntry is the log entry that is stored in the app log
type AppLogEntry struct {
	Id    string `json:"id"`
	Event *Event `json:"event"`
}
