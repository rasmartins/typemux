@typemux("1.0.0")
namespace chat

// Message represents a chat message
type Message {
  id: string = 1
  content: string = 2
  sender: string = 3
  timestamp: timestamp = 4
}

// MessageRequest for sending messages
type MessageRequest {
  content: string = 1
  sender: string = 2
}

// MessageQuery for querying specific messages
type MessageQuery {
  messageId: string = 1
}

// Empty type for requests with no parameters
type Empty {}

/// Chat service with queries, mutations, and subscriptions
service ChatService {
  // Query: Get a specific message by ID
  rpc GetMessage(MessageQuery) returns (Message)

  // Query: List all messages
  rpc ListMessages(Empty) returns (Message)

  // Mutation: Send a new message
  rpc SendMessage(MessageRequest) returns (Message)

  // Mutation: Delete a message
  rpc DeleteMessage(MessageQuery) returns (Empty)

  // Subscription: Watch for new messages (inferred from stream)
  rpc WatchMessages(Empty) returns (stream Message)

  // Subscription: Watch messages from a specific sender (inferred from stream)
  rpc WatchMessagesBySender(MessageQuery) returns (stream Message)
}
