import os
from time import sleep
import grpc
from retrying import retry

from pyservices.store.crudsvc import CrudStoreService
from eventstore import service_pb2_grpc as esgrpc
from crudstore import service_pb2_grpc as crudgrpc

import concurrent.futures as futures


@retry(wait_exponential_multiplier=1000, wait_exponential_max=10000)
def get_store(store_uri: str) -> esgrpc.EventstoreServiceStub:
    channel = grpc.insecure_channel(store_uri)
    return esgrpc.EventstoreServiceStub(channel=channel)


def run():
    db_uri = os.getenv("DB_URI")
    if not db_uri:
        raise RuntimeError("DB_URI is required env variable")

    listview_db_uri = os.getenv("DB_URI_LISTVIEW")
    if not listview_db_uri:
        raise RuntimeError("DB_URI_LISTVIEW is required env variable")

    event_store_uri = os.getenv("EVENT_STORE_ENDPOINT")
    if not event_store_uri:
        raise RuntimeError("EVENT_STORE_ENDPOINT is required env variable")

    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))

    store = get_store(store_uri=event_store_uri)

    crud_store_svc = CrudStoreService(estore=store)
    crudgrpc.add_CrudStoreServiceServicer_to_server(crud_store_svc, server)

    server.add_insecure_port('[::]:9090')
    server.start()

    while True:
        sleep(60)


if __name__ == "__main__":
    run()
