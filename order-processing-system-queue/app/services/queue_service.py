import queue
import time
import logging
import threading
from datetime import datetime
from typing import Dict, Any

from sqlmodel import Session, select
from app.db.models import Order, OrderStatus
from app.db.database import engine

# Setup logger
logger = logging.getLogger(__name__)

# Global in-memory queue for order processing
order_queue = queue.Queue()

# Global metrics for monitoring
metrics = {
    "total_processed": 0,
    "processing_times": [],
    "status_counts": {
        OrderStatus.PENDING.value: 0,
        OrderStatus.PROCESSING.value: 0,
        OrderStatus.COMPLETED.value: 0
    }
}


def update_metrics():
    with Session(engine) as session:
        # Count orders by status
        for status in OrderStatus:
            orders = session.exec(
                select(Order).where(Order.status == status)
            ).all()
            metrics["status_counts"][status.value] = len(orders)


def get_metrics() -> Dict[str, Any]:
    update_metrics()
    
    # Calculate average processing time
    avg_processing_time = 0
    if metrics["processing_times"]:
        avg_processing_time = sum(metrics["processing_times"]) / len(metrics["processing_times"])
    
    return {
        "total_processed": metrics["total_processed"],
        "average_processing_time": round(avg_processing_time, 2),
        "status_counts": metrics["status_counts"]
    }


def process_order(order_id: int, session: Session) -> None:
    try:
        # Find the order in the database
        order = session.get(Order, order_id)
        if not order:
            logger.error(f"Order {order_id} not found")
            return

        # Update to processing status
        order.status = OrderStatus.PROCESSING
        order.processing_started_at = datetime.utcnow()
        session.add(order)
        session.commit()
        
        # Simulate processing time (3-5 seconds)
        processing_time = 3 + (order_id % 3)
        time.sleep(processing_time)
        
        # Update to completed status
        order.status = OrderStatus.COMPLETED
        order.processing_completed_at = datetime.utcnow()
        session.add(order)
        session.commit()
        
        # Update metrics
        metrics["total_processed"] += 1
        processing_duration = (order.processing_completed_at - order.processing_started_at).total_seconds()
        metrics["processing_times"].append(processing_duration)
        
        logger.info(f"Order {order_id} processed successfully in {processing_duration:.2f} seconds")
    
    except Exception as e:
        logger.error(f"Error processing order {order_id}: {str(e)}")
        session.rollback()


def order_processor_worker():
    from app.db.database import engine
    
    logger.info("Order processor worker started")
    
    while True:
        try:
            # Get order_id from queue
            order_id = order_queue.get(block=True)
            
            # Process the order within a new session
            with Session(engine) as session:
                process_order(order_id, session)
            
            # Mark task as done
            order_queue.task_done()
        
        except Exception as e:
            logger.error(f"Error in order processor worker: {str(e)}")


def start_order_processor():
    worker_thread = threading.Thread(
        target=order_processor_worker,
        daemon=True 
    )
    worker_thread.start()
    logger.info("Order processor started")


def enqueue_order(order_id: int) -> None:
    order_queue.put(order_id)
    logger.info(f"Order {order_id} added to processing queue") 