from typing import Dict
from pydantic import BaseModel


class StatusCount(BaseModel):
    pending: int
    processing: int
    completed: int


class Metrics(BaseModel):
    total_processed: int
    average_processing_time: float  # in seconds
    status_counts: Dict[str, int] 