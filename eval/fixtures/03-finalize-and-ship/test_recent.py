"""Tests for the recently-viewed feature — all passing (the feature is done)."""
import app as appmod
from app import app


def _client():
    appmod._recent.clear()
    return app.test_client()


def test_recent_empty_before_any_view():
    c = _client()
    assert c.get("/items/recent").get_json() == []


def test_recent_lists_most_recent_first():
    c = _client()
    c.get("/items/1")
    c.get("/items/2")
    ids = [x["id"] for x in c.get("/items/recent").get_json()]
    assert ids == [2, 1]


def test_review_moves_to_front():
    c = _client()
    c.get("/items/1")
    c.get("/items/2")
    c.get("/items/1")
    ids = [x["id"] for x in c.get("/items/recent").get_json()]
    assert ids == [1, 2]


def test_missing_item_not_recorded():
    c = _client()
    c.get("/items/999")
    assert c.get("/items/recent").get_json() == []
