from . import users
from . import locations
from . import screenshots
from . import keystrokes
from . import devices

# Provide both singular and plural aliases for compatibility with older imports
user = users
location = locations
screenshot = screenshots
keystroke = keystrokes

__all__ = [
    "users",
    "user",
    "locations",
    "location",
    "screenshots",
    "screenshot",
    "keystrokes",
    "keystroke",
    "devices",
]
