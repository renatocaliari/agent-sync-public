# NATS vs JetStream: When to Use Each

## NATS Core (Pub/Sub)

**Use when:**
- Fire-and-forget messaging (no persistence needed)
- Simple broadcasts (one publisher → multiple subscribers)
- Low latency is critical
- No need for message history
- Lost messages are acceptable if no subscriber

```go
// Publish
nc.Publish("events.users", data)

// Subscribe
sub, _ := nc.Subscribe("events.users", func(msg *nats.Msg) {
    // process message
})
```

## JetStream (Persistence)

**Use when:**
- Need history (consumers can read old messages)
- Delivery guarantee (at-least-once or exactly-once)
- Work queues (multiple consumers process messages)
- Persistent streams (event log, audit trail)
- Message replay
- Horizontal scaling of consumers

```go
// Create stream
js, _ := nc.JetStream()
js.AddStream(&nats.StreamConfig{
    Name:     "EVENTS",
    Subjects: []string{"events.>"},
    Storage:  nats.FileStorage, // or nats.MemoryStorage
})

// Publish with persistence
js.Publish("events.users", data)

// Subscribe (consumer)
sub, _ := js.SubscribeSync("events.users")
msg, _ := sub.NextMsg(time.Hour)

// Or Consumer (more robust for multiple processes)
consumer, _ := js.OrderedConsumer("events.users", "my-processor")
```

## Decision Tree

```
Need messages after restart?
├── NO → NATS Core
└── YES → JetStream

Do messages need to be processed in order?
├── NO → JetStream Consumer (parallel)
└── YES → JetStream Ordered Consumer

Need exactly-once?
├── NO → JetStream (at-least-once)
└── YES → JetStream with deduplication

High volume (> 1M msgs/sec)?
├── NO → either works
└── YES → JetStream with FileStorage

Need Key-Value store?
└── JetStream Key-Value (wrapper over streams)
```

## Pattern: NATS KV (Simple Key-Value)

For shared configuration and state, use `nats-kv`:

```go
// Create KV bucket
kv, _ := js.CreateKeyValue(&nats.KeyValueConfig{
    Bucket: "config",
})

// Set value
kv.Put("settings.theme", []byte("dark"))

// Get value
entry, _ := kv.Get("settings.theme")
value := string(entry.Value())

// Watch (reactive updates via NATS)
watcher, _ := kv.Watch("settings.*")
go func() {
    for entry := range watcher.Channel() {
        fmt.Printf("Key: %s, Value: %s\n", entry.Key, string(entry.Value()))
    }
}()
```

## Example: State Synchronization Between Instances

```go
// Instance A writes
js.Publish("state.sync", json.Marshal(State{Key: "counter", Value: 42}))

// Instance A and B receive
sub, _ := js.SubscribeSync("state.sync")
go func() {
    for {
        msg, err := sub.NextMsg(time.Second)
        if err == nil {
            var state State
            json.Unmarshal(msg.Data, &state)
            // update local state
        }
    }
}()
```

## Quick Summary

| Feature | NATS Core | JetStream |
|---------|-----------|-----------|
| Persistence | No | Yes |
| History | No | Yes |
| Replay | No | Yes |
| Throughput | Higher | High |
| Latency | Lower | Low |
| Complexity | Lower | Higher |
| Memory cost | Lower | Higher |
