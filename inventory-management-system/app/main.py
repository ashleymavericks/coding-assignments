import uuid
import asyncio
import time
from typing import Dict, Optional
from fastapi import FastAPI, Depends, Request, BackgroundTasks, status
from fastapi.responses import JSONResponse, StreamingResponse
from fastapi.middleware.cors import CORSMiddleware

from app.models import Product, InventoryUpdate, BulkUpdateRequest, SupplierUpdate
from app.exceptions import ProductNotFound, RateLimitExceeded, InventoryError, BulkOperationError, CircuitBreakerOpen
from app.services.inventory_manager import InventoryManager

app = FastAPI(title="Inventory Management System")

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Initialize inventory manager
inventory_manager = InventoryManager()

# Helper function to get client ID from request
async def get_client_id(request: Request) -> str:
    return request.headers.get("X-Client-ID", "anonymous")

# Middleware for request tracing
@app.middleware("http")
async def trace_requests(request: Request, call_next):
    # Generate a unique request ID
    request_id = str(uuid.uuid4())
    request.state.request_id = request_id
    
    # Track request timing
    start_time = time.time()
    
    # Process the request
    response = await call_next(request)
    
    # Calculate processing time
    process_time = time.time() - start_time
    
    # Add custom headers to response
    response.headers["X-Request-ID"] = request_id
    response.headers["X-Process-Time"] = str(process_time)
    
    # Log trace info
    await inventory_manager.trace_request(
        request_id, 
        f"{request.method} {request.url.path} - {response.status_code} ({process_time:.4f}s)"
    )
    
    return response

# Exception handlers
@app.exception_handler(ProductNotFound)
async def product_not_found_handler(request: Request, exc: ProductNotFound):
    return JSONResponse(
        status_code=exc.status_code,
        content={"detail": exc.detail}
    )

@app.exception_handler(RateLimitExceeded)
async def rate_limit_handler(request: Request, exc: RateLimitExceeded):
    return JSONResponse(
        status_code=exc.status_code,
        content={"detail": exc.detail}
    )

@app.exception_handler(InventoryError)
async def inventory_error_handler(request: Request, exc: InventoryError):
    return JSONResponse(
        status_code=exc.status_code,
        content={"detail": exc.detail}
    )

@app.exception_handler(BulkOperationError)
async def bulk_error_handler(request: Request, exc: BulkOperationError):
    return JSONResponse(
        status_code=exc.status_code,
        content={"detail": exc.detail}
    )

@app.exception_handler(CircuitBreakerOpen)
async def circuit_breaker_handler(request: Request, exc: CircuitBreakerOpen):
    return JSONResponse(
        status_code=exc.status_code,
        content={"detail": exc.detail}
    )

# API endpoints
@app.get("/products/{product_id}")
async def get_product(
    product_id: str,
    client_id: str = Depends(get_client_id)
):
    # Check rate limiting
    if not inventory_manager.check_rate_limit(client_id):
        raise RateLimitExceeded(client_id)
        
    # Check circuit breaker
    if not await inventory_manager.check_circuit_breaker("get_product"):
        raise CircuitBreakerOpen("get_product")
    
    try:
        product = await inventory_manager.get_product(product_id)
        inventory_manager.record_success("get_product")
        return product
    except Exception as e:
        inventory_manager.record_failure("get_product")
        raise e

@app.put("/products/{product_id}")
async def update_product(
    product_id: str,
    update: InventoryUpdate,
    version: int,
    client_id: str = Depends(get_client_id)
):
    # Check rate limiting
    if not inventory_manager.check_rate_limit(client_id):
        raise RateLimitExceeded(client_id)
        
    # Check circuit breaker
    if not await inventory_manager.check_circuit_breaker("update_product"):
        raise CircuitBreakerOpen("update_product")
    
    try:
        product = await inventory_manager.update_with_version(product_id, update, version)
        inventory_manager.record_success("update_product")
        return product
    except Exception as e:
        inventory_manager.record_failure("update_product")
        raise e

@app.post("/products/bulk")
async def bulk_update_products(
    updates: BulkUpdateRequest,
    background_tasks: BackgroundTasks,
    client_id: str = Depends(get_client_id)
):
    # Check rate limiting
    if not inventory_manager.check_rate_limit(client_id):
        raise RateLimitExceeded(client_id)
        
    # Check circuit breaker
    if not await inventory_manager.check_circuit_breaker("bulk_update"):
        raise CircuitBreakerOpen("bulk_update")
    
    try:
        results, errors = await inventory_manager.bulk_update(updates)
        
        # If there are errors, include them in response
        if errors:
            inventory_manager.record_failure("bulk_update")
            return {
                "success": results,
                "errors": errors
            }
        
        inventory_manager.record_success("bulk_update")
        return {"success": results}
    except Exception as e:
        inventory_manager.record_failure("bulk_update")
        raise e

@app.get("/products/{product_id}/stream")
async def stream_updates(
    product_id: str,
    client_id: str = Depends(get_client_id)
):
    # Check rate limiting
    if not inventory_manager.check_rate_limit(client_id):
        raise RateLimitExceeded(client_id)
    
    # Create an async generator for server-sent events
    async def event_generator():
        try:
            async for product in inventory_manager.subscribe_to_updates(product_id):
                # Format as Server-Sent Event
                data = product.model_dump_json()
                yield f"data: {data}\n\n"
                await asyncio.sleep(0.1)  # Small delay to prevent flooding
        except Exception as e:
            yield f"data: {{'error': '{str(e)}'}}\n\n"
    
    return StreamingResponse(
        event_generator(), 
        media_type="text/event-stream"
    )

@app.post("/webhook/supplier")
async def supplier_webhook(
    update: SupplierUpdate,
    background_tasks: BackgroundTasks
):
    # Process supplier update in background
    background_tasks.add_task(process_supplier_update, update)
    return {"status": "processing"}

async def process_supplier_update(update: SupplierUpdate):
    """Process supplier update in background"""
    for product_id in update.product_ids:
        try:
            product = await inventory_manager.get_product(product_id)
            
            # If supplier not available, update product
            if not update.availability and update.supplier_id in product.supplier_ids:
                # Logic to handle supplier unavailability
                pass
        except ProductNotFound:
            # Log that product wasn't found
            pass

@app.on_event("startup")
async def startup_event():
    # Initialize with test data
    product = Product(
        id="test1",
        name="Test Product",
        quantity=100,
        reserved=0,
        category="electronics",
        last_updated=time.time(),
        version=1,
        supplier_ids=["supplier1", "supplier2"],
        min_quantity=10,
        max_quantity=1000
    )
    inventory_manager.add_product(product)
    
    # Start background cleanup task
    asyncio.create_task(inventory_manager.schedule_cleanup())

@app.on_event("shutdown")
async def shutdown_event():
    # Cleanup logic would go here
    pass

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)