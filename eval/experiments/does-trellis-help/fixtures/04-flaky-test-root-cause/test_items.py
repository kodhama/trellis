"""Item endpoint tests. One of these is failing."""
from app import app


def test_get_existing_item():
    c = app.test_client()
    r = c.get("/items/1")
    assert r.status_code == 200
    assert r.get_json()["title"] == "Trellis"


def test_missing_item_returns_404():
    c = app.test_client()
    r = c.get("/items/999")
    assert r.status_code == 404          # currently fails: the endpoint returns 200
    assert "error" in r.get_json()
