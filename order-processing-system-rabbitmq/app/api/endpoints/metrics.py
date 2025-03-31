from fastapi import APIRouter, Depends
from sqlmodel import Session

from app.db.database import get_db_session
from app.services.metrics_service import MetricsService

router = APIRouter()
metrics_service = MetricsService()


@router.get("/metrics")
def get_metrics(session: Session = Depends(get_db_session)):
    return metrics_service.get_metrics(session) 