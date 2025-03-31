import asyncio
import time
from collections import defaultdict
from datetime import datetime
from typing import Dict, List, Set, AsyncGenerator, Optional, Tuple
import uuid

from app.models import Product, InventoryUpdate, BulkUpdateRequest, OperationType
from app.exceptions import ProductNotFound, RateLimitExceeded, InventoryError, CircuitBreakerOpen

class InventoryManager:
    def __init__(self):
        self._inventory: Dict[str, Product] = {}
        self._locks: Dict[str, asyncio.Lock] = {}
        self._rate_limits: Dict[str, List[float]] = {}
        self._cache: Dict[str, Tuple[Product, float]] = {}
        self._subscribers: Dict[str, Set[asyncio.Queue]] = defaultdict(set)
        self._circuit_breakers: Dict[str, dict] = {}
        self.CACHE_TTL = 30  # seconds
        self.RATE_LIMIT_WINDOW = 60  # seconds
        self.RATE_LIMIT_MAX_REQUESTS = 100  # max requests per window
        self.CIRCUIT_BREAKER_THRESHOLD = 5  # failures before opening
        self.CIRCUIT_BREAKER_TIMEOUT = 30  # seconds to stay open

    def _get_lock(self, product_id: str) -> asyncio.Lock:
        """Get or create a lock for a product"""
        if product_id not in self._locks:
            self._locks[product_id] = asyncio.Lock()
        return self._locks[product_id]

    async def _manage_cache(self, product_id: str) -> None:
        """Update or invalidate cache for a product"""
        if product_id in self._inventory:
            self._cache[product_id] = (self._inventory[product_id], time.time())
        elif product_id in self._cache:
            del self._cache[product_id]

    def check_rate_limit(self, client_id: str) -> bool:
        """Check if client has exceeded rate limit using sliding window"""
        now = time.time()
        
        # Initialize if client not seen before
        if client_id not in self._rate_limits:
            self._rate_limits[client_id] = []
        
        # Remove timestamps older than the window
        self._rate_limits[client_id] = [
            ts for ts in self._rate_limits[client_id] 
            if now - ts <= self.RATE_LIMIT_WINDOW
        ]
        
        # Check if under limit
        if len(self._rate_limits[client_id]) >= self.RATE_LIMIT_MAX_REQUESTS:
            return False
        
        # Add current timestamp
        self._rate_limits[client_id].append(now)
        return True

    async def check_circuit_breaker(self, operation: str) -> bool:
        """Check if circuit breaker allows operation"""
        if operation not in self._circuit_breakers:
            self._circuit_breakers[operation] = {
                "failures": 0,
                "last_failure": 0,
                "status": "closed"
            }
        
        breaker = self._circuit_breakers[operation]
        now = time.time()
        
        # Check if open circuit should be reset (timeout expired)
        if (breaker["status"] == "open" and 
            now - breaker["last_failure"] > self.CIRCUIT_BREAKER_TIMEOUT):
            breaker["status"] = "half-open"
            breaker["failures"] = 0
        
        # Return False if circuit is open
        if breaker["status"] == "open":
            return False
            
        return True

    def record_success(self, operation: str) -> None:
        """Record successful operation for circuit breaker"""
        if operation in self._circuit_breakers:
            self._circuit_breakers[operation]["status"] = "closed"
            self._circuit_breakers[operation]["failures"] = 0

    def record_failure(self, operation: str) -> None:
        """Record failed operation for circuit breaker"""
        if operation not in self._circuit_breakers:
            self._circuit_breakers[operation] = {
                "failures": 0,
                "last_failure": 0,
                "status": "closed"
            }
            
        breaker = self._circuit_breakers[operation]
        breaker["failures"] += 1
        breaker["last_failure"] = time.time()
        
        if breaker["failures"] >= self.CIRCUIT_BREAKER_THRESHOLD:
            breaker["status"] = "open"

    async def get_product(self, product_id: str) -> Product:
        """Get product by ID, using cache if available"""
        # Check cache
        if product_id in self._cache:
            product, timestamp = self._cache[product_id]
            if time.time() - timestamp <= self.CACHE_TTL:
                return product
        
        # Cache miss or expired
        if product_id not in self._inventory:
            raise ProductNotFound(product_id)
            
        # Update cache and return
        await self._manage_cache(product_id)
        return self._inventory[product_id]

    async def update_with_version(self, product_id: str, update: InventoryUpdate, version: int) -> Product:
        """Update product with optimistic locking using version number"""
        async with self._get_lock(product_id):
            if product_id not in self._inventory:
                raise ProductNotFound(product_id)
                
            product = self._inventory[product_id]
            
            # Check version for optimistic locking
            if product.version != version:
                raise InventoryError(f"Version mismatch: expected {product.version}, got {version}")
            
            # Apply update based on operation type
            new_product = await self._apply_update(product, update)
            
            # Increment version and update timestamp
            new_product.version += 1
            new_product.last_updated = datetime.now()
            
            # Save and notify subscribers
            self._inventory[product_id] = new_product
            await self._manage_cache(product_id)
            await self._notify_subscribers(product_id, new_product)
            
            return new_product

    async def _apply_update(self, product: Product, update: InventoryUpdate) -> Product:
        """Apply update operation to product"""
        # Create a copy to avoid mutating the original
        updated = product.model_copy(deep=True)
        
        if update.operation == OperationType.ADD:
            updated.quantity += update.quantity
            
        elif update.operation == OperationType.SUBTRACT:
            if product.quantity < update.quantity:
                raise InventoryError(f"Not enough inventory: have {product.quantity}, requested {update.quantity}")
            updated.quantity -= update.quantity
            
        elif update.operation == OperationType.SET:
            updated.quantity = update.quantity
            
        elif update.operation == OperationType.RESERVE:
            if product.quantity - product.reserved < update.quantity:
                raise InventoryError(f"Cannot reserve {update.quantity} units, only {product.quantity - product.reserved} available")
            updated.reserved += update.quantity
            
        elif update.operation == OperationType.RELEASE:
            if product.reserved < update.quantity:
                raise InventoryError(f"Cannot release {update.quantity} units, only {product.reserved} reserved")
            updated.reserved -= update.quantity
        
        # Check if inventory is below minimum threshold after update
        if updated.quantity < updated.min_quantity:
            # This would be a good place to trigger a restock alert
            pass
            
        # Check if inventory exceeds maximum threshold
        if updated.max_quantity is not None and updated.quantity > updated.max_quantity:
            raise InventoryError(f"Quantity {updated.quantity} exceeds maximum allowed {updated.max_quantity}")
            
        return updated

    async def bulk_update(self, updates: BulkUpdateRequest) -> Dict[str, Product]:
        """Process multiple updates in bulk"""
        results = {}
        errors = {}
        
        # Process each update
        for item in updates.updates:
            try:
                # Apply update with optimistic locking
                product = await self.update_with_version(
                    item.product_id, 
                    item.update, 
                    item.version
                )
                results[item.product_id] = product
            except Exception as e:
                # Record failures for rollback if needed
                errors[item.product_id] = str(e)
        
        # If any errors, you might want to implement rollback logic here
        # For simplicity, we'll just report the errors
        
        return results, errors

    async def _notify_subscribers(self, product_id: str, product: Product) -> None:
        """Send updates to all subscribers for a product"""
        if product_id in self._subscribers:
            dead_queues = set()
            
            for queue in self._subscribers[product_id]:
                try:
                    await queue.put(product)
                except Exception:
                    # Mark for removal if queue is closed or full
                    dead_queues.add(queue)
            
            # Clean up dead queues
            self._subscribers[product_id] -= dead_queues

    async def subscribe_to_updates(self, product_id: str) -> AsyncGenerator[Product, None]:
        """Subscribe to real-time updates for a product"""
        if product_id not in self._inventory:
            raise ProductNotFound(product_id)
            
        # Create queue for this subscriber
        queue = asyncio.Queue(maxsize=100)
        self._subscribers[product_id].add(queue)
        
        try:
            # Send initial state
            yield self._inventory[product_id]
            
            # Stream updates
            while True:
                try:
                    product = await queue.get()
                    yield product
                except asyncio.CancelledError:
                    break
        finally:
            # Clean up subscription when client disconnects
            if product_id in self._subscribers:
                self._subscribers[product_id].discard(queue)

    async def schedule_cleanup(self) -> None:
        """Background task to clean up expired cache entries"""
        while True:
            now = time.time()
            
            # Clean up expired cache entries
            expired_keys = [
                k for k, (_, timestamp) in self._cache.items() 
                if now - timestamp > self.CACHE_TTL
            ]
            for key in expired_keys:
                del self._cache[key]
            
            # Clean up empty subscriber lists
            empty_subs = [
                k for k, v in self._subscribers.items() if not v
            ]
            for key in empty_subs:
                del self._subscribers[key]
            
            # Wait before next cleanup
            await asyncio.sleep(60)  # Run every minute

    async def trace_request(self, request_id: str, operation: str) -> None:
        """Log request for tracing purposes"""
        # In a real implementation, you'd likely use a distributed tracing system
        # For this example, we'll just print to console
        print(f"TRACE [{datetime.now().isoformat()}] {request_id}: {operation}")

    def add_product(self, product: Product) -> None:
        """Add a new product to inventory"""
        self._inventory[product.id] = product