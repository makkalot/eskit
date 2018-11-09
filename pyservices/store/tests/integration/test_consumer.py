import unittest
import uuid
import os
from functools import partial

import grpc
from retrying import retry
import threading
from queue import Queue

from pyservices.store.clients.consumer import ApplicationLogConsumer, ConsumerOffset
from pyservices.store.clients import get_store_client
from eventstore import event_pb2 as esdata
from eventstore import service_pb2 as es
from common import originator_pb2 as common


@retry(stop_max_delay=10000, retry_on_exception=lambda ex: ex.code() == grpc.StatusCode.UNAVAILABLE)
def wait_health(client):
    _ = client.Healtz(request=es.HealthRequest())


class TestConsumer(unittest.TestCase):
    def setUp(self):
        if not os.getenv("EVENTSTORE_ENDPOINT"):
            raise RuntimeError("EVENTSTORE_ENDPOINT env variable is required")

        store_endpoint = os.getenv("EVENTSTORE_ENDPOINT")

        consumer_endpoint = os.getenv("CONSUMERSTORE_ENDPOINT")
        if not consumer_endpoint:
            raise RuntimeError("CONSUMERSTORE_ENDPOINT env variable is required")

        self.consumer_name = str(uuid.uuid4())

        self.client = get_store_client(store_uri=store_endpoint, consumer_uri=consumer_endpoint)
        wait_health(client=self.client.estore)
        wait_health(client=self.client.consumerstore)

    def print_consumer(self, queue: Queue, log_entry: esdata.AppLogEntry) -> None:
        # print(f" Consumer CB ID : {log_entry.id} - Event : {log_entry.event}")
        if log_entry.event.event_type.split(".")[0] == "PyConsumerUser":
            queue.put_nowait(log_entry)

    def test(self):
        queue = Queue()

        consumer = ApplicationLogConsumer(
            es_client=self.client.estore,
            consumer_client=self.client.consumerstore,
            consumer_name=self.consumer_name,
            cb=partial(self.print_consumer, queue),
            offset=ConsumerOffset.FROM_SAVED,
        )

        originator = common.Originator(
            id=str(uuid.uuid4()),
            version="1",
        )

        event_1 = esdata.Event(
            originator=originator,
            event_type="PyConsumerUser.Created",
            payload="{}"
        )

        _ = self.client.estore.Append(
            request=es.AppendEventRequest(
                event=event_1
            )
        )

        event = threading.Event()
        future = consumer.consume_async(cancel_event=event)
        log_entry = queue.get()

        assert log_entry.event.SerializeToString() == event_1.SerializeToString()
        future.cancel()
        event.set()

        # append a second marker one to stop the stream
        _ = self.client.estore.Append(
            request=es.AppendEventRequest(
                event=esdata.Event(
                    originator=common.Originator(
                        id=str(uuid.uuid4()),
                        version="1",
                    ),
                    event_type="Terminator.Created",
                    payload="{}"
                )
            )
        )

        future.result()

        # now start the consumer again and try to send a second event into the stream

        consumer = ApplicationLogConsumer(
            es_client=self.client.estore,
            consumer_client=self.client.consumerstore,
            consumer_name=self.consumer_name,
            cb=partial(self.print_consumer, queue),
            offset=ConsumerOffset.FROM_SAVED,
        )

        originator = common.Originator(
            id=originator.id,
            version="2",
        )

        event_2 = esdata.Event(
            originator=originator,
            event_type="PyConsumerUser.Updated",
            payload="{}"
        )

        _ = self.client.estore.Append(
            request=es.AppendEventRequest(
                event=event_2
            )
        )

        event = threading.Event()
        future = consumer.consume_async(cancel_event=event)
        log_entry = queue.get()
        assert log_entry.event.SerializeToString() == event_2.SerializeToString()
        future.cancel()
        event.set()

        # append a second marker one to stop the stream
        _ = self.client.estore.Append(
            request=es.AppendEventRequest(
                event=esdata.Event(
                    originator=common.Originator(
                        id=str(uuid.uuid4()),
                        version="1",
                    ),
                    event_type="Terminator.Created",
                    payload="{}"
                )
            )
        )

        future.result()
