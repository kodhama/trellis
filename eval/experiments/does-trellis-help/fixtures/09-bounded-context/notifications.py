"""Fire change notifications. (Unrelated to item freshness — just logs.)"""

_log = []


def notify_changed(item_id):
    _log.append(("changed", item_id))
