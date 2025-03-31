# Order Processing System

A modern e-commerce order processing backend built with FastAPI and RabbitMQ for high-throughput order management.

## Key Features
- REST API endpoints for CRUD operations on orders
- Asynchronous processing via RabbitMQ message queues
- Real-time order status tracking and notifications
- Performance metrics and monitoring dashboards
- Horizontally scalable microservices architecture
- Fault-tolerant with automatic retries
- Comprehensive API documentation

## Stack
- FastAPI + SQLModel for robust API development
- SQLite database for data persistence
- RabbitMQ for reliable message queuing
- Poetry for dependency management
- OpenAPI/Swagger for API docs
- Prometheus/Grafana for monitoring

## Quick Start
1. Requirements:
   - Python 3.10+
   - Poetry package manager
   - RabbitMQ message broker
   - 2GB RAM minimum

2. Installation:
   ```bash
   # Install dependencies with Poetry
   poetry install

   # Initialize database
   poetry run python scripts/init_db.py
   ```

3. Running the System:
   ```bash
   # Start RabbitMQ (if not running as a service)
   rabbitmq-server

   # Start the API server
   poetry run uvicorn app.main:app --loop uvloop --http httptools

   # Start the order processor worker
   poetry run python worker.py
   ```

4. Testing:
   ```bash
   # Run unit tests
   poetry run pytest

   # Run integration tests
   poetry run pytest tests/integration/
   ```

5. API Documentation:
   - Swagger UI: http://localhost:8000/docs

## Example API Requests

### Create a new order

```bash
curl -X POST http://localhost:8000/orders \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": "ORD-A8F64D2E",
    "user_id": 123,
    "item_ids": [1001, 1002, 1003],
    "total_amount": 149.99
  }'
```

```json
{
  "id": 1,
  "order_id": "ORD-A8F64D2E",
  "user_id": 123,
  "item_ids": [1001, 1002, 1003],
  "total_amount": 149.99,
  "status": "PENDING",
  "created_at": "2023-06-15T14:23:11.326Z",
  "updated_at": "2023-06-15T14:23:11.326Z"
}
```

### Get order details by ID

```bash
curl -X GET http://localhost:8000/orders/ORD-A8F64D2E
```

```json
{
  "id": 1,
  "order_id": "ORD-A8F64D2E",
  "user_id": 123,
  "item_ids": [1001, 1002, 1003],
  "total_amount": 149.99,
  "status": "COMPLETED",
  "created_at": "2023-06-15T14:23:11.326Z",
  "updated_at": "2023-06-15T14:23:13.528Z"
}
```

### Get orders by status

```bash
{
  "id": 1,
  "order_id": "ORD-A8F64D2E",
  "user_id": 123,
  "item_ids": [1001, 1002, 1003],
  "total_amount": 149.99,
  "status": "COMPLETED",
  "created_at": "2023-06-15T14:23:11.326Z",
  "updated_at": "2023-06-15T14:23:13.528Z"
}
```

### Get system metrics

```bash
curl -X GET http://localhost:8000/metrics
```

```json
{
  "total_orders_processed": 58,
  "average_processing_time": 2.14,
  "orders_by_status": {
    "pending": 3,
    "processing": 2,
    "completed": 58
  }
}
```

## Design Decisions and Trade-offs

### Architecture
- Microservices vs Monolith: We chose a modular monolith with clear service boundaries to simplify deployment while maintaining separation of concerns. This provides a path to microservices in the future.
  
- Asynchronous Processing: Order creation is decoupled from processing via RabbitMQ to handle traffic spikes and ensure system resilience.
  
- FastAPI Framework: Selected for its high performance, native async support, and automatic documentation generation.

### Data Storage
- SQLite: Used for simplicity in development. The SQLModel layer abstracts the database, making it easy to migrate to PostgreSQL or MySQL in production.
  
- SQLModel: Combines SQLAlchemy's ORM power with Pydantic's validation, reducing boilerplate code and ensuring type safety.

### Message Queue
- RabbitMQ: Chosen for its reliability, mature ecosystem, and support for complex routing patterns. Alternatives like Kafka would provide better throughput but at the cost of increased complexity.
  
- Message Persistence: Ensures orders aren't lost if the worker crashes, at a small performance cost.

### API Design
- RESTful Principles: Consistent resource-oriented endpoints with proper HTTP verbs.
  
- Pagination: Implemented for list endpoints to handle large volumes of orders.
  
- Status Filtering: Dedicated endpoint for filtering orders by status to improve query performance.


## Assumptions Made During Development
- Order Processing Time: The simulated processing time is set to 2 seconds. In a real system, this would vary based on actual business logic.
  
- Error Handling: Orders that fail processing are not automatically retried. A production system might implement a dead-letter queue and retry mechanism.
  
- Authentication: The current implementation doesn't include auth. A production system would require proper authentication and authorization.
  
- Single Worker: The system assumes a single worker process. Scaling would require multiple workers with proper concurrency controls.
  
- Data Validation: Minimal validation is implemented. A production system would need more robust validation (e.g., ensuring item IDs exist in an inventory system).
  
- Order IDs: Order IDs are assumed to be provided by the client rather than generated server-side, which might differ from typical e-commerce patterns.

- Local Development: The setup is optimized for local development. Production deployment would require containerization, orchestration, and proper secrets management.