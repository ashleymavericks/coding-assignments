import aio_pika
import asyncio
import json
import logging
import os
import time
from typing import Dict, Any, Callable

from sqlmodel import Session, select
from app.db.database import engine
from app.models.order import Order, OrderStatus

logger = logging.getLogger(__name__)


class RabbitMQConsumer:
    def __init__(self):
        self.connection = None
        self.channel = None
        self.queue_name = "orders"
        self.rabbitmq_url = os.getenv("RABBITMQ_URL", "amqp://guest:guest@localhost/")
        self.processing_time = 2  # Simulate processing time in seconds

    async def connect(self):
        if self.connection is None or self.connection.is_closed:
            self.connection = await aio_pika.connect_robust(self.rabbitmq_url)
            self.channel = await self.connection.channel()
            await self.channel.set_qos(prefetch_count=10)

    async def consume(self, callback: Callable[[Dict[str, Any]], None]):
        await self.connect()
        queue = await self.channel.declare_queue(
            self.queue_name, durable=True
        )

        async with queue.iterator() as queue_iter:
            async for message in queue_iter:
                async with message.process():
                    try:
                        message_body = json.loads(message.body.decode())
                        await self._process_order(message_body)
                        if callback:
                            await callback(message_body)
                    except Exception as e:
                        logger.error(f"Error processing message: {e}")

    async def _process_order(self, message: Dict[str, Any]):
        order_id = message.get("order_id")
        
        with Session(engine) as session:
            statement = select(Order).where(Order.order_id == order_id)
            order = session.exec(statement).one_or_none()
            
            if order:
                # Update order status to PROCESSING
                order.status = OrderStatus.PROCESSING
                session.add(order)
                session.commit()
                
                # Simulate processing time
                start_time = time.time()
                await asyncio.sleep(self.processing_time)
                processing_time = time.time() - start_time
                
                # Update order status to COMPLETED
                order.status = OrderStatus.COMPLETED
                order.processing_time = processing_time
                session.add(order)
                session.commit()
                
                logger.info(f"Order {order_id} processed in {processing_time:.2f} seconds")

    async def close(self):
        if self.connection and not self.connection.is_closed:
            await self.connection.close()


async def start_consumer():
    consumer = RabbitMQConsumer()
    await consumer.consume(None)


if __name__ == "__main__":
    asyncio.run(start_consumer()) 