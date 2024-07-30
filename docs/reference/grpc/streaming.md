# Streaming

One of the main motivations for this project was to support streaming aspects of gRPC in the HTTP Gateway as well.

!!! warning
    This feature and API is fairly new and while it functions, it has not been put in a high latency production
    environment and is not battle-tested. Conduct your own tests and experiments to ensure it is a good fit for your
    project and needs.

## Streaming Modes

There are three different supported modes of streaming. It is recommended to read the documentation for the streaming
modes that you plan on using to ensure you are aware of all features and behaviors.

| Streaming Mode      | Description                               | HTTP Method |
| --- | --- | --- |
| Server-Sent Events (SSE) | Using SSE, client subscribes to an event stream and server sends events (messages) to the client. | GET |
| WebSocket | Using WebSocket both client and server can send messages. | GET |
| ChunkedTransfer | It is a way to stream a message in multiple chunks but is subject to short time out. | * |

### 1. Server-sent events (SSE)

[Server-sent events (SSE)](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events) is a standard for enabling servers to push real-time updates to clients over a single, long-lived HTTP connection. Unlike WebSockets, which allow for bidirectional communication, SSE provides a simpler, one-way communication channel from the server to the client.

With SSE, the server can continuously send updates to the client as new data becomes available, making it ideal for applications that require real-time notifications, such as live feeds, news updates, or stock price tickers. Clients subscribe to the stream by opening a persistent HTTP connection and can receive automatic updates without the need for repeated polling.

Key features of SSE include:

* Simple Implementation: SSE uses standard HTTP protocols, making it easy to implement and compatible with existing web infrastructure.
* Automatic Reconnection: Built-in support for automatic reconnection in case the connection drops.
* Event Identification: Supports event IDs, allowing clients to track the last received event and resume from where they left off.

To use SSE, the client simply opens a connection to an endpoint that delivers events, and the server streams text-based event data, typically in the form of plain text or JSON.

You can use SSE for any gRPC method as long as:

1. Server returns a stream which includes server-streaming and bidirectional modes.
2. HTTP method is _GET_

!!! note
    Even though bidirectional gRPC methods are accepted and can be used, it's important to note that while requests _can_
    be streamed to the server using Chunked-Transfer, SSE is server-streaming only.
    As a result, the client cannot arbitrarily send messages. It can stream its request and then wait for server to
    push messages.

#### Note on closing the stream

When using SSEs in the browser, if the connection gets closed for virtually any reason other than the client closing
the stream, browser tries to establish the connection again.

For this reason, __IF__ the gRPC server implementation closes the stream, an _End-of-stream (EOS)_ message gets pushed
by the server to indicate that the server has reached the end of the stream and that there are no more events. To
handle this correctly, you will need to close the stream upon receiving this message. This message can be customized if
you prefer to send a different content instead.

The default EOS has _ID_ `EOS` with _Event_ `EOS` and no data.

```text
id: EOS
event: EOS
data:
```

!!! example
    ```javascript
    const notificationsSource = new EventSource("/notifications");

    notificationsSource.onmessage = (event) => {
      console.log(`New Event: ${event.data}`)
    };

    notificationsSource.addEventListener('EOS', (event) => {
      console.log(`End of stream received, server does not want to send more events. Closing the event source.`);
      notificationsSource.close();
    });
    ```
    
!!! info
    If your gRPC method implementation is designed to end the stream, you might benefit from re-evaluating whether or
    not SSE is the right choice for your specific need.

#### Custom Events

One of the features of the SSE is the ability to push messages with specific _event_ names.
At the present moment, all SSE messages (save for the EOS message) get no `event` or `id` value.
In future, you will be able to use gRPC metadata to specify these properties but this is not possible presently.

Because of this all messages will be received using the `onmessage` handler if you are using JavaScript in the browser:

```javascript
const eventSource = new EventSource("/path/to/endpoint");

eventSource.onmessage = (event) => {
  // event.message
};
```

#### Important Note for HTTP 1.1

!!! warning
    When not used over HTTP/2, SSE suffers from a limitation to the maximum number of open connections, which can be especially painful when opening multiple tabs, as the limit is per browser and is set to a very low number (6).

    The issue has been marked as "Won't fix" in Chrome and Firefox. This limit is per browser + domain, which means that you can open 6 SSE connections across all of the tabs to www.example1.com and another 6 SSE connections to www.example2.com (per Stackoverflow).

    When using HTTP/2, the maximum number of simultaneous HTTP streams is negotiated between the server and the client (defaults to 100).

### 2. WebSockets

TBD

### 3. Chunked Transfer

TBD
