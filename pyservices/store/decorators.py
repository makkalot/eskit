import logging
import sys

import grpc
import traceback

from pyservices.store.crud import NotFoundError
from pyservices.store.estore import ConcurrencyError


def get_logger():
    logger = logging.getLogger()
    logger.setLevel(logging.DEBUG)
    ch = logging.StreamHandler(sys.stdout)
    ch.setLevel(logging.DEBUG)
    formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
    ch.setFormatter(formatter)
    logger.addHandler(ch)
    return logger


logger = get_logger()


def log_traceback(logger, tb, limit=None):
    extracted_list = traceback.extract_tb(tb, limit=limit)
    lines = traceback.StackSummary.from_list(extracted_list).format()
    logger.debug("".join(lines))


def grpc_catch(resp_cls):
    def wrap(fn):
        def _decorate(self, *args, **kwargs):
            context = args[-1]
            try:
                return fn(self, *args, **kwargs)
            except NotFoundError as ex:
                context.set_code(grpc.StatusCode.NOT_FOUND)
                context.set_details(str(ex))
                return resp_cls()
            except ValueError as ex:
                context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
                context.set_details(str(ex))
                return resp_cls()
            except ConcurrencyError as ex:
                context.set_code(grpc.StatusCode.FAILED_PRECONDITION)
                context.set_details(str(ex))
                return resp_cls()
            except Exception as ex:
                log_traceback(logger, ex.__traceback__)
                context.set_code(grpc.StatusCode.INTERNAL)
                context.set_details(str(ex))
                return resp_cls()

        return _decorate

    return wrap
