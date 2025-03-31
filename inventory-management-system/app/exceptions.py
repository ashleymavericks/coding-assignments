from fastapi import HTTPException, status

class ProductNotFound(HTTPException):
    def __init__(self, product_id: str):
        super().__init__(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Product with ID {product_id} not found"
        )

class RateLimitExceeded(HTTPException):
    def __init__(self, client_id: str):
        super().__init__(
            status_code=status.HTTP_429_TOO_MANY_REQUESTS,
            detail=f"Rate limit exceeded for client {client_id}"
        )

class InventoryError(HTTPException):
    def __init__(self, message: str):
        super().__init__(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=message
        )

class BulkOperationError(HTTPException):
    def __init__(self, errors: dict):
        super().__init__(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail={"message": "Errors in bulk operation", "errors": errors}
        )

class CircuitBreakerOpen(HTTPException):
    def __init__(self, operation: str):
        super().__init__(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail=f"Circuit breaker open for operation: {operation}"
        ) 