# Static typing of error tag sets: a module-level error_tags call
# mints a type whose attribute table is exactly the tag set.

errors = error_tags("NotFound", "Timeout")

def find(id: int)! -> str:
    if id < 0:
        return errors.NotFound
    return "user_" + str(id)

def caller() -> str:
    return find(3) catch "guest"

def typo()! -> str:
    return errors.NotFonud

def wrong_success()! -> str:
    return 42

def constructed()! -> str:
    return errors.Timeout(message="too slow")
