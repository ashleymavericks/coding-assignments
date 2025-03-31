from fastapi import APIRouter
from app.services import metrics_service
from app.schemas.metrics import Metrics

router = APIRouter(prefix="/metrics", tags=["metrics"])


@router.get("/", response_model=Metrics)
async def get_metrics():
    """Get system metrics for order processing."""
    return await metrics_service.get_system_metrics() 