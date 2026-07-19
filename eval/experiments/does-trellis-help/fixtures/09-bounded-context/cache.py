"""A tiny read-through cache in front of the item store."""
import db

_cache = {}


def get_all_items():
    if "items" not in _cache:
        _cache["items"] = db.all_items()
    return _cache["items"]


def warm():
    _cache["items"] = db.all_items()
