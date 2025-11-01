from . import users
from . import locations
from . import screenshots
from . import keystrokes
from . import app_activity
from . import devices

# Provide both singular and plural aliases for compatibility with older imports
user = users
location = locations
screenshot = screenshots
keystroke = keystrokes
app_activity = app_activity

__all__ = [
    "users",
    "user",
    "locations",
    "location",
    "screenshots",
    "screenshot",
    "keystrokes",
    "keystroke",
    "app_activity",
    "devices",
]
