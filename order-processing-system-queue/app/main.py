import logging
import uvloop
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.db.database import create_db_and_tables
from app.api.routes import orders, metrics
from app.services.queue_service import start_order_processor

logger = logging.getLogger(__name__)

uvloop.install()
app = FastAPI(
    description="A backend system for processing orders with an in-memory queue"
)


app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # In production, replace with specific origins
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(orders.router)
app.include_router(metrics.router)


@app.on_event("startup")
async def startup_event():
    create_db_and_tables()
    logger.info("Database tables created")
    
    start_order_processor()
    logger.info("Order processor started")


@app.on_event("shutdown")
async def shutdown_event():
    logger.info("Shutting down application...")


@app.get("/")
async def root():
    return {
        "message": "Order Processing System API",
        "docs_url": "/docs",
        "health": "ok"
    }


@app.get("/health")
async def health():
    return {"status": "ok"} 