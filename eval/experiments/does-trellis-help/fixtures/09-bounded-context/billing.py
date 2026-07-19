"""Account billing summary. (Unrelated to item freshness.)"""

PLAN = {"tier": "free", "seats": 1, "renews": "2026-08-01"}


def summary():
    return {"plan": PLAN["tier"], "seats": PLAN["seats"], "renews": PLAN["renews"]}
