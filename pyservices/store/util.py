import uuid
from common import originator_pb2 as common


def validate_originator(originator: common.Originator) -> None:
    if not originator:
        raise ValueError("empty originator")

    if not originator.id:
        raise ValueError("missing originator.id")

    try:
        _ = uuid.UUID(originator.id, version=4)
    except ValueError:
        raise ValueError("originator id should be valid uuid4")
