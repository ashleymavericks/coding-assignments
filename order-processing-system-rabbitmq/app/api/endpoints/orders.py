from typing import List
from fastapi import APIRouter, Depends, HTTPException, BackgroundTasks
from sqlmodel import Session

from app.db.database import get_db_session
from app.models.order import OrderCreate, OrderRead, OrderStatus
from app.services.order_service import OrderService
from app.queue.producer import RabbitMQProducer

router = APIRouter()
order_service = OrderService()
producer = RabbitMQProducer()


@router.post("/orders", response_model=OrderRead, status_code=201)
async def create_order(
    order: OrderCreate,
    background_tasks: BackgroundTasks,
    session: Session = Depends(get_db_session)
):
    db_order = order_service.create_order(session, order)
    
    # Send order to queue for processing
    message = {
        "order_id": db_order.order_id,
        "user_id": db_order.user_id,
        "item_ids": db_order.item_ids,
        "total_amount": db_order.total_amount
    }
    
    background_tasks.add_task(producer.publish, message)
    return db_order


@router.get("/orders", response_model=List[OrderRead])
def get_orders(
    skip: int = 0, 
    limit: int = 100, 
    session: Session = Depends(get_db_session)
):
    orders = order_service.get_orders(session, skip=skip, limit=limit)
    return orders


@router.get("/orders/{order_id}", response_model=OrderRead)
def get_order(order_id: str, session: Session = Depends(get_db_session)):
    order = order_service.get_order(session, order_id)
    if not order:
        raise HTTPException(status_code=404, detail="Order not found")
    return order


@router.get("/orders/status/{status}", response_model=List[OrderRead])
def get_orders_by_status(
    status: OrderStatus, 
    session: Session = Depends(get_db_session)
):
    orders = order_service.get_orders_by_status(session, status)
    return orders 