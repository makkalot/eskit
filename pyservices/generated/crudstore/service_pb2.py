# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: crudstore/service.proto

import sys
_b=sys.version_info[0]<3 and (lambda x:x) or (lambda x:x.encode('latin1'))
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from common import originator_pb2 as common_dot_originator__pb2
from crudstore import schema_pb2 as crudstore_dot_schema__pb2
from google.api import annotations_pb2 as google_dot_api_dot_annotations__pb2


DESCRIPTOR = _descriptor.FileDescriptor(
  name='crudstore/service.proto',
  package='contracts.crudstore',
  syntax='proto3',
  serialized_options=_b('Z5github.com/makkalot/eskit/generated/grpc/go/crudstore'),
  serialized_pb=_b('\n\x17\x63rudstore/service.proto\x12\x13\x63ontracts.crudstore\x1a\x17\x63ommon/originator.proto\x1a\x16\x63rudstore/schema.proto\x1a\x1cgoogle/api/annotations.proto\"g\n\rCreateRequest\x12\x13\n\x0b\x65ntity_type\x18\x01 \x01(\t\x12\x30\n\noriginator\x18\x02 \x01(\x0b\x32\x1c.contracts.common.Originator\x12\x0f\n\x07payload\x18\x03 \x01(\t\"B\n\x0e\x43reateResponse\x12\x30\n\noriginator\x18\x01 \x01(\x0b\x32\x1c.contracts.common.Originator\"g\n\rUpdateRequest\x12\x13\n\x0b\x65ntity_type\x18\x01 \x01(\t\x12\x30\n\noriginator\x18\x02 \x01(\x0b\x32\x1c.contracts.common.Originator\x12\x0f\n\x07payload\x18\x03 \x01(\t\"B\n\x0eUpdateResponse\x12\x30\n\noriginator\x18\x01 \x01(\x0b\x32\x1c.contracts.common.Originator\"V\n\rDeleteRequest\x12\x13\n\x0b\x65ntity_type\x18\x01 \x01(\t\x12\x30\n\noriginator\x18\x02 \x01(\x0b\x32\x1c.contracts.common.Originator\"B\n\x0e\x44\x65leteResponse\x12\x30\n\noriginator\x18\x01 \x01(\x0b\x32\x1c.contracts.common.Originator\"d\n\nGetRequest\x12\x13\n\x0b\x65ntity_type\x18\x01 \x01(\t\x12\x30\n\noriginator\x18\x02 \x01(\x0b\x32\x1c.contracts.common.Originator\x12\x0f\n\x07\x64\x65leted\x18\x03 \x01(\x08\"P\n\x0bGetResponse\x12\x30\n\noriginator\x18\x01 \x01(\x0b\x32\x1c.contracts.common.Originator\x12\x0f\n\x07payload\x18\x02 \x01(\t\"^\n\x0bListRequest\x12\x13\n\x0b\x65ntity_type\x18\x01 \x01(\t\x12\x15\n\rpagination_id\x18\x02 \x01(\t\x12\r\n\x05limit\x18\x03 \x01(\r\x12\x14\n\x0cskip_payload\x18\x04 \x01(\x08\"j\n\x10ListResponseItem\x12\x13\n\x0b\x65ntity_type\x18\x01 \x01(\t\x12\x30\n\noriginator\x18\x02 \x01(\x0b\x32\x1c.contracts.common.Originator\x12\x0f\n\x07payload\x18\x03 \x01(\t\"\\\n\x0cListResponse\x12\x36\n\x07results\x18\x01 \x03(\x0b\x32%.contracts.crudstore.ListResponseItem\x12\x14\n\x0cnext_page_id\x18\x02 \x01(\t\"`\n\x13RegisterTypeRequest\x12\x31\n\x04spec\x18\x01 \x01(\x0b\x32#.contracts.crudstore.CrudEntitySpec\x12\x16\n\x0eskip_duplicate\x18\x02 \x01(\x08\"\x16\n\x14RegisterTypeResponse\"%\n\x0eGetTypeRequest\x12\x13\n\x0b\x65ntity_type\x18\x01 \x01(\t\"D\n\x0fGetTypeResponse\x12\x31\n\x04spec\x18\x01 \x01(\x0b\x32#.contracts.crudstore.CrudEntitySpec\"F\n\x11UpdateTypeRequest\x12\x31\n\x04spec\x18\x01 \x01(\x0b\x32#.contracts.crudstore.CrudEntitySpec\"\x14\n\x12UpdateTypeResponse\"!\n\x10ListTypesRequest\x12\r\n\x05limit\x18\x01 \x01(\r\"I\n\x11ListTypesResponse\x12\x34\n\x07results\x18\x01 \x03(\x0b\x32#.contracts.crudstore.CrudEntitySpec\"\x0f\n\rHealthRequest\"!\n\x0eHealthResponse\x12\x0f\n\x07message\x18\x01 \x01(\t2\x91\x07\n\x10\x43rudStoreService\x12\x65\n\x06Healtz\x12\".contracts.crudstore.HealthRequest\x1a#.contracts.crudstore.HealthResponse\"\x12\x82\xd3\xe4\x93\x02\x0c\x12\n/v1/healtz\x12S\n\x06\x43reate\x12\".contracts.crudstore.CreateRequest\x1a#.contracts.crudstore.CreateResponse\"\x00\x12S\n\x06Update\x12\".contracts.crudstore.UpdateRequest\x1a#.contracts.crudstore.UpdateResponse\"\x00\x12S\n\x06\x44\x65lete\x12\".contracts.crudstore.DeleteRequest\x1a#.contracts.crudstore.DeleteResponse\"\x00\x12J\n\x03Get\x12\x1f.contracts.crudstore.GetRequest\x1a .contracts.crudstore.GetResponse\"\x00\x12M\n\x04List\x12 .contracts.crudstore.ListRequest\x1a!.contracts.crudstore.ListResponse\"\x00\x12\x65\n\x0cRegisterType\x12(.contracts.crudstore.RegisterTypeRequest\x1a).contracts.crudstore.RegisterTypeResponse\"\x00\x12V\n\x07GetType\x12#.contracts.crudstore.GetTypeRequest\x1a$.contracts.crudstore.GetTypeResponse\"\x00\x12_\n\nUpdateType\x12&.contracts.crudstore.UpdateTypeRequest\x1a\'.contracts.crudstore.UpdateTypeResponse\"\x00\x12\\\n\tListTypes\x12%.contracts.crudstore.ListTypesRequest\x1a&.contracts.crudstore.ListTypesResponse\"\x00\x42\x37Z5github.com/makkalot/eskit/generated/grpc/go/crudstoreb\x06proto3')
  ,
  dependencies=[common_dot_originator__pb2.DESCRIPTOR,crudstore_dot_schema__pb2.DESCRIPTOR,google_dot_api_dot_annotations__pb2.DESCRIPTOR,])




_CREATEREQUEST = _descriptor.Descriptor(
  name='CreateRequest',
  full_name='contracts.crudstore.CreateRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='entity_type', full_name='contracts.crudstore.CreateRequest.entity_type', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='originator', full_name='contracts.crudstore.CreateRequest.originator', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='payload', full_name='contracts.crudstore.CreateRequest.payload', index=2,
      number=3, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=127,
  serialized_end=230,
)


_CREATERESPONSE = _descriptor.Descriptor(
  name='CreateResponse',
  full_name='contracts.crudstore.CreateResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='originator', full_name='contracts.crudstore.CreateResponse.originator', index=0,
      number=1, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=232,
  serialized_end=298,
)


_UPDATEREQUEST = _descriptor.Descriptor(
  name='UpdateRequest',
  full_name='contracts.crudstore.UpdateRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='entity_type', full_name='contracts.crudstore.UpdateRequest.entity_type', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='originator', full_name='contracts.crudstore.UpdateRequest.originator', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='payload', full_name='contracts.crudstore.UpdateRequest.payload', index=2,
      number=3, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=300,
  serialized_end=403,
)


_UPDATERESPONSE = _descriptor.Descriptor(
  name='UpdateResponse',
  full_name='contracts.crudstore.UpdateResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='originator', full_name='contracts.crudstore.UpdateResponse.originator', index=0,
      number=1, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=405,
  serialized_end=471,
)


_DELETEREQUEST = _descriptor.Descriptor(
  name='DeleteRequest',
  full_name='contracts.crudstore.DeleteRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='entity_type', full_name='contracts.crudstore.DeleteRequest.entity_type', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='originator', full_name='contracts.crudstore.DeleteRequest.originator', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=473,
  serialized_end=559,
)


_DELETERESPONSE = _descriptor.Descriptor(
  name='DeleteResponse',
  full_name='contracts.crudstore.DeleteResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='originator', full_name='contracts.crudstore.DeleteResponse.originator', index=0,
      number=1, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=561,
  serialized_end=627,
)


_GETREQUEST = _descriptor.Descriptor(
  name='GetRequest',
  full_name='contracts.crudstore.GetRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='entity_type', full_name='contracts.crudstore.GetRequest.entity_type', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='originator', full_name='contracts.crudstore.GetRequest.originator', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='deleted', full_name='contracts.crudstore.GetRequest.deleted', index=2,
      number=3, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=629,
  serialized_end=729,
)


_GETRESPONSE = _descriptor.Descriptor(
  name='GetResponse',
  full_name='contracts.crudstore.GetResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='originator', full_name='contracts.crudstore.GetResponse.originator', index=0,
      number=1, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='payload', full_name='contracts.crudstore.GetResponse.payload', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=731,
  serialized_end=811,
)


_LISTREQUEST = _descriptor.Descriptor(
  name='ListRequest',
  full_name='contracts.crudstore.ListRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='entity_type', full_name='contracts.crudstore.ListRequest.entity_type', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='pagination_id', full_name='contracts.crudstore.ListRequest.pagination_id', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='limit', full_name='contracts.crudstore.ListRequest.limit', index=2,
      number=3, type=13, cpp_type=3, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='skip_payload', full_name='contracts.crudstore.ListRequest.skip_payload', index=3,
      number=4, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=813,
  serialized_end=907,
)


_LISTRESPONSEITEM = _descriptor.Descriptor(
  name='ListResponseItem',
  full_name='contracts.crudstore.ListResponseItem',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='entity_type', full_name='contracts.crudstore.ListResponseItem.entity_type', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='originator', full_name='contracts.crudstore.ListResponseItem.originator', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='payload', full_name='contracts.crudstore.ListResponseItem.payload', index=2,
      number=3, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=909,
  serialized_end=1015,
)


_LISTRESPONSE = _descriptor.Descriptor(
  name='ListResponse',
  full_name='contracts.crudstore.ListResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='results', full_name='contracts.crudstore.ListResponse.results', index=0,
      number=1, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='next_page_id', full_name='contracts.crudstore.ListResponse.next_page_id', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=1017,
  serialized_end=1109,
)


_REGISTERTYPEREQUEST = _descriptor.Descriptor(
  name='RegisterTypeRequest',
  full_name='contracts.crudstore.RegisterTypeRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='spec', full_name='contracts.crudstore.RegisterTypeRequest.spec', index=0,
      number=1, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='skip_duplicate', full_name='contracts.crudstore.RegisterTypeRequest.skip_duplicate', index=1,
      number=2, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=1111,
  serialized_end=1207,
)


_REGISTERTYPERESPONSE = _descriptor.Descriptor(
  name='RegisterTypeResponse',
  full_name='contracts.crudstore.RegisterTypeResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=1209,
  serialized_end=1231,
)


_GETTYPEREQUEST = _descriptor.Descriptor(
  name='GetTypeRequest',
  full_name='contracts.crudstore.GetTypeRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='entity_type', full_name='contracts.crudstore.GetTypeRequest.entity_type', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=1233,
  serialized_end=1270,
)


_GETTYPERESPONSE = _descriptor.Descriptor(
  name='GetTypeResponse',
  full_name='contracts.crudstore.GetTypeResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='spec', full_name='contracts.crudstore.GetTypeResponse.spec', index=0,
      number=1, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=1272,
  serialized_end=1340,
)


_UPDATETYPEREQUEST = _descriptor.Descriptor(
  name='UpdateTypeRequest',
  full_name='contracts.crudstore.UpdateTypeRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='spec', full_name='contracts.crudstore.UpdateTypeRequest.spec', index=0,
      number=1, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=1342,
  serialized_end=1412,
)


_UPDATETYPERESPONSE = _descriptor.Descriptor(
  name='UpdateTypeResponse',
  full_name='contracts.crudstore.UpdateTypeResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=1414,
  serialized_end=1434,
)


_LISTTYPESREQUEST = _descriptor.Descriptor(
  name='ListTypesRequest',
  full_name='contracts.crudstore.ListTypesRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='limit', full_name='contracts.crudstore.ListTypesRequest.limit', index=0,
      number=1, type=13, cpp_type=3, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=1436,
  serialized_end=1469,
)


_LISTTYPESRESPONSE = _descriptor.Descriptor(
  name='ListTypesResponse',
  full_name='contracts.crudstore.ListTypesResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='results', full_name='contracts.crudstore.ListTypesResponse.results', index=0,
      number=1, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=1471,
  serialized_end=1544,
)


_HEALTHREQUEST = _descriptor.Descriptor(
  name='HealthRequest',
  full_name='contracts.crudstore.HealthRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=1546,
  serialized_end=1561,
)


_HEALTHRESPONSE = _descriptor.Descriptor(
  name='HealthResponse',
  full_name='contracts.crudstore.HealthResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='message', full_name='contracts.crudstore.HealthResponse.message', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=1563,
  serialized_end=1596,
)

_CREATEREQUEST.fields_by_name['originator'].message_type = common_dot_originator__pb2._ORIGINATOR
_CREATERESPONSE.fields_by_name['originator'].message_type = common_dot_originator__pb2._ORIGINATOR
_UPDATEREQUEST.fields_by_name['originator'].message_type = common_dot_originator__pb2._ORIGINATOR
_UPDATERESPONSE.fields_by_name['originator'].message_type = common_dot_originator__pb2._ORIGINATOR
_DELETEREQUEST.fields_by_name['originator'].message_type = common_dot_originator__pb2._ORIGINATOR
_DELETERESPONSE.fields_by_name['originator'].message_type = common_dot_originator__pb2._ORIGINATOR
_GETREQUEST.fields_by_name['originator'].message_type = common_dot_originator__pb2._ORIGINATOR
_GETRESPONSE.fields_by_name['originator'].message_type = common_dot_originator__pb2._ORIGINATOR
_LISTRESPONSEITEM.fields_by_name['originator'].message_type = common_dot_originator__pb2._ORIGINATOR
_LISTRESPONSE.fields_by_name['results'].message_type = _LISTRESPONSEITEM
_REGISTERTYPEREQUEST.fields_by_name['spec'].message_type = crudstore_dot_schema__pb2._CRUDENTITYSPEC
_GETTYPERESPONSE.fields_by_name['spec'].message_type = crudstore_dot_schema__pb2._CRUDENTITYSPEC
_UPDATETYPEREQUEST.fields_by_name['spec'].message_type = crudstore_dot_schema__pb2._CRUDENTITYSPEC
_LISTTYPESRESPONSE.fields_by_name['results'].message_type = crudstore_dot_schema__pb2._CRUDENTITYSPEC
DESCRIPTOR.message_types_by_name['CreateRequest'] = _CREATEREQUEST
DESCRIPTOR.message_types_by_name['CreateResponse'] = _CREATERESPONSE
DESCRIPTOR.message_types_by_name['UpdateRequest'] = _UPDATEREQUEST
DESCRIPTOR.message_types_by_name['UpdateResponse'] = _UPDATERESPONSE
DESCRIPTOR.message_types_by_name['DeleteRequest'] = _DELETEREQUEST
DESCRIPTOR.message_types_by_name['DeleteResponse'] = _DELETERESPONSE
DESCRIPTOR.message_types_by_name['GetRequest'] = _GETREQUEST
DESCRIPTOR.message_types_by_name['GetResponse'] = _GETRESPONSE
DESCRIPTOR.message_types_by_name['ListRequest'] = _LISTREQUEST
DESCRIPTOR.message_types_by_name['ListResponseItem'] = _LISTRESPONSEITEM
DESCRIPTOR.message_types_by_name['ListResponse'] = _LISTRESPONSE
DESCRIPTOR.message_types_by_name['RegisterTypeRequest'] = _REGISTERTYPEREQUEST
DESCRIPTOR.message_types_by_name['RegisterTypeResponse'] = _REGISTERTYPERESPONSE
DESCRIPTOR.message_types_by_name['GetTypeRequest'] = _GETTYPEREQUEST
DESCRIPTOR.message_types_by_name['GetTypeResponse'] = _GETTYPERESPONSE
DESCRIPTOR.message_types_by_name['UpdateTypeRequest'] = _UPDATETYPEREQUEST
DESCRIPTOR.message_types_by_name['UpdateTypeResponse'] = _UPDATETYPERESPONSE
DESCRIPTOR.message_types_by_name['ListTypesRequest'] = _LISTTYPESREQUEST
DESCRIPTOR.message_types_by_name['ListTypesResponse'] = _LISTTYPESRESPONSE
DESCRIPTOR.message_types_by_name['HealthRequest'] = _HEALTHREQUEST
DESCRIPTOR.message_types_by_name['HealthResponse'] = _HEALTHRESPONSE
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

CreateRequest = _reflection.GeneratedProtocolMessageType('CreateRequest', (_message.Message,), dict(
  DESCRIPTOR = _CREATEREQUEST,
  __module__ = 'crudstore.service_pb2'
  # @@protoc_insertion_point(class_scope:contracts.crudstore.CreateRequest)
  ))
_sym_db.RegisterMessage(CreateRequest)

CreateResponse = _reflection.GeneratedProtocolMessageType('CreateResponse', (_message.Message,), dict(
  DESCRIPTOR = _CREATERESPONSE,
  __module__ = 'crudstore.service_pb2'
  # @@protoc_insertion_point(class_scope:contracts.crudstore.CreateResponse)
  ))
_sym_db.RegisterMessage(CreateResponse)

UpdateRequest = _reflection.GeneratedProtocolMessageType('UpdateRequest', (_message.Message,), dict(
  DESCRIPTOR = _UPDATEREQUEST,
  __module__ = 'crudstore.service_pb2'
  # @@protoc_insertion_point(class_scope:contracts.crudstore.UpdateRequest)
  ))
_sym_db.RegisterMessage(UpdateRequest)

UpdateResponse = _reflection.GeneratedProtocolMessageType('UpdateResponse', (_message.Message,), dict(
  DESCRIPTOR = _UPDATERESPONSE,
  __module__ = 'crudstore.service_pb2'
  # @@protoc_insertion_point(class_scope:contracts.crudstore.UpdateResponse)
  ))
_sym_db.RegisterMessage(UpdateResponse)

DeleteRequest = _reflection.GeneratedProtocolMessageType('DeleteRequest', (_message.Message,), dict(
  DESCRIPTOR = _DELETEREQUEST,
  __module__ = 'crudstore.service_pb2'
  # @@protoc_insertion_point(class_scope:contracts.crudstore.DeleteRequest)
  ))
_sym_db.RegisterMessage(DeleteRequest)

DeleteResponse = _reflection.GeneratedProtocolMessageType('DeleteResponse', (_message.Message,), dict(
  DESCRIPTOR = _DELETERESPONSE,
  __module__ = 'crudstore.service_pb2'
  # @@protoc_insertion_point(class_scope:contracts.crudstore.DeleteResponse)
  ))
_sym_db.RegisterMessage(DeleteResponse)

GetRequest = _reflection.GeneratedProtocolMessageType('GetRequest', (_message.Message,), dict(
  DESCRIPTOR = _GETREQUEST,
  __module__ = 'crudstore.service_pb2'
  # @@protoc_insertion_point(class_scope:contracts.crudstore.GetRequest)
  ))
_sym_db.RegisterMessage(GetRequest)

GetResponse = _reflection.GeneratedProtocolMessageType('GetResponse', (_message.Message,), dict(
  DESCRIPTOR = _GETRESPONSE,
  __module__ = 'crudstore.service_pb2'
  # @@protoc_insertion_point(class_scope:contracts.crudstore.GetResponse)
  ))
_sym_db.RegisterMessage(GetResponse)

ListRequest = _reflection.GeneratedProtocolMessageType('ListRequest', (_message.Message,), dict(
  DESCRIPTOR = _LISTREQUEST,
  __module__ = 'crudstore.service_pb2'
  # @@protoc_insertion_point(class_scope:contracts.crudstore.ListRequest)
  ))
_sym_db.RegisterMessage(ListRequest)

ListResponseItem = _reflection.GeneratedProtocolMessageType('ListResponseItem', (_message.Message,), dict(
  DESCRIPTOR = _LISTRESPONSEITEM,
  __module__ = 'crudstore.service_pb2'
  # @@protoc_insertion_point(class_scope:contracts.crudstore.ListResponseItem)
  ))
_sym_db.RegisterMessage(ListResponseItem)

ListResponse = _reflection.GeneratedProtocolMessageType('ListResponse', (_message.Message,), dict(
  DESCRIPTOR = _LISTRESPONSE,
  __module__ = 'crudstore.service_pb2'
  # @@protoc_insertion_point(class_scope:contracts.crudstore.ListResponse)
  ))
_sym_db.RegisterMessage(ListResponse)

RegisterTypeRequest = _reflection.GeneratedProtocolMessageType('RegisterTypeRequest', (_message.Message,), dict(
  DESCRIPTOR = _REGISTERTYPEREQUEST,
  __module__ = 'crudstore.service_pb2'
  # @@protoc_insertion_point(class_scope:contracts.crudstore.RegisterTypeRequest)
  ))
_sym_db.RegisterMessage(RegisterTypeRequest)

RegisterTypeResponse = _reflection.GeneratedProtocolMessageType('RegisterTypeResponse', (_message.Message,), dict(
  DESCRIPTOR = _REGISTERTYPERESPONSE,
  __module__ = 'crudstore.service_pb2'
  # @@protoc_insertion_point(class_scope:contracts.crudstore.RegisterTypeResponse)
  ))
_sym_db.RegisterMessage(RegisterTypeResponse)

GetTypeRequest = _reflection.GeneratedProtocolMessageType('GetTypeRequest', (_message.Message,), dict(
  DESCRIPTOR = _GETTYPEREQUEST,
  __module__ = 'crudstore.service_pb2'
  # @@protoc_insertion_point(class_scope:contracts.crudstore.GetTypeRequest)
  ))
_sym_db.RegisterMessage(GetTypeRequest)

GetTypeResponse = _reflection.GeneratedProtocolMessageType('GetTypeResponse', (_message.Message,), dict(
  DESCRIPTOR = _GETTYPERESPONSE,
  __module__ = 'crudstore.service_pb2'
  # @@protoc_insertion_point(class_scope:contracts.crudstore.GetTypeResponse)
  ))
_sym_db.RegisterMessage(GetTypeResponse)

UpdateTypeRequest = _reflection.GeneratedProtocolMessageType('UpdateTypeRequest', (_message.Message,), dict(
  DESCRIPTOR = _UPDATETYPEREQUEST,
  __module__ = 'crudstore.service_pb2'
  # @@protoc_insertion_point(class_scope:contracts.crudstore.UpdateTypeRequest)
  ))
_sym_db.RegisterMessage(UpdateTypeRequest)

UpdateTypeResponse = _reflection.GeneratedProtocolMessageType('UpdateTypeResponse', (_message.Message,), dict(
  DESCRIPTOR = _UPDATETYPERESPONSE,
  __module__ = 'crudstore.service_pb2'
  # @@protoc_insertion_point(class_scope:contracts.crudstore.UpdateTypeResponse)
  ))
_sym_db.RegisterMessage(UpdateTypeResponse)

ListTypesRequest = _reflection.GeneratedProtocolMessageType('ListTypesRequest', (_message.Message,), dict(
  DESCRIPTOR = _LISTTYPESREQUEST,
  __module__ = 'crudstore.service_pb2'
  # @@protoc_insertion_point(class_scope:contracts.crudstore.ListTypesRequest)
  ))
_sym_db.RegisterMessage(ListTypesRequest)

ListTypesResponse = _reflection.GeneratedProtocolMessageType('ListTypesResponse', (_message.Message,), dict(
  DESCRIPTOR = _LISTTYPESRESPONSE,
  __module__ = 'crudstore.service_pb2'
  # @@protoc_insertion_point(class_scope:contracts.crudstore.ListTypesResponse)
  ))
_sym_db.RegisterMessage(ListTypesResponse)

HealthRequest = _reflection.GeneratedProtocolMessageType('HealthRequest', (_message.Message,), dict(
  DESCRIPTOR = _HEALTHREQUEST,
  __module__ = 'crudstore.service_pb2'
  # @@protoc_insertion_point(class_scope:contracts.crudstore.HealthRequest)
  ))
_sym_db.RegisterMessage(HealthRequest)

HealthResponse = _reflection.GeneratedProtocolMessageType('HealthResponse', (_message.Message,), dict(
  DESCRIPTOR = _HEALTHRESPONSE,
  __module__ = 'crudstore.service_pb2'
  # @@protoc_insertion_point(class_scope:contracts.crudstore.HealthResponse)
  ))
_sym_db.RegisterMessage(HealthResponse)


DESCRIPTOR._options = None

_CRUDSTORESERVICE = _descriptor.ServiceDescriptor(
  name='CrudStoreService',
  full_name='contracts.crudstore.CrudStoreService',
  file=DESCRIPTOR,
  index=0,
  serialized_options=None,
  serialized_start=1599,
  serialized_end=2512,
  methods=[
  _descriptor.MethodDescriptor(
    name='Healtz',
    full_name='contracts.crudstore.CrudStoreService.Healtz',
    index=0,
    containing_service=None,
    input_type=_HEALTHREQUEST,
    output_type=_HEALTHRESPONSE,
    serialized_options=_b('\202\323\344\223\002\014\022\n/v1/healtz'),
  ),
  _descriptor.MethodDescriptor(
    name='Create',
    full_name='contracts.crudstore.CrudStoreService.Create',
    index=1,
    containing_service=None,
    input_type=_CREATEREQUEST,
    output_type=_CREATERESPONSE,
    serialized_options=None,
  ),
  _descriptor.MethodDescriptor(
    name='Update',
    full_name='contracts.crudstore.CrudStoreService.Update',
    index=2,
    containing_service=None,
    input_type=_UPDATEREQUEST,
    output_type=_UPDATERESPONSE,
    serialized_options=None,
  ),
  _descriptor.MethodDescriptor(
    name='Delete',
    full_name='contracts.crudstore.CrudStoreService.Delete',
    index=3,
    containing_service=None,
    input_type=_DELETEREQUEST,
    output_type=_DELETERESPONSE,
    serialized_options=None,
  ),
  _descriptor.MethodDescriptor(
    name='Get',
    full_name='contracts.crudstore.CrudStoreService.Get',
    index=4,
    containing_service=None,
    input_type=_GETREQUEST,
    output_type=_GETRESPONSE,
    serialized_options=None,
  ),
  _descriptor.MethodDescriptor(
    name='List',
    full_name='contracts.crudstore.CrudStoreService.List',
    index=5,
    containing_service=None,
    input_type=_LISTREQUEST,
    output_type=_LISTRESPONSE,
    serialized_options=None,
  ),
  _descriptor.MethodDescriptor(
    name='RegisterType',
    full_name='contracts.crudstore.CrudStoreService.RegisterType',
    index=6,
    containing_service=None,
    input_type=_REGISTERTYPEREQUEST,
    output_type=_REGISTERTYPERESPONSE,
    serialized_options=None,
  ),
  _descriptor.MethodDescriptor(
    name='GetType',
    full_name='contracts.crudstore.CrudStoreService.GetType',
    index=7,
    containing_service=None,
    input_type=_GETTYPEREQUEST,
    output_type=_GETTYPERESPONSE,
    serialized_options=None,
  ),
  _descriptor.MethodDescriptor(
    name='UpdateType',
    full_name='contracts.crudstore.CrudStoreService.UpdateType',
    index=8,
    containing_service=None,
    input_type=_UPDATETYPEREQUEST,
    output_type=_UPDATETYPERESPONSE,
    serialized_options=None,
  ),
  _descriptor.MethodDescriptor(
    name='ListTypes',
    full_name='contracts.crudstore.CrudStoreService.ListTypes',
    index=9,
    containing_service=None,
    input_type=_LISTTYPESREQUEST,
    output_type=_LISTTYPESRESPONSE,
    serialized_options=None,
  ),
])
_sym_db.RegisterServiceDescriptor(_CRUDSTORESERVICE)

DESCRIPTOR.services_by_name['CrudStoreService'] = _CRUDSTORESERVICE

# @@protoc_insertion_point(module_scope)
