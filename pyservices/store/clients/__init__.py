import grpc
from eventstore import service_pb2_grpc as esgrpc
from crudstore import service_pb2_grpc as crudgrpc
from consumerstore import service_pb2_grpc as consumergrpc


class CombinedClient(object):
    estore: esgrpc.EventstoreServiceStub
    crudstore: crudgrpc.CrudStoreServiceStub
    consumerstore: consumergrpc.ConsumerServiceStub

    def __init__(self, store_uri: str, consumer_uri: str) -> None:
        channel = grpc.insecure_channel(store_uri)
        self.estore = esgrpc.EventstoreServiceStub(channel=channel)
        self.crudstore = crudgrpc.CrudStoreServiceStub(channel=channel)

        consumer_channel = grpc.insecure_channel(consumer_uri)
        self.consumerstore = consumergrpc.ConsumerServiceStub(channel=consumer_channel)


def get_store_client(store_uri: str, consumer_uri: str) -> CombinedClient:
    return CombinedClient(
        store_uri=store_uri,
        consumer_uri=consumer_uri,
    )
