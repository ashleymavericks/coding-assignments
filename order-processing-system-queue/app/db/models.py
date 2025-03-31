from enum import Enum
from typing import List, Optional
from datetime import datetime
from sqlmodel import Field, SQLModel, JSON, Column


class OrderStatus(str, Enum):
    PENDING = "pending"
    PROCESSING = "processing"
    COMPLETED = "completed"


class OrderBase(SQLModel):
    user_id: int
    item_ids: List[int] = Field(sa_column=Column(JSON))
    total_amount: float


class Order(OrderBase, table=True):
    __tablename__ = "orders"

    order_id: Optional[int] = Field(default=None, primary_key=True)
    status: OrderStatus = Field(default=OrderStatus.PENDING)
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)
    processing_started_at: Optional[datetime] = Field(default=None)
    processing_completed_at: Optional[datetime] = Field(default=None)


class OrderCreate(OrderBase):
    order_id: int


class OrderRead(OrderBase):
    order_id: int
    status: OrderStatus
    created_at: datetime
    updated_at: datetime


class OrderUpdate(SQLModel):
    status: Optional[OrderStatus] = None
    processing_started_at: Optional[datetime] = None
    processing_completed_at: Optional[datetime] = None 