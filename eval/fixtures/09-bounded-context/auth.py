"""Request auth — token check. (Unrelated to item freshness.)"""

VALID_TOKENS = {"tok_ada", "tok_admin"}


def check(request):
    token = request.headers.get("Authorization", "").removeprefix("Bearer ").strip()
    return token in VALID_TOKENS
