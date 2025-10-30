# examples.maps API Documentation

## Table of Contents

- [Types](#types)
- [Services](#services)

## Types

### Product

Product information

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | Yes |  |
| `name` | `string` | Yes |  |
| `price` | `float64` | Yes |  |
| `description` | `string` | No |  |


### User

User profile

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | Yes |  |
| `username` | `string` | Yes |  |
| `email` | `string` | Yes |  |


### Settings

Configuration settings

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `theme` | `string` | No |  |
| `language` | `string` | No |  |
| `notifications` | `bool` | No |  |


### Inventory

Inventory tracking with maps of custom types

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `productsByWarehouse` | `map<string, Product>` | Yes | Map of warehouse ID to Product Proto: map<string, Product> GraphQL: [InventoryProductsEntry!]! with key/value fields OpenAPI: object with Product values |
| `quantities` | `map<string, int32>` | Yes | Map of product ID to quantity (primitive value) Proto: map<string, int32> GraphQL: JSON scalar or [QuantityEntry!]! OpenAPI: object with integer values |
| `supplierProducts` | `map<string, >` | No | Map of supplier ID to list of products Proto: map<string, ProductList> (requires wrapper) GraphQL: [SupplierProductsEntry!]! OpenAPI: object with array values |


### UserPreferences

User preferences with various map types

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `userId` | `string` | Yes |  |
| `featureSettings` | `map<string, Settings>` | No | Map of feature name to Settings |
| `friends` | `map<string, User>` | No | Map of friend ID to User profile |
| `metadata` | `map<string, string>` | No | Simple key-value pairs (string to string) |


### ShoppingCart

Shopping cart with product maps

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `cartId` | `string` | Yes |  |
| `userId` | `string` | Yes |  |
| `items` | `map<string, Product>` | Yes | Map of product ID to Product details |
| `itemQuantities` | `map<string, int32>` | Yes | Map of product ID to quantity |
| `totalPrice` | `float64` | No | Total price |


### GetInventoryRequest

Request to get inventory

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `warehouseId` | `string` | Yes |  |


### UpdateCartRequest

Request to update cart

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `cartId` | `string` | Yes |  |
| `productId` | `string` | Yes |  |
| `quantity` | `int32` | Yes |  |


### CartResponse

Response with cart details

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `cart` | `ShoppingCart` | Yes |  |


## Services

### InventoryService

Service demonstrating map operations

#### Methods

##### GetInventory

Get inventory for a warehouse

**Request:** `GetInventoryRequest`

**Response:** `Inventory`

**HTTP:** `GET /api/v1/inventory/{warehouseId}`

##### UpdateCart

Update shopping cart

**Request:** `UpdateCartRequest`

**Response:** `CartResponse`

**HTTP:** `PUT /api/v1/cart/{cartId}`


