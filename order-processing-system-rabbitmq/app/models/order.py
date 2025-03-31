from datetime import datetime
from enum import Enum
from typing import List, Optional

from sqlmodel import Field, SQLModel, JSON, Column


class OrderStatus(str, Enum):
    PENDING = "PENDING"
    PROCESSING = "PROCESSING"
    COMPLETED = "COMPLETED"


class OrderBase(SQLModel):
    user_id: int
    item_ids: List[int] = Field(sa_column=Column(JSON))
    total_amount: float


class Order(OrderBase, table=True):
    id: Optional[int] = Field(default=None, primary_key=True)
    order_id: str = Field(index=True, unique=True)
    status: OrderStatus = Field(default=OrderStatus.PENDING)
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)
    processing_time: Optional[float] = None


class OrderCreate(OrderBase):
    order_id: str


class OrderRead(OrderBase):
    id: int
    order_id: str
    status: OrderStatus
    created_at: datetime
    updated_at: datetime


class OrderUpdate(SQLModel):
    status: Optional[OrderStatus] = None
    processing_time: Optional[float] = None 