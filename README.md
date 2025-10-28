# Enclave Backend

Enclave is a secure chat backend implemented in Go using the Echo web framework, Ent for data access, SQLite for persistence, and a GraphQL API layer.

## Getting started

```bash
# Install dependencies and generate Ent code (already generated in repo)
go mod tidy

# Run the development server
go run ./cmd/enclave
```

The service listens on `:8080` by default. Configure the listen address and SQLite database path using environment variables:

- `ENCLAVE_ADDR` – HTTP listen address (default `:8080`).
- `ENCLAVE_DATABASE` – Path to the SQLite database file (default `enclave.db`).

GraphQL requests are served at `http://localhost:8080/graphql`. Include an `X-User-ID` header to authorize requests for authenticated operations.

A simple health check endpoint is available at `/healthz`.

### Notifications and subscriptions

Notifications are persisted using the Ent `Notification` model. Each notification stores an encrypted payload (`cipherText`) and metadata about the recipient, originating room, and related message. Notifications remain opaque to the server—the payload should be encrypted client-side using the same scheme as chat messages.

Real-time delivery is exposed via a GraphQL subscription published at `ws://localhost:8080/graphql/ws` using the [graphql-ws protocol](https://github.com/apollographql/subscriptions-transport-ws/blob/master/PROTOCOL.md). Establish a WebSocket connection and send a `connection_init` payload containing the authenticated user ID:

```json
{ "type": "connection_init", "payload": { "authToken": "123" } }
```

The server pushes new notifications to active subscribers under the `notifications` field:

```graphql
subscription {
  notifications {
    id
    kind
    cipherText
    read
    createdAt
  }
}
```

Notifications can be created, updated (including toggling the `read` flag), and deleted through the standard GraphQL mutations. Authorization ensures only the intended recipient—or room admins when targeting room members—can manage individual notifications.
