# example.go API Documentation

## Table of Contents

- [Types](#types)
- [Enums](#enums)
- [Unions](#unions)
- [Services](#services)

## Types

### Product

Product represents an item in the inventory

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | No | Unique product identifier |
| `name` | `string` | No | Product name |
| `description` | `string` | No | Product description |
| `price` | `int64` | No | Price in cents |
| `inStock` | `bool` | No | Whether the product is in stock |
| `tags` | `[]string` | No | Product categories |
| `attributes` | `map<string, string>` | No | Product metadata |
| `discount` | `float32` | No | Discount percentage (0-100) |


### Order

Order represents a customer purchase

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | No | Unique order identifier |
| `customerId` | `string` | No | Customer ID who placed the order |
| `productIds` | `[]string` | No | List of products in the order |
| `status` | `OrderStatus` | No | Current order status |
| `totalAmount` | `int64` | No | Total order amount in cents |
| `createdAt` | `timestamp` | No | Order creation timestamp |
| `shippingAddress` | `string` | No | Delivery address |


### CreditCard

Credit card payment details

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `cardNumber` | `string` | No |  |
| `expiryDate` | `string` | No |  |
| `cvv` | `string` | No |  |


### PayPal

PayPal payment details

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `email` | `string` | No |  |


### GetProductRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | No |  |


### GetProductResponse

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `product` | `Product` | No |  |


### ListProductsRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `limit` | `int32` | No |  |
| `offset` | `int32` | No |  |


### ListProductsResponse

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `products` | `[]Product` | No |  |
| `total` | `int32` | No |  |


### DeleteProductResponse

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `success` | `bool` | No |  |


### GetOrderRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | No |  |


### GetOrderResponse

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `order` | `Order` | No |  |


## Enums

### OrderStatus

Order status enumeration

| Value | Number | Description |
|-------|--------|-------------|
| `PENDING` | 1 | Order is pending processing |
| `PROCESSING` | 2 | Order is being processed |
| `SHIPPED` | 3 | Order has been shipped |
| `DELIVERED` | 4 | Order has been delivered |
| `CANCELLED` | 5 | Order was cancelled |


## Unions

### PaymentMethod

Payment method union type

**Possible types:**

- `CreditCard`
- `PayPal`


## Services

### ProductService

Product service for managing inventory

#### Methods

##### GetProduct

Get a product by ID

**Request:** `GetProductRequest`

**Response:** `GetProductResponse`

##### ListProducts

List all products with pagination

**Request:** `ListProductsRequest`

**Response:** `ListProductsResponse`

##### CreateProduct

Create a new product

**Request:** `Product`

**Response:** `Product`

##### UpdateProduct

Update an existing product

**Request:** `Product`

**Response:** `Product`

##### DeleteProduct

Delete a product

**Request:** `GetProductRequest`

**Response:** `DeleteProductResponse`


### OrderService

Order service for managing customer orders

#### Methods

##### CreateOrder

Create a new order

**Request:** `Order`

**Response:** `Order`

##### GetOrder

Get an order by ID

**Request:** `GetOrderRequest`

**Response:** `GetOrderResponse`


