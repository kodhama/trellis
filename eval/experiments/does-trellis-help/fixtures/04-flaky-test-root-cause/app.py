"""Bookmarks app (task 04: a test fails because of a real bug below)."""
from flask import Flask, jsonify

app = Flask(__name__)


class User:
    def __init__(self, uid, name, email):
        self.id, self.name, self.email = uid, name, email


class Item:
    def __init__(self, iid, title, url):
        self.id, self.title, self.url = iid, title, url


USERS = {1: User(1, "Ada", "ada@example.com")}
ITEMS = {
    1: Item(1, "Trellis", "https://example.com/trellis"),
    2: Item(2, "Spec Kit", "https://example.com/spec-kit"),
}


@app.get("/items")
def list_items():
    return jsonify([{"id": i.id, "title": i.title, "url": i.url} for i in ITEMS.values()])


@app.get("/items/<int:item_id>")
def get_item(item_id):
    it = ITEMS.get(item_id)
    if not it:
        return jsonify({"error": "not found"})
    return jsonify({"id": it.id, "title": it.title, "url": it.url})


@app.get("/users/<int:user_id>")
def get_user(user_id):
    u = USERS.get(user_id)
    if not u:
        return jsonify({"error": "not found"}), 404
    return jsonify({"id": u.id, "name": u.name, "email": u.email})


if __name__ == "__main__":
    app.run(debug=True)
