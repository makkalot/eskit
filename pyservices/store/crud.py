from copy import deepcopy
import datetime
import jsonpatch
import json

from common import originator_pb2 as common
import pyservices.generated.eventstore.service_pb2 as es
import pyservices.generated.eventstore.event_pb2 as esdata
import pyservices.generated.eventstore.service_pb2_grpc as esgrpc


class NotFoundError(Exception):
    pass


class CrudStore(object):
    def __init__(self, store: esgrpc.EventstoreServiceStub) -> None:
        self.store = store

    def create(self, entity_type: str, originator: common.Originator, payload: str) -> common.Originator:
        """
        creates a new entity_type.Created event type inside the event store
        """
        if not originator:
            raise ValueError("originator can not be empty")

        if not originator.version:
            originator.version = "1"

        payload_dict = json.loads(payload)
        if "originator" in payload_dict:
            del payload_dict["originator"]

        payload = json.dumps(payload_dict)

        event = esdata.Event(
            originator=originator,
            event_type="{}.Created".format(entity_type),
            payload=payload,
            occured_on=int(datetime.datetime.utcnow().timestamp())
        )

        self.store.Append(es.AppendEventRequest(event=event))
        return originator

    def update(self, entity_type: str, originator: common.Originator, payload: str) -> common.Originator:
        """
        update fetches the latest version that's inside of the db
        tries to produce a json patch for it and appends it to the event stream
        """
        if not originator.version:
            raise ValueError("missing version")

        new_originator = deepcopy(originator)
        new_originator.version = str(int(new_originator.version) + 1)

        latest_obj, _ = self.get(entity_type=entity_type, originator=originator)
        apply_obj = json.loads(payload)
        if "originator" in apply_obj:
            del apply_obj["originator"]

        patch = jsonpatch.make_patch(latest_obj, apply_obj)
        if not patch:
            raise ValueError("nothing to update")
        patch_event_payload = patch.to_string()

        event = esdata.Event(
            originator=new_originator,
            event_type="{}.Updated".format(entity_type),
            payload=patch_event_payload,
            occured_on=int(datetime.datetime.utcnow().timestamp())
        )
        self.store.Append(es.AppendEventRequest(event=event))
        return new_originator

    def get(self, entity_type: str, originator: common.Originator, deleted: bool = False,
            include_originator=False) -> (dict, common.Originator):
        """
        Tries to run all the events till to the point and construct back the object from patches
        """
        resp = self.store.GetEvents(es.GetEventsRequest(originator=originator))
        if not resp:
            raise NotFoundError("not found")

        events = resp.events
        if not events:
            raise NotFoundError("not found")

        # print("Get EVENTS : ", events)
        if not deleted and self._is_deleted_event(events[-1].event_type):
            raise NotFoundError("object deleted : {}".format(originator))

        # print("Events : {}".format(events))
        obj = json.loads(events[0].payload)
        latest_originator = events[0].originator
        for event in events[1:]:
            if not self._is_crud_event(event.event_type):
                raise ValueError("don't know how to play event : {}".format(event.event_type))

            latest_originator = event.originator
            # ignore patching the deleted events
            if self._is_deleted_event(event.event_type):
                continue

            json_patch = jsonpatch.JsonPatch.from_string(event.payload)
            # print("Applying patch : {}".format(json_patch))
            # print("To : {}".format(obj))

            obj = json_patch.apply(obj)
            # print("Applied : ", obj)

        # print("Latest Object : ", obj)
        originator = None
        if include_originator:
            originator = latest_originator
        return obj, originator

    def delete(self, entity_type: str, originator: common.Originator) -> common.Originator:
        payload, latest_originator = self.get(entity_type=entity_type, originator=originator, include_originator=True)
        if not latest_originator:
            raise ValueError("no originator found in get payload")

        new_originator = common.Originator(
            id=latest_originator.id,
            version=str(int(latest_originator.version) + 1)
        )

        event = esdata.Event(
            originator=new_originator,
            event_type="{}.Deleted".format(entity_type),
            payload="{}",
            occured_on=int(datetime.datetime.utcnow().timestamp())
        )
        self.store.Append(es.AppendEventRequest(event=event))
        return new_originator

    @staticmethod
    def _is_crud_event(event_type: str) -> bool:
        parts = event_type.split(".")
        if not parts[-1].lower() in ["created", "updated", "deleted"]:
            return False
        return True

    def _extract_event_name(self, event_type):
        return event_type.split(".")[-1].lower()

    def _is_deleted_event(self, event_type):
        event_name = self._extract_event_name(event_type)
        return event_name == "deleted"
