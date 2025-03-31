import aio_pika
import json
import os
from typing import Dict, Any


class RabbitMQProducer:
    def __init__(self):
        self.connection = None
        self.channel = None
        self.exchange = None
        self.queue_name = "orders"
        self.routing_key = "order.created"
        self.rabbitmq_url = os.getenv("RABBITMQ_URL", "amqp://guest:guest@localhost/")

    async def connect(self):
        if self.connection is None or self.connection.is_closed:
            self.connection = await aio_pika.connect_robust(self.rabbitmq_url)
            self.channel = await self.connection.channel()
            self.exchange = await self.channel.declare_exchange(
                "orders", aio_pika.ExchangeType.DIRECT, durable=True
            )
            queue = await self.channel.declare_queue(
                self.queue_name, durable=True
            )
            await queue.bind(self.exchange, self.routing_key)

    async def publish(self, message: Dict[str, Any]):
        await self.connect()
        await self.exchange.publish(
            aio_pika.Message(
                body=json.dumps(message).encode(),
                delivery_mode=aio_pika.DeliveryMode.PERSISTENT,
            ),
            routing_key=self.routing_key,
        )

    async def close(self):
        if self.connection and not self.connection.is_closed:
            await self.connection.close() 