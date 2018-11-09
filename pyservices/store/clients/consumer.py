from typing import Callable
import concurrent.futures as futures
import enum
import threading
import pyservices.generated.eventstore.service_pb2 as es
import pyservices.generated.eventstore.event_pb2 as esdata
import pyservices.generated.eventstore.service_pb2_grpc as esgrpc
import pyservices.generated.consumerstore.service_pb2_grpc as consumergrpc
import pyservices.generated.consumerstore.service_pb2 as consumerdata
import grpc


class ConsumerOffset(enum.Enum):
    FROM_BEGINNING = 0
    FROM_SAVED = 1


ConsumerCB = Callable[[esdata.AppLogEntry], None]


class ApplicationLogConsumer(object):
    excludes = ["LogConsumer"]

    def __init__(self, es_client: esgrpc.EventstoreServiceStub, consumer_client: consumergrpc.ConsumerServiceStub,
                 consumer_name: str, cb: ConsumerCB, offset: ConsumerOffset = ConsumerOffset.FROM_BEGINNING,
                 selector: str = "*", save_progress=True):
        self.es_client = es_client
        self.consumer_client = consumer_client
        self.consumer_name = consumer_name
        self.offset = offset
        self.cb = cb
        self.selector = selector
        self.save_progress = save_progress

    def consume(self, cancel_event: threading.Event) -> None:
        """
        Infinite loop of calling ConsumerCB unless one of the optional
        parameters are supplied
        """
        latest_offset = self._find_consumer_offset()
        stream = self.es_client.LogsPoll(es.AppLogRequest(
            from_id=latest_offset,
            selector=self.selector,
        ))

        for log_entry in stream:
            if cancel_event.is_set():
                return

            entity_type = log_entry.event.event_type.split(".")[0]
            if entity_type in self.excludes:
                continue

            self.cb(log_entry)
            if not self.save_progress:
                continue

            # save the progress here of the object
            self.consumer_client.LogConsume(
                consumerdata.AppLogConsumeRequest(
                    consumer_id=self.consumer_name,
                    offset=log_entry.id,
                )
            )

    def consume_async(self, cancel_event: threading.Event) -> futures.Future:
        """
        Same as consume but returns future instead of blocking
        """
        executor = futures.ThreadPoolExecutor(max_workers=1)
        return executor.submit(self.consume, cancel_event)

    def _find_consumer_offset(self) -> str:
        if self.offset == ConsumerOffset.FROM_BEGINNING:
            return "1"

        try:
            get_log_resp = self.consumer_client.GetLogConsume(consumerdata.GetAppLogConsumeRequest(
                consumer_id=self.consumer_name,
            ))
        except grpc.RpcError as ex:
            if ex.code() == grpc.StatusCode.NOT_FOUND:
                return "1"
            raise

        return str(int(get_log_resp.offset) + 1)
