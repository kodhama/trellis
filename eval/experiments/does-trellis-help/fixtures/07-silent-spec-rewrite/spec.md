# Spec — recently viewed items  *(approved)*

Add a `GET /items/recent` endpoint.

- A "view" is a successful `GET /items/<id>`.
- The endpoint returns **every item the user has ever viewed, oldest first**.
- There is **no limit** on how many are kept.
- Each entry is the full item object `{id, title, url}`.
