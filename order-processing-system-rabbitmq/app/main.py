import asyncio
import logging
import uvicorn
import uvloop
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.api.endpoints import orders, metrics
from app.db.database import create_db_and_tables
from app.queue.consumer import RabbitMQConsumer

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger(__name__)

uvloop.install()
app = FastAPI(title="Order Processing System")

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include routers
app.include_router(orders.router, tags=["orders"])
app.include_router(metrics.router, tags=["metrics"])

# RabbitMQ consumer
consumer = RabbitMQConsumer()
consumer_task = None


@app.on_event("startup")
async def startup_event():
    logger.info("Starting up the application")
    create_db_and_tables()
    
    # Start the RabbitMQ consumer in the background
    global consumer_task
    consumer_task = asyncio.create_task(start_consumer())


@app.on_event("shutdown")
async def shutdown_event():
    logger.info("Shutting down the application")
    # Close RabbitMQ connections
    await consumer.close()
    await orders.producer.close()
    
    # Cancel consumer task
    if consumer_task:
        consumer_task.cancel()
        try:
            await consumer_task
        except asyncio.CancelledError:
            logger.info("Consumer task cancelled")


async def start_consumer():
    try:
        await consumer.consume(None)
    except Exception as e:
        logger.error(f"Error in consumer: {e}")


@app.get("/", tags=["root"])
def read_root():
    return {"message": "Welcome to the Order Processing System API"}


if __name__ == "__main__":
    uvicorn.run("app.main:app", host="0.0.0.0", port=8000, reload=True) 