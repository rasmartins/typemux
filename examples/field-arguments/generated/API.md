# api API Documentation

## Table of Contents

- [Types](#types)

## Types

### User

Field Arguments Example
This example demonstrates the new field-level parameterized query feature
similar to GraphQL field arguments
User entity

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | Yes |  |
| `name` | `string` | Yes |  |
| `email` | `string` | Yes |  |
| `username` | `string` | Yes |  |
| `age` | `int32` | No |  |
| `isActive` | `bool` | No |  |


### Post

Post entity

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | Yes |  |
| `title` | `string` | Yes |  |
| `content` | `string` | Yes |  |
| `authorId` | `string` | Yes |  |
| `published` | `bool` | Yes |  |
| `createdAt` | `timestamp` | Yes |  |


### Comment

Comment entity

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | Yes |  |
| `postId` | `string` | Yes |  |
| `authorId` | `string` | Yes |  |
| `content` | `string` | Yes |  |
| `createdAt` | `timestamp` | Yes |  |


### SearchMetadata

Search result metadata

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `totalResults` | `int32` | Yes |  |
| `page` | `int32` | Yes |  |
| `pageSize` | `int32` | Yes |  |
| `hasNextPage` | `bool` | Yes |  |


### PostSearchResults

Search results for posts

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `posts` | `[]Post` | Yes |  |
| `metadata` | `SearchMetadata` | Yes |  |


### PostFilter

Filter options for posts

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `published` | `bool?` | No |  |
| `authorId` | `string?` | No |  |
| `minDate` | `timestamp?` | No |  |
| `maxDate` | `timestamp?` | No |  |


### Query

Query type demonstrating various field argument patterns

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `user` | `User` | No | Get a single user by ID Simple required argument |
| `users` | `[]User` | No | Get multiple users with optional pagination Multiple arguments with defaults |
| `searchUsers` | `[]User` | No | Search users with validation Argument with validation constraints |
| `findUser` | `User?` | No | Get user by username or email Optional arguments - at least one should be provided |
| `posts` | `[]Post` | No | Get posts with complex filtering Mix of required, optional, and filter objects |
| `searchPosts` | `PostSearchResults` | No | Advanced search with complex filter |
| `post` | `Post?` | No | Get a specific post |
| `comments` | `[]Comment` | No | Get comments for a post with pagination |
| `allPosts` | `[]Post` | No | Field without arguments (traditional style) |
| `featuredPosts` | `[]Post` | No | Get featured posts (no arguments) |


### Mutation

Mutation type demonstrating field arguments for mutations

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `createUser` | `User` | No | Create a new user |
| `updateUser` | `User` | No | Update user information |
| `deleteUser` | `bool` | No | Delete a user |
| `createPost` | `Post` | No | Create a new post |
| `publishPost` | `Post` | No | Publish a post |
| `addComment` | `Comment` | No | Add a comment to a post |


### AdminQuery

Example with format-specific annotations on arguments

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `userById` | `User` | No | Get user with GraphQL-specific annotations on arguments |
| `advancedSearch` | `[]Post` | No | Search with multiple format-specific customizations |


### UserProfile

Nested type to show field arguments work at any level

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `user` | `User` | Yes |  |
| `posts` | `[]Post` | No | Posts authored by this user with arguments |
| `recentComments` | `[]Comment` | No | Recent comments with pagination |
| `followerCount` | `int32` | No | Follower count (no arguments) |


### Dashboard

Type showing mix of fields with and without arguments

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `currentUser` | `User` | Yes | Current user (no arguments) |
| `notifications` | `[]string` | No | Notifications with pagination |
| `activityFeed` | `[]string` | No | Recent activity feed |
| `stats` | `map<string, int32>` | No | Summary stats (no arguments) |


