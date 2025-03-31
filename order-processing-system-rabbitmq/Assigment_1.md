# Assignment

## Take-Home Assignment SSE-1: Backend System for Order Processing

### Objective

Evaluate the candidate's ability to design and implement a backend system that demonstrates proficiency in building modular, maintainable, and scalable systems while covering database design, queuing, distributed system fundamentals, and metrics reporting. You will have a total of 48 hours to complete this assignment.

### Problem Statement

Build a backend system to manage and process orders in an e-commerce platform. The system should:

#### Core Functionality:

- Provide a RESTful API to accept orders with fields such as:
  - user_id
  - order_id 
  - item_ids
  - total_amount

- Simulate asynchronous order processing using an in-memory queue (e.g., Python queue.Queue or equivalent)

- Provide an API to check the status of orders:
  - Pending
  - Processing 
  - Completed

- Implement an API to fetch key metrics, including:
  - Total number of orders processed
  - Average processing time for orders
  - Count of orders in each status:
    - Pending
    - Processing
    - Completed

#### Constraints:

- Database:
  - Use SQLite/PostgreSQL/MySQL for order storage

- Queue:
  - Use an in-memory queue for asynchronous processing

- Scalability:
  - Ensure the system can handle 1,000 concurrent orders (simulate load)

### Deliverables

#### Functional Backend Code:

- A fully functioning backend service with:
  - RESTful APIs for order management and metrics reporting
  - Modular components for queuing, database operations, and metrics computation

#### Database Design:

- Schema to store orders with fields: order_id, user_id, item_ids, total_amount, and status
- SQL scripts for schema creation and sample data population

#### Queue Processing:

- An asynchronous queue that processes orders and updates their status

#### Metrics API:

- Accurate reporting of metrics such as total orders processed, average processing time, and current order statuses

#### Documentation:

- A README.md file that includes:
  - Instructions for setting up and running the application
  - Example API requests and responses (e.g: using curl or Postman)
  - Explanation of design decisions and trade-offs
  - Assumptions made during development

#### Tests:

- Unit tests for key components:
  - API endpoints
  - Database operations
  - Queue processing