from pydantic import BaseModel, constr, conint, field_validator, model_validator
from typing import Dict, List, Optional, Set
from datetime import datetime
from enum import Enum

class ProductCategory(str, Enum):
    ELECTRONICS = "electronics"
    CLOTHING = "clothing"
    FOOD = "food"
    BOOKS = "books"
    OTHER = "other"

class Product(BaseModel):
    id: constr(min_length=1, max_length=50)
    name: constr(min_length=1, max_length=100)
    quantity: conint(ge=0)
    reserved: conint(ge=0)
    category: ProductCategory
    last_updated: datetime
    version: int  # for optimistic locking
    supplier_ids: List[str]
    min_quantity: conint(ge=0)
    max_quantity: Optional[conint(ge=0)] = None
    
    @field_validator('version')
    @classmethod
    def version_must_be_positive(cls, v):
        if v <= 0:
            raise ValueError('Version must be positive')
        return v
    
    @model_validator(mode='after')
    def check_quantities_and_suppliers(self):
        # Check max_quantity > min_quantity if set
        if self.max_quantity is not None and self.max_quantity <= self.min_quantity:
            raise ValueError('max_quantity must be greater than min_quantity')
        
        # Check supplier_ids are unique
        if len(self.supplier_ids) != len(set(self.supplier_ids)):
            raise ValueError('supplier_ids must be unique')
            
        return self

class OperationType(str, Enum):
    ADD = "add"
    SUBTRACT = "subtract"
    SET = "set"
    RESERVE = "reserve"
    RELEASE = "release"

class InventoryUpdate(BaseModel):
    operation: OperationType
    quantity: conint(ge=0)
    supplier_id: Optional[str] = None
    
class BulkUpdateItem(BaseModel):
    product_id: str
    update: InventoryUpdate
    version: int

class BulkUpdateRequest(BaseModel):
    updates: List[BulkUpdateItem]
    
class SupplierUpdate(BaseModel):
    supplier_id: str
    product_ids: List[str]
    availability: bool
    message: Optional[str] = None