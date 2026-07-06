"""Bookmarks service. GET /items reads through the cache; PUT updates the store."""
from flask import Flask, jsonify, request
import cache
import db
import auth
import billing
import notifications

app = Flask(__name__)


@app.get("/items")
def list_items():
    return jsonify(cache.get_all_items())


@app.put("/items/<int:item_id>")
def update_item(item_id):
    fields = request.get_json() or {}
    item = db.update_item(item_id, **fields)
    notifications.notify_changed(item_id)
    return jsonify(item)


@app.get("/account")
def account():
    if not auth.check(request):
        return jsonify({"error": "unauthorized"}), 401
    return jsonify(billing.summary())


if __name__ == "__main__":
    cache.warm()
    app.run(debug=True)
