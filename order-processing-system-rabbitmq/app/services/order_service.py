import uuid
from typing import List, Dict, Any, Optional
from datetime import datetime

from sqlmodel import Session, select
from app.models.order import Order, OrderCreate, OrderStatus, OrderUpdate


class OrderService:
    def create_order(self, session: Session, order: OrderCreate) -> Order:
        db_order = Order.from_orm(order)
        session.add(db_order)
        session.commit()
        session.refresh(db_order)
        return db_order

    def get_order(self, session: Session, order_id: str) -> Optional[Order]:
        statement = select(Order).where(Order.order_id == order_id)
        return session.exec(statement).one_or_none()

    def get_orders(self, session: Session, skip: int = 0, limit: int = 100) -> List[Order]:
        statement = select(Order).offset(skip).limit(limit)
        return session.exec(statement).all()

    def update_order(self, session: Session, order_id: str, order_update: OrderUpdate) -> Optional[Order]:
        db_order = self.get_order(session, order_id)
        if not db_order:
            return None
            
        order_data = order_update.dict(exclude_unset=True)
        for key, value in order_data.items():
            setattr(db_order, key, value)
            
        db_order.updated_at = datetime.utcnow()
        session.add(db_order)
        session.commit()
        session.refresh(db_order)
        return db_order

    def get_orders_by_status(self, session: Session, status: OrderStatus) -> List[Order]:
        statement = select(Order).where(Order.status == status)
        return session.exec(statement).all()

    def generate_order_id(self) -> str:
        return f"ORD-{uuid.uuid4().hex[:8].upper()}" 