package proto

import (
	"testing"
	"time"

	"github.com/makkalot/eskit/lib/types"
	pbcommon "github.com/makkalot/eskit/generated/grpc/go/common"
	pbcrud "github.com/makkalot/eskit/generated/grpc/go/crudstore"
	pbevent "github.com/makkalot/eskit/generated/grpc/go/eventstore"
)

func TestOriginatorAdapters(t *testing.T) {
	t.Run("ToProto and FromProto round-trip", func(t *testing.T) {
		original := &types.Originator{
			ID:      "test-id-123",
			Version: "42",
		}

		pb := OriginatorToProto(original)
		if pb.Id != original.ID {
			t.Errorf("expected Id '%s', got '%s'", original.ID, pb.Id)
		}
		if pb.Version != original.Version {
			t.Errorf("expected Version '%s', got '%s'", original.Version, pb.Version)
		}

		result := OriginatorFromProto(pb)
		if result.ID != original.ID {
			t.Errorf("round-trip failed: expected ID '%s', got '%s'", original.ID, result.ID)
		}
		if result.Version != original.Version {
			t.Errorf("round-trip failed: expected Version '%s', got '%s'", original.Version, result.Version)
		}
	})

	t.Run("nil handling", func(t *testing.T) {
		pb := OriginatorToProto(nil)
		if pb != nil {
			t.Error("expected nil proto, got non-nil")
		}

		native := OriginatorFromProto(nil)
		if native != nil {
			t.Error("expected nil native, got non-nil")
		}
	})

	t.Run("FromProto to ToProto round-trip", func(t *testing.T) {
		original := &pbcommon.Originator{
			Id:      "proto-id",
			Version: "5",
		}

		native := OriginatorFromProto(original)
		result := OriginatorToProto(native)

		if result.Id != original.Id {
			t.Errorf("round-trip failed: expected Id '%s', got '%s'", original.Id, result.Id)
		}
		if result.Version != original.Version {
			t.Errorf("round-trip failed: expected Version '%s', got '%s'", original.Version, result.Version)
		}
	})
}

func TestEventAdapters(t *testing.T) {
	t.Run("ToProto and FromProto round-trip", func(t *testing.T) {
		timestamp := time.Date(2024, 1, 15, 12, 30, 45, 0, time.UTC)
		original := &types.Event{
			Originator: &types.Originator{
				ID:      "event-id-1",
				Version: "1",
			},
			EventType:  "User.Created",
			Payload:    `{"name":"Alice","email":"alice@example.com"}`,
			OccurredOn: timestamp,
		}

		pb := EventToProto(original)
		if pb.EventType != original.EventType {
			t.Errorf("expected EventType '%s', got '%s'", original.EventType, pb.EventType)
		}
		if pb.Payload != original.Payload {
			t.Errorf("expected Payload '%s', got '%s'", original.Payload, pb.Payload)
		}
		if pb.Originator.Id != original.Originator.ID {
			t.Errorf("expected Originator.Id '%s', got '%s'", original.Originator.ID, pb.Originator.Id)
		}

		result := EventFromProto(pb)
		if result.EventType != original.EventType {
			t.Errorf("round-trip failed: expected EventType '%s', got '%s'", original.EventType, result.EventType)
		}
		if result.Payload != original.Payload {
			t.Errorf("round-trip failed: expected Payload '%s', got '%s'", original.Payload, result.Payload)
		}
		// Check timestamp (allowing for Unix second precision)
		if result.OccurredOn.Unix() != original.OccurredOn.Unix() {
			t.Errorf("round-trip failed: expected OccurredOn %v, got %v", original.OccurredOn, result.OccurredOn)
		}
	})

	t.Run("nil handling", func(t *testing.T) {
		pb := EventToProto(nil)
		if pb != nil {
			t.Error("expected nil proto, got non-nil")
		}

		native := EventFromProto(nil)
		if native != nil {
			t.Error("expected nil native, got non-nil")
		}
	})

	t.Run("zero timestamp handling", func(t *testing.T) {
		original := &types.Event{
			Originator: &types.Originator{ID: "id", Version: "1"},
			EventType:  "Test.Event",
			Payload:    "{}",
			OccurredOn: time.Time{}, // zero time
		}

		pb := EventToProto(original)
		if pb.OccuredOn != 0 {
			t.Errorf("expected OccuredOn 0, got %d", pb.OccuredOn)
		}

		result := EventFromProto(pb)
		if !result.OccurredOn.IsZero() {
			t.Error("expected zero time, got non-zero")
		}
	})
}

func TestAppLogEntryAdapters(t *testing.T) {
	t.Run("ToProto and FromProto round-trip", func(t *testing.T) {
		original := &types.AppLogEntry{
			ID: "log-entry-123",
			Event: &types.Event{
				Originator: &types.Originator{
					ID:      "entity-1",
					Version: "3",
				},
				EventType:  "Order.Placed",
				Payload:    `{"orderId":"order-123","total":100.50}`,
				OccurredOn: time.Now().UTC(),
			},
		}

		pb := AppLogEntryToProto(original)
		if pb.Id != original.ID {
			t.Errorf("expected Id '%s', got '%s'", original.ID, pb.Id)
		}
		if pb.Event == nil {
			t.Fatal("expected Event to be non-nil")
		}
		if pb.Event.EventType != original.Event.EventType {
			t.Errorf("expected Event.EventType '%s', got '%s'", original.Event.EventType, pb.Event.EventType)
		}

		result := AppLogEntryFromProto(pb)
		if result.ID != original.ID {
			t.Errorf("round-trip failed: expected ID '%s', got '%s'", original.ID, result.ID)
		}
		if result.Event.EventType != original.Event.EventType {
			t.Errorf("round-trip failed: expected Event.EventType '%s', got '%s'", original.Event.EventType, result.Event.EventType)
		}
	})

	t.Run("nil handling", func(t *testing.T) {
		pb := AppLogEntryToProto(nil)
		if pb != nil {
			t.Error("expected nil proto, got non-nil")
		}

		native := AppLogEntryFromProto(nil)
		if native != nil {
			t.Error("expected nil native, got non-nil")
		}
	})
}

func TestSchemaSpecAdapters(t *testing.T) {
	t.Run("ToProto and FromProto round-trip", func(t *testing.T) {
		original := &types.SchemaSpec{
			SchemaVersion: 5,
			JSONSchema:    `{"type":"object","properties":{"name":{"type":"string"}}}`,
		}

		pb := SchemaSpecToProto(original)
		if pb.SchemaVersion != original.SchemaVersion {
			t.Errorf("expected SchemaVersion %d, got %d", original.SchemaVersion, pb.SchemaVersion)
		}
		if pb.JsonSchema != original.JSONSchema {
			t.Errorf("expected JsonSchema '%s', got '%s'", original.JSONSchema, pb.JsonSchema)
		}

		result := SchemaSpecFromProto(pb)
		if result.SchemaVersion != original.SchemaVersion {
			t.Errorf("round-trip failed: expected SchemaVersion %d, got %d", original.SchemaVersion, result.SchemaVersion)
		}
		if result.JSONSchema != original.JSONSchema {
			t.Errorf("round-trip failed: expected JSONSchema '%s', got '%s'", original.JSONSchema, result.JSONSchema)
		}
	})

	t.Run("nil handling", func(t *testing.T) {
		pb := SchemaSpecToProto(nil)
		if pb != nil {
			t.Error("expected nil proto, got non-nil")
		}

		native := SchemaSpecFromProto(nil)
		if native != nil {
			t.Error("expected nil native, got non-nil")
		}
	})
}

func TestCrudEntitySpecAdapters(t *testing.T) {
	t.Run("ToProto and FromProto round-trip", func(t *testing.T) {
		original := &types.CrudEntitySpec{
			EntityType: "Product",
			SchemaSpec: &types.SchemaSpec{
				SchemaVersion: 2,
				JSONSchema:    `{"type":"object"}`,
			},
		}

		pb := CrudEntitySpecToProto(original)
		if pb.EntityType != original.EntityType {
			t.Errorf("expected EntityType '%s', got '%s'", original.EntityType, pb.EntityType)
		}
		if pb.SchemaSpec == nil {
			t.Fatal("expected SchemaSpec to be non-nil")
		}
		if pb.SchemaSpec.SchemaVersion != original.SchemaSpec.SchemaVersion {
			t.Errorf("expected SchemaSpec.SchemaVersion %d, got %d",
				original.SchemaSpec.SchemaVersion, pb.SchemaSpec.SchemaVersion)
		}

		result := CrudEntitySpecFromProto(pb)
		if result.EntityType != original.EntityType {
			t.Errorf("round-trip failed: expected EntityType '%s', got '%s'", original.EntityType, result.EntityType)
		}
		if result.SchemaSpec.SchemaVersion != original.SchemaSpec.SchemaVersion {
			t.Errorf("round-trip failed: expected SchemaVersion %d, got %d",
				original.SchemaSpec.SchemaVersion, result.SchemaSpec.SchemaVersion)
		}
	})

	t.Run("nil handling", func(t *testing.T) {
		pb := CrudEntitySpecToProto(nil)
		if pb != nil {
			t.Error("expected nil proto, got non-nil")
		}

		native := CrudEntitySpecFromProto(nil)
		if native != nil {
			t.Error("expected nil native, got non-nil")
		}
	})

	t.Run("nil SchemaSpec handling", func(t *testing.T) {
		original := &types.CrudEntitySpec{
			EntityType: "User",
			SchemaSpec: nil,
		}

		pb := CrudEntitySpecToProto(original)
		if pb.SchemaSpec != nil {
			t.Error("expected SchemaSpec to be nil")
		}

		result := CrudEntitySpecFromProto(pb)
		if result.SchemaSpec != nil {
			t.Error("round-trip failed: expected SchemaSpec to be nil")
		}
	})
}

func TestComplexRoundTrips(t *testing.T) {
	t.Run("AppLogEntry with full nested structure", func(t *testing.T) {
		// Create a complex nested structure
		original := &types.AppLogEntry{
			ID: "complex-123",
			Event: &types.Event{
				Originator: &types.Originator{
					ID:      "nested-entity",
					Version: "10",
				},
				EventType:  "Complex.Event.Type",
				Payload:    `{"nested":{"data":{"value":123}}}`,
				OccurredOn: time.Date(2024, 6, 15, 10, 30, 45, 123456789, time.UTC),
			},
		}

		// Convert to proto
		pb := AppLogEntryToProto(original)

		// Convert back
		result := AppLogEntryFromProto(pb)

		// Verify all nested fields
		if result.ID != original.ID {
			t.Error("ID mismatch in complex round-trip")
		}
		if result.Event.Originator.ID != original.Event.Originator.ID {
			t.Error("Nested Originator.ID mismatch")
		}
		if result.Event.Originator.Version != original.Event.Originator.Version {
			t.Error("Nested Originator.Version mismatch")
		}
		if result.Event.EventType != original.Event.EventType {
			t.Error("EventType mismatch")
		}
		if result.Event.Payload != original.Event.Payload {
			t.Error("Payload mismatch")
		}
		// Timestamp precision is limited to seconds
		if result.Event.OccurredOn.Unix() != original.Event.OccurredOn.Unix() {
			t.Errorf("OccurredOn mismatch: expected %v, got %v",
				original.Event.OccurredOn, result.Event.OccurredOn)
		}
	})

	t.Run("Verify proto types can be used in gRPC", func(t *testing.T) {
		// This test ensures our adapters produce valid protobuf types
		native := &types.Event{
			Originator: &types.Originator{ID: "test", Version: "1"},
			EventType:  "Test.Event",
			Payload:    "{}",
			OccurredOn: time.Now().UTC(),
		}

		pb := EventToProto(native)

		// These type assertions ensure the proto types are correct
		var _ *pbevent.Event = pb
		var _ *pbcommon.Originator = pb.Originator
	})
}
