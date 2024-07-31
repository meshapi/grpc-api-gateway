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
| [Server-Sent Events (SSE)](#1-server-sent-events-sse) | Using SSE, client subscribes to an event stream and server sends events (messages) to the client. | GET |
| [WebSocket](#2-websockets) | Using WebSocket both client and server can send messages. | GET |
| [Chunked-Transfer](#3-chunked-transfer) | It is a way to stream a message in multiple chunks but is subject to short time out. | * |

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

To use a custom EOS message:

```go
gateway.NewServeMux(gateway.WithSSEConfig(gateway.SSEConfig{EndOfStreamMessage: nil}))
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


#### Error Handling

If an error is returned by the server, that error gets handled using the `SSEErrorHandlerFunc` and
the connection will be closed. If you would like to communicate errors without closing the connection,
include that error structure into your proto response messages.

By default, the error received from the server gets marshaled and sent to the client.
To customize this behavior and use your own error handler for SSE connections, use `WithSSEErrorHandler` option:

```go
gateway.NewServeMux(gateway.WithSSEErrorHandler(myCustomErrorHandler))
```

#### Using outside of browsers

Server-sent events can be used outside of browsers as well. Reading specification of SSE is the best way to learn how
to correctly use this feature. However, to use the SSE streaming mode, send `Accept: text/event-stream` header in your
HTTP request.

### 2. WebSockets

WebSockets is a communication protocol providing full-duplex communication channels over a single TCP connection.
It allows real-time data transfer between clients and servers with low latency, enabling interactive web applications
such as live chat, streaming, and multiplayer games.
WebSockets are initiated through an HTTP handshake and then upgraded to a persistent connection, facilitating efficient data exchange.

While WebSocket sounds like an ideal choice for streaming, there are some common limitations and challenges of using WebSockets:

* __Scalability__: Managing a large number of concurrent WebSocket connections can be resource-intensive and complex, requiring robust server infrastructure and load balancing strategies.

* __Network Issues__: WebSocket connections can be susceptible to interruptions due to network instability, requiring efficient reconnection strategies to maintain a seamless user experience.

* __Security__: Ensuring secure WebSocket communication involves implementing measures such as encryption (TLS/SSL), authentication, and protection against common attacks like cross-site WebSocket hijacking.

* __Browser Compatibility__: While most modern browsers support WebSockets, some older versions may not, necessitating fallback mechanisms like long polling or Server-Sent Events (SSE).

* __Firewall and Proxy Restrictions__: Some firewalls and proxies may block or interfere with WebSocket traffic, requiring additional configurations to allow WebSocket connections.

* __Error Handling__: Properly managing and handling errors and edge cases in WebSocket communication is crucial for maintaining application stability and providing a good user experience.

* __State Management__: Keeping track of client state across WebSocket connections can be challenging, especially in distributed systems or applications requiring high availability and fault tolerance.

#### Enabling WebSockets

To enable the gRPC API Gateway to support a WebSocket interface for your gRPC streaming API while
allowing you to choose and manage the technology used for connections, an indirection is employed.

By default, the gateway does not include a WebSocket handler. You will need to provide a custom
connection upgrader that handles message sending and receiving. This approach allows you to manage
aspects such as authorization, compression, and other WebSocket-specific concerns. The gRPC API Gateway
functions as a mapping library, not a WebSocket library. It is designed to integrate seamlessly with
various WebSocket libraries, giving you the flexibility to choose the one that best fits your needs.

!!! info
    [gorilla/websocket](https://github.com/gorilla/websocket) is a popular Go implementation of
    the WebSocket protocol. There is a rudimentary wrapper for integrating this library with the
    gRPC API Gateway. This wrapper is used in all the examples in this section.

    To use this wrapper:

    ```sh
    go get github.com/meshapi/grpc-api-gateway/websocket/wrapper/gorillawrapper
    ```

To enable WebSockets in your gateway, use `WithWebSocketUpgrader` option:

```go
gateway.NewServeMux(gateway.WithWebsocketUpgrader(websocketUpgradeFunc))
```

The upgrader function has the following signature:

`func(w http.ResponseWriter, r *http.Request) (websocket.Connection, error)`

!!! example
    Below is an example using the [gorilla/websocket](https://github.com/gorilla/websocket) library:

    ```go linenums="1"
    import (
      ws "github.com/meshapi/grpc-api-gateway/websocket"
      "github.com/meshapi/grpc-api-gateway/websocket/wrapper/gorillawrapper"
	  "github.com/meshapi/grpc-api-gateway/gateway"

	  "github.com/gorilla/websocket"
    )

    // ...
    
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // NB: not ideal for production code.
		},
	}

	websocketUpgradeFunc := gateway.WebsocketUpgradeFunc(
        func(w http.ResponseWriter, r *http.Request) (ws.Connection, error) {
            connection, err := upgrader.Upgrade(w, r, nil) //(1)!
            if err != nil {
                log.Printf("ws error: %s", err)
                return nil, fmt.Errorf("failed to upgrade: %w", err)
            }

            return gorillawrapper.New(connection), nil //(2)!
        })

	grpcGateway := gateway.NewServeMux(gateway.WithWebsocketUpgrader(websocketUpgradeFunc))
    ```

    1. WebSocket connection is prepared here using the WebSocket library of choice.
    2. A thin adaptor is used to wrap the WebSocket connection only to return a `ws.Connection` type.

!!! info
    If no WebSocket upgrader is specified using `WithWebSocketUpgrader`, all requests asking for a
    WebSocket protocol upgrade receive an error indicating the streaming method is not supported.

#### Error Handling

If an error occurs while receiving or sending messages, a WebSocket-specific error handler will
be invoked with the encountered error.
After handling the error, both the WebSocket connection and the gRPC streams will be closed.
Consequently, a reconnection will be required to resume sending or receiving messages.

To communicate an error without closing or interrupting the connection,
you should include an error structure in your proto response messages.

### 3. Chunked Transfer

Chunked Transfer is a streaming method but unlike the other streaming modes, it is not long-lived.
This mode can be used to stream a large message in chunks. For instance, say a user wants to load
a large number of items. Fetching these items does not take long but sending them over the wire does.
You can benefit from Chunked-Transfer to process items as they are received by using Chunked-Transfer
encoding.

#### Error Handling

Similar to the other methods, if any error is encountered, the stream get interrupted immediately and the error handler
gets called. To use a custom error handler logic, use `WithStreamErrorHandler` option:

```go linenums="1"
gateway.NewServeMux(gateway.WithStramErrorHandler(myCustomHandler))
```

## Toggles / Disable streaming

All streaming modes are _enabled_ by default. Note that _enabled_ here does not mean they are available, just that they
are allowed so that when the conditions are met, they are _allowed_ to be used.

WebSockets can only be requested on endpoints with `GET` method and you need to have added
an _Upgrader_ to use the WebSockets streaming mode.
Server-sent events (SSE) need to be `GET` endpoints and require `Accept` header with `text/event-streaming` in the
request.

To disable a specific streaming mode for an endpoint binding, use the [Stream](/grpc-api-gateway/reference/grpc/config#stream) config.

!!! example
    Imagine an endpoint for chatting. This method is bidirectional streaming and _can_
    accept _Chunked-Transfer_ encoding or _Server-sent events_ but it does not make sense for
    it to do so because with Chunked-Transfer, we cannot have a long-lived connection and with
    SSE, client cannot send messages to the server.

    === "Configuration"
        ```yaml title="chat_gateway.yaml"
        gateway:
          endpoints:
            - get: "/chat"
              selector: "~.ChatService.StartChat"
              stream:
                disable_sse: true
                disable_chunked_transfer: true
        ```

    === "Proto Annotations"
        ```proto title="service.proto"
        service ChatService {
            rpc StartChat(ChatRequest) returns (ChatResponse) {
                option (meshapi.gateway.http) = {
                    get: "/chat",
                    stream: {
                      disable_sse: true,
                      disable_chunked_transfer: true
                    }
                };
            }
        }
        ```

    Now, these endpoint bindings do NOT accept SSE or Chunked-Transfer and 
    return streaming not supported error.
