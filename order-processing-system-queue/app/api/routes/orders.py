from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, status
from sqlmodel import Session

from app.db.database import get_db_session
from app.db.models import Order, OrderCreate, OrderRead, OrderStatus
from app.services import order_service

router = APIRouter(prefix="/orders", tags=["orders"])


@router.post("/", response_model=OrderRead, status_code=status.HTTP_201_CREATED)
async def create_order(
    order: OrderCreate,
    session: Session = Depends(get_db_session)
):
    return await order_service.create_order(order, session)


@router.get("/{order_id}", response_model=OrderRead)
async def get_order(
    order_id: int,
    session: Session = Depends(get_db_session)
):
    db_order = await order_service.get_order(order_id, session)
    if db_order is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Order with ID {order_id} not found"
        )
    return db_order


@router.get("/", response_model=List[OrderRead])
async def get_orders(
    status: Optional[OrderStatus] = None,
    skip: int = 0,
    limit: int = 100,
    session: Session = Depends(get_db_session)
):
    return await order_service.get_orders(skip, limit, status, session) 