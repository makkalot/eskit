# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
import grpc

from users import service_pb2 as users_dot_service__pb2


class UserServiceStub(object):
  # missing associated documentation comment in .proto file
  pass

  def __init__(self, channel):
    """Constructor.

    Args:
      channel: A grpc.Channel.
    """
    self.Healtz = channel.unary_unary(
        '/contracts.users.UserService/Healtz',
        request_serializer=users_dot_service__pb2.HealthRequest.SerializeToString,
        response_deserializer=users_dot_service__pb2.HealthResponse.FromString,
        )
    self.Create = channel.unary_unary(
        '/contracts.users.UserService/Create',
        request_serializer=users_dot_service__pb2.CreateRequest.SerializeToString,
        response_deserializer=users_dot_service__pb2.CreateResponse.FromString,
        )
    self.Get = channel.unary_unary(
        '/contracts.users.UserService/Get',
        request_serializer=users_dot_service__pb2.GetRequest.SerializeToString,
        response_deserializer=users_dot_service__pb2.GetResponse.FromString,
        )
    self.Update = channel.unary_unary(
        '/contracts.users.UserService/Update',
        request_serializer=users_dot_service__pb2.UpdateRequest.SerializeToString,
        response_deserializer=users_dot_service__pb2.UpdateResponse.FromString,
        )
    self.Delete = channel.unary_unary(
        '/contracts.users.UserService/Delete',
        request_serializer=users_dot_service__pb2.DeleteRequest.SerializeToString,
        response_deserializer=users_dot_service__pb2.DeleteResponse.FromString,
        )


class UserServiceServicer(object):
  # missing associated documentation comment in .proto file
  pass

  def Healtz(self, request, context):
    # missing associated documentation comment in .proto file
    pass
    context.set_code(grpc.StatusCode.UNIMPLEMENTED)
    context.set_details('Method not implemented!')
    raise NotImplementedError('Method not implemented!')

  def Create(self, request, context):
    # missing associated documentation comment in .proto file
    pass
    context.set_code(grpc.StatusCode.UNIMPLEMENTED)
    context.set_details('Method not implemented!')
    raise NotImplementedError('Method not implemented!')

  def Get(self, request, context):
    # missing associated documentation comment in .proto file
    pass
    context.set_code(grpc.StatusCode.UNIMPLEMENTED)
    context.set_details('Method not implemented!')
    raise NotImplementedError('Method not implemented!')

  def Update(self, request, context):
    # missing associated documentation comment in .proto file
    pass
    context.set_code(grpc.StatusCode.UNIMPLEMENTED)
    context.set_details('Method not implemented!')
    raise NotImplementedError('Method not implemented!')

  def Delete(self, request, context):
    # missing associated documentation comment in .proto file
    pass
    context.set_code(grpc.StatusCode.UNIMPLEMENTED)
    context.set_details('Method not implemented!')
    raise NotImplementedError('Method not implemented!')


def add_UserServiceServicer_to_server(servicer, server):
  rpc_method_handlers = {
      'Healtz': grpc.unary_unary_rpc_method_handler(
          servicer.Healtz,
          request_deserializer=users_dot_service__pb2.HealthRequest.FromString,
          response_serializer=users_dot_service__pb2.HealthResponse.SerializeToString,
      ),
      'Create': grpc.unary_unary_rpc_method_handler(
          servicer.Create,
          request_deserializer=users_dot_service__pb2.CreateRequest.FromString,
          response_serializer=users_dot_service__pb2.CreateResponse.SerializeToString,
      ),
      'Get': grpc.unary_unary_rpc_method_handler(
          servicer.Get,
          request_deserializer=users_dot_service__pb2.GetRequest.FromString,
          response_serializer=users_dot_service__pb2.GetResponse.SerializeToString,
      ),
      'Update': grpc.unary_unary_rpc_method_handler(
          servicer.Update,
          request_deserializer=users_dot_service__pb2.UpdateRequest.FromString,
          response_serializer=users_dot_service__pb2.UpdateResponse.SerializeToString,
      ),
      'Delete': grpc.unary_unary_rpc_method_handler(
          servicer.Delete,
          request_deserializer=users_dot_service__pb2.DeleteRequest.FromString,
          response_serializer=users_dot_service__pb2.DeleteResponse.SerializeToString,
      ),
  }
  generic_handler = grpc.method_handlers_generic_handler(
      'contracts.users.UserService', rpc_method_handlers)
  server.add_generic_rpc_handlers((generic_handler,))