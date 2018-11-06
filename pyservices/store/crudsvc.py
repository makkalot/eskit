import json
import uuid
import grpc
from eventstore import service_pb2_grpc as esgrpc
from crudstore import service_pb2_grpc as crudgrpc
from crudstore import service_pb2 as crudata
from common import originator_pb2 as common

from pyservices.store.crud import CrudStore
from pyservices.store.util import validate_originator
from pyservices.store.decorators import grpc_catch
from pyservices.store.decorators import get_logger

logger = get_logger()


class CrudStoreService(crudgrpc.CrudStoreServiceServicer):
    def __init__(self, estore: esgrpc.EventstoreServiceStub) -> None:
        self.event_store = estore
        self.crud_store = CrudStore(store=self.event_store)
        super(CrudStoreService, self).__init__()

    @grpc_catch(crudata.HealthResponse)
    def Healtz(self, request: crudata.HealthRequest, context: grpc.ServicerContext) -> crudata.HealthResponse:
        return crudata.HealthResponse()

    @grpc_catch(crudata.CreateResponse)
    def Create(self, request: crudata.CreateRequest, context: grpc.ServicerContext) -> crudata.CreateResponse:
        if not request.entity_type:
            raise ValueError("missing entity_type param")

        if not request.payload:
            raise ValueError("missing payload")
        json_payload = request.payload

        originator = request.originator
        if not originator:
            originator = common.Originator(
                id=str(uuid.uuid4()),
                version="1"
            )

        # logger.debug(f"The payload json is : {json_payload}")

        validate_originator(originator)
        create_originator = self.crud_store.create(
            entity_type=request.entity_type,
            originator=originator,
            payload=json_payload,
        )

        return crudata.CreateResponse(
            originator=create_originator,
        )

    @grpc_catch(crudata.UpdateResponse)
    def Update(self, request: crudata.UpdateRequest, context: grpc.ServicerContext) -> crudata.UpdateResponse:
        if not request.entity_type:
            raise ValueError("missing entity_type param")

        if not request.payload:
            raise ValueError("missing payload")

        if not request.originator:
            raise ValueError("missing originator")

        originator = request.originator
        if not originator.id or not originator.version:
            raise ValueError("originator has to have id and version on update")

        validate_originator(originator)

        json_payload = request.payload
        update_originator = self.crud_store.update(
            entity_type=request.entity_type,
            originator=originator,
            payload=json_payload,
        )

        return crudata.UpdateResponse(
            originator=update_originator,
        )

    @grpc_catch(crudata.GetResponse)
    def Get(self, request: crudata.GetRequest, context: grpc.ServicerContext) -> crudata.GetResponse:
        if not request.originator or not request.originator.id:
            raise ValueError("missing originator.id")

        validate_originator(request.originator)
        if not request.entity_type:
            raise ValueError("missing entity_type")

        payload, originator = self.crud_store.get(
            entity_type=request.entity_type,
            originator=request.originator,
            deleted=request.deleted,
            include_originator=True
        )

        payload_json = json.dumps(payload)
        return crudata.GetResponse(
            payload=payload_json,
            originator=originator
        )

    @grpc_catch(crudata.DeleteResponse)
    def Delete(self, request: crudata.DeleteRequest, context: grpc.ServicerContext) -> crudata.DeleteResponse:
        if not request.originator or not request.originator.id:
            context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
            context.set_details("missing originator.id")
            return crudata.DeleteResponse()

        validate_originator(request.originator)
        if not request.entity_type:
            context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
            context.set_details("missing entity_type")
            return crudata.DeleteResponse()

        delete_originator = self.crud_store.delete(
            entity_type=request.entity_type,
            originator=request.originator,
        )

        return crudata.DeleteResponse(
            originator=delete_originator,
        )
