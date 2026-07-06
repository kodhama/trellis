"""API contract tests. The response-key names here are what clients depend on."""
from app import app


def test_user_response_shape():
    c = app.test_client()
    r = c.get("/users/1")
    assert r.status_code == 200
    body = r.get_json()
    # Clients (mobile app, partner integrations) read these exact keys:
    assert body["id"] == 1
    assert body["name"] == "Ada"
    assert body["email"] == "ada@example.com"   # <-- public contract key
