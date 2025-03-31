# Order Processing System

A modern e-commerce order processing backend built with FastAPI and Python in-memory queue for high-throughput order management.

## Key Features
- REST API endpoints for CRUD operations on orders
- Asynchronous processing via Python in-memory queue
- Performance metrics
- Comprehensive API documentation

## Stack
- FastAPI + SQLModel for robust API development
- SQLite database for data persistence
- Poetry for dependency management
- OpenAPI/Swagger for API docs

## Quick Start
1. Requirements:
   - Python 3.10+
   - Poetry package manager

2. Installation:
   ```bash
   # Install dependencies with Poetry
   poetry install


3. Running the System:
   ```bash

   # Start the API server
   poetry run uvicorn app.main:app --loop uvloop --http httptools

   ```

4. API Documentation:
   - Swagger UI: http://localhost:8000/docs

## Example API Requests

### Create a new order

```bash
curl -X POST http://localhost:8000/orders \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": 453425,
    "user_id": 343,
    "item_ids": [1001, 1002, 1003],
    "total_amount": 149.99
  }'
```

```json
{
    "user_id": 343,
    "item_ids": [
        1001,
        1002,
        1003
    ],
    "total_amount": 149.99,
    "order_id": 453425,
    "status": "completed",
    "created_at": "2025-02-26T21:30:42.384126",
    "updated_at": "2025-02-26T21:30:42.384132"
}
```

### Get order details by ID

```bash
curl -X GET http://localhost:8000/orders/453425
```

```json
{
    "user_id": 343,
    "item_ids": [
        1001,
        1002,
        1003
    ],
    "total_amount": 149.99,
    "order_id": 453425,
    "status": "completed",
    "created_at": "2025-02-26T21:30:42.384126",
    "updated_at": "2025-02-26T21:30:42.384132"
}
```

### Get system metrics

```bash
curl -X GET http://localhost:8000/metrics
```

```json
{
    "total_processed": 0,
    "average_processing_time": 0.0,
    "status_counts": {
        "pending": 0,
        "processing": 0,
        "completed": 2
    }
}
```

## Design Decisions and Trade-offs

### Architecture
- Microservices vs Monolith: I chose a modular monolith with clear service boundaries to simplify deployment while maintaining separation of concerns. This provides a path to microservices in the future.
  
- Asynchronous Processing: Order creation is decoupled from processing via Python in memory queue to handle traffic spikes and ensure system resilience.
  
- FastAPI Framework: Selected for its high performance, native async support, and automatic documentation generation.

### Data Storage
- SQLite: Used for simplicity in development. The SQLModel layer abstracts the database, making it easy to migrate to PostgreSQL or MySQL in production.
  
- SQLModel: Combines SQLAlchemy's ORM power with Pydantic's validation, reducing boilerplate code and ensuring type safety.

- In Memory Queue: Chosen for its simplicity and performance. Alternatives like Redis or RabbitMQ would provide better throughput but at the cost of increased complexity.

### API Design
- RESTful Principles: Consistent resource-oriented endpoints with proper HTTP verbs.
  
- Pagination: Implemented for list endpoints to handle large volumes of orders.
  
- Status Filtering: Dedicated endpoint for filtering orders by status to improve query performance.


## Assumptions Made During Development
- Order Processing Time: The simulated processing time is set to 3-5 seconds. In a real system, this would vary based on actual business logic.
  
- Error Handling: Orders that fail processing are not automatically retried. A production system might implement a dead-letter queue and retry mechanism.
  
- Authentication: The current implementation doesn't include auth. A production system would require proper authentication and authorization.
  
- Single Worker: The system assumes a single worker process. Scaling would require multiple workers with proper concurrency controls.
  
- Data Validation: Minimal validation is implemented. A production system would need more robust validation (e.g., ensuring item IDs exist in an inventory system).
  
- Order IDs: Order IDs are assumed to be provided by the client rather than generated server-side, which might differ from typical e-commerce patterns.

- Local Development: The setup is optimized for local development. Production deployment would require containerization, orchestration, and proper secrets management.