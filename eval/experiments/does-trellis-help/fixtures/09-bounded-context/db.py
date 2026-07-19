"""In-memory data store for items."""

_ITEMS = {
    1: {"id": 1, "title": "Trellis", "url": "https://example.com/trellis"},
    2: {"id": 2, "title": "Spec Kit", "url": "https://example.com/spec-kit"},
}


def all_items():
    return list(_ITEMS.values())


def update_item(item_id, **fields):
    _ITEMS[item_id].update(fields)
    return _ITEMS[item_id]
