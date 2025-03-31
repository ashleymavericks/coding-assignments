from typing import Dict, Any
from app.services.queue_service import get_metrics


async def get_system_metrics() -> Dict[str, Any]:
    return get_metrics() 