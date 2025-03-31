from typing import Dict, Any
from sqlmodel import Session, select, func
from app.models.order import Order, OrderStatus


class MetricsService:
    def get_metrics(self, session: Session) -> Dict[str, Any]:
        # Total orders
        total_orders = session.exec(select(func.count()).select_from(Order)).one()
        
        # Average processing time
        avg_processing_time = session.exec(
            select(func.avg(Order.processing_time))
            .where(Order.processing_time.isnot(None))
        ).one() or 0.0
        
        # Orders by status
        pending_count = session.exec(
            select(func.count()).select_from(Order).where(Order.status == OrderStatus.PENDING)
        ).one()
        
        processing_count = session.exec(
            select(func.count()).select_from(Order).where(Order.status == OrderStatus.PROCESSING)
        ).one()
        
        completed_count = session.exec(
            select(func.count()).select_from(Order).where(Order.status == OrderStatus.COMPLETED)
        ).one()
        
        return {
            "total_orders_processed": completed_count,
            "average_processing_time": round(avg_processing_time, 2) if avg_processing_time else 0.0,
            "orders_by_status": {
                "pending": pending_count,
                "processing": processing_count,
                "completed": completed_count
            }
        } 