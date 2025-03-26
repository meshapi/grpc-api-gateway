# Streaming

One of the main motivations for this project was to support streaming aspects of gRPC in the HTTP Gateway as well.

!!! warning
    This feature and API are relatively new. While it is functional, it has not been extensively tested in high-latency production environments. Please conduct thorough testing and experimentation to ensure it meets the requirements of your project.

## Streaming Modes

This project supports three distinct streaming modes. It is highly recommended to review the documentation for each streaming mode you intend to use to fully understand their features and behaviors.

| Streaming Mode      | Description                               | HTTP Method |
| --- | --- | --- |
| [Server-Sent Events (SSE)](#1-server-sent-events-sse) | SSE allows the client to subscribe to a stream of events sent by the server over a single HTTP connection. | GET |
| [WebSocket](#2-websockets) | WebSocket enables bidirectional communication between the client and server over a single connection. | GET |
| [Chunked-Transfer](#3-chunked-transfer) | This method streams a message in multiple chunks, making it suitable for transferring large payloads efficiently. However, it is subject to short timeouts. | * |

### 1. Server-sent events (SSE)

[Server-sent events (SSE)](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events) is a technology that allows servers to push real-time updates to clients over a single, long-lived HTTP connection. Unlike WebSockets, which support bidirectional communication, SSE offers a simpler, one-way communication channel from the server to the client.

SSE enables the server to continuously send updates to the client as new data becomes available. This makes it ideal for applications that require real-time notifications, such as live feeds, news updates, or stock price tickers. Clients subscribe to the stream by opening a persistent HTTP connection, allowing them to receive automatic updates without the need for repeated polling.

Key features of SSE include:

* **Ease of Implementation**: SSE leverages standard HTTP protocols, ensuring straightforward integration and compatibility with existing web infrastructure.
* **Automatic Reconnection**: Provides built-in support for automatic reconnection if the connection is lost, enhancing reliability.
* **Event Identification**: Supports event IDs, enabling clients to track the last received event and resume from where they left off seamlessly.

To utilize SSE, the client opens a connection to an endpoint that delivers events, and the server streams text-based event data, typically in the form of plain text or JSON.

You can use SSE for any gRPC method as long as:

1. The server returns a stream, which includes both server-streaming and bidirectional modes.
2. The HTTP method is _GET_.

!!! note
    While bidirectional gRPC methods are supported, it's crucial to understand that SSE is strictly server-streaming. Although requests can be streamed to the server using Chunked-Transfer, the client cannot send messages arbitrarily with SSE. Instead, the client streams its request and then waits for the server to push messages.

#### Note on closing the stream

When using SSEs in the browser, if the connection is closed for any reason other than the client closing the stream, the browser will attempt to re-establish the connection.

To handle this, if the gRPC server implementation closes the stream, it sends an _End-of-stream (EOS)_ message to indicate that the stream has ended and no more events will be sent. Upon receiving this message, you should close the stream. This message can be customized if you prefer to send different content.

The default EOS message has an _ID_ of `EOS`, an _Event_ of `EOS`, and no data.

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
    If your gRPC method implementation is designed to end the stream, it may be worth reconsidering whether SSE is the most suitable choice for your needs.

#### Custom Events

One of the key features of SSE is the ability to push messages with specific _event_ names. Currently, all SSE messages (except for the EOS message) do not include `event` or `id` values. In the future, it will be possible to use gRPC metadata to specify these properties, but this functionality is not available at present.

As a result, all messages will be received using the `onmessage` handler when using JavaScript in the browser:

```javascript
const eventSource = new EventSource("/path/to/endpoint");

eventSource.onmessage = (event) => {
  // event.message
};
```

#### Important Note for HTTP 1.1

!!! warning
    When not used over HTTP/2, SSE is limited by the maximum number of open connections, which can be particularly problematic when opening multiple tabs. This limit is set to a very low number (6) per browser.

    This issue has been marked as "Won't fix" in both Chrome and Firefox. The limit is per browser and domain, meaning you can open 6 SSE connections across all tabs to www.example1.com and another 6 SSE connections to www.example2.com.

    When using HTTP/2, the maximum number of simultaneous HTTP streams is negotiated between the server and the client, with a default of 100.


#### Error Handling

If an error is returned by the server, it is handled using the `SSEErrorHandlerFunc`, and the connection will be closed. To communicate errors without closing the connection, include the error structure in your proto response messages.

By default, the error received from the server is marshaled and sent to the client. To customize this behavior and use your own error handler for SSE connections, use the `WithSSEErrorHandler` option:

```go
gateway.NewServeMux(gateway.WithSSEErrorHandler(myCustomErrorHandler))
```

#### Using outside of browsers

Server-sent events can be used outside of browsers as well. Reading specification of SSE is the best way to learn how
to correctly use this feature. However, to use the SSE streaming mode, send `Accept: text/event-stream` header in your
HTTP request.

### 2. WebSockets

WebSockets is a communication protocol that provides full-duplex communication channels over a single TCP connection. It enables real-time data transfer between clients and servers with low latency, making it ideal for interactive web applications such as live chat, streaming, and multiplayer games. WebSockets are initiated through an HTTP handshake and then upgraded to a persistent connection, facilitating efficient data exchange.

While WebSockets offer significant advantages for streaming, there are several limitations and challenges to consider:

* __Scalability__: Managing a large number of concurrent WebSocket connections can be resource-intensive and complex, requiring robust server infrastructure and effective load balancing strategies.

* __Network Issues__: WebSocket connections can be prone to interruptions due to network instability, necessitating efficient reconnection strategies to maintain a seamless user experience.

* __Security__: Secure WebSocket communication requires implementing encryption (TLS/SSL), authentication, and protection against common attacks such as cross-site WebSocket hijacking.

* __Browser Compatibility__: Although most modern browsers support WebSockets, some older versions do not. This necessitates fallback mechanisms like long polling or Server-Sent Events (SSE).

* __Firewall and Proxy Restrictions__: WebSocket traffic can be blocked or interfered with by some firewalls and proxies, requiring additional configurations to ensure proper connections.

* __Error Handling__: Effective error management and handling of edge cases in WebSocket communication are essential for maintaining application stability and ensuring a good user experience.

* __State Management__: Tracking client state across WebSocket connections can be complex, especially in distributed systems or applications that require high availability and fault tolerance.

#### Enabling WebSockets

To enable the gRPC API Gateway to support a WebSocket interface for your gRPC streaming API, you need to provide a custom connection upgrader. This allows you to manage the specifics of WebSocket connections, such as authorization, compression, and other concerns.

By default, the gateway does not include a WebSocket handler. You must supply a custom connection upgrader to handle message sending and receiving. This approach gives you control over WebSocket-specific details and ensures the gRPC API Gateway can integrate seamlessly with various WebSocket libraries, allowing you to choose the one that best fits your needs.

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

If an error occurs while receiving or sending messages, a WebSocket-specific error handler will be triggered to manage the encountered error. After the error is handled, both the WebSocket connection and the gRPC streams will be terminated. As a result, a reconnection will be necessary to continue sending or receiving messages.

To report an error without closing or interrupting the connection, include an error structure in your proto response messages.

### 3. Chunked Transfer

Chunked Transfer is a streaming method that, unlike other streaming modes, is not long-lived. This mode is ideal for streaming large messages in chunks. For example, if a user needs to load a large number of items, fetching these items might be quick, but transmitting them over the network can be time-consuming. Chunked-Transfer encoding allows you to process items as they are received, making the transfer more efficient.

#### Error Handling

Similar to the other methods, if any error is encountered, the stream get interrupted immediately and the error handler
gets called. To use a custom error handler logic, use `WithStreamErrorHandler` option:

```go linenums="1"
gateway.NewServeMux(gateway.WithStramErrorHandler(myCustomHandler))
```

## Toggles / Disable streaming

All streaming modes are _enabled_ by default. However, _enabled_ does not imply they are immediately available; it means they are permitted to be used when the appropriate conditions are met.

WebSockets can only be utilized on endpoints with the `GET` method, and you must have added an _Upgrader_ to enable the WebSockets streaming mode. Server-sent events (SSE) also require `GET` endpoints and the `Accept` header with `text/event-stream` in the request.

To disable a specific streaming mode for an endpoint binding, use the [Stream](/grpc-api-gateway/reference/grpc/config#stream) configuration.

!!! example
    Imagine an endpoint for a chat application. This method supports bidirectional streaming and _can_ technically accept _Chunked-Transfer_ encoding or _Server-sent events_. However, using these modes is impractical because Chunked-Transfer does not support long-lived connections, and SSE does not allow the client to send messages to the server.

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
