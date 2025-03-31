from datetime import datetime
from typing import List, Optional, Dict, Any

from sqlmodel import Session, select
from app.db.models import Order, OrderCreate, OrderStatus
from app.services.queue_service import enqueue_order


async def create_order(order: OrderCreate, session: Session) -> Order:
    # Create order object
    db_order = Order.from_orm(order)
    db_order.created_at = datetime.utcnow()
    db_order.updated_at = datetime.utcnow()
    
    # Save to database
    session.add(db_order)
    session.commit()
    session.refresh(db_order)
    
    # Add to processing queue
    enqueue_order(db_order.order_id)
    
    return db_order


async def get_order(order_id: int, session: Session) -> Optional[Order]:
    return session.get(Order, order_id)


async def get_orders(
    skip: int = 0, 
    limit: int = 100, 
    status: Optional[OrderStatus] = None,
    session: Session = None
) -> List[Order]:
    query = select(Order)
    
    if status:
        query = query.where(Order.status == status)
    
    query = query.offset(skip).limit(limit)
    return session.exec(query).all()


async def update_order_status(
    order_id: int, 
    status: OrderStatus, 
    session: Session
) -> Optional[Order]:
    db_order = session.get(Order, order_id)
    if not db_order:
        return None
        
    db_order.status = status
    db_order.updated_at = datetime.utcnow()
    
    if status == OrderStatus.PROCESSING and not db_order.processing_started_at:
        db_order.processing_started_at = datetime.utcnow()
    
    if status == OrderStatus.COMPLETED and not db_order.processing_completed_at:
        db_order.processing_completed_at = datetime.utcnow()
    
    session.add(db_order)
    session.commit()
    session.refresh(db_order)
    
    return db_order 