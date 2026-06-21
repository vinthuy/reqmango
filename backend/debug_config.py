import sys
sys.path.insert(0, '.')
from app.core.config import settings
print(f"DATABASE_URL: {settings.DATABASE_URL}")
