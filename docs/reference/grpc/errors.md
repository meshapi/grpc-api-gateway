# Handling Errors

Handling errors is always crucial, and this document provides an in-depth look into error handling:


## Gateway Errors

There are several types of errors that can originate from the gateway itself. These include scenarios where a route is not found, marshaling or unmarshaling operations fail, a path parameter is missing or of an unexpected type, or an invalid enum value is received.

All of these errors are distinct instances prefixed with `Err` within the `gateway` package. Each error instance encapsulates specific information that provides additional context and details.

All of these errors can be mapped to gRPC error statuses with the appropriate error codes. Additionally, there are useful methods available to convert these gRPC codes to their corresponding HTTP status codes.

The gateway comes with default error handlers that process all errors uniformly. If you wish to implement custom error handling logic, you can inspect the error to determine whether it originated from the gateway or from a service.


!!! example
    ```go
  	errorHandler := func(
  		ctx context.Context,
  		sm *gateway.ServeMux,
  		m gateway.Marshaler,
  		w http.ResponseWriter,
  		r *http.Request,
  		err error) {

  		pathParamError := gateway.ErrPathParameterInvalidEnum{}
  		if errors.As(err, &pathParamError) {
  			// handle it differently.
  			log.Printf("invalid enum value for %s", pathParamError.Name)
  			w.WriteHeader(http.StatusNotFound)
  		}

  		// use the default handler.
  		gateway.DefaultHTTPErrorHandler(ctx, sm, m, w, r, err)
  	}
  	httpGateway := gateway.NewServeMux(gateway.WithErrorHandler(errorHandler))
    ```


| Function | Description |
| --- | --- |
| `WithErrorHandler` | Manages all non-streaming HTTP errors, providing a centralized mechanism for error handling. |
| `WithRoutingErrorHandler` | Manages errors related to routes not being found, applicable to both streaming and unary modes. |
| `WithStreamErrorHandler` | Manages errors occurring during Chunked-Transfer encoding. |
| `WithWebsocketErrorHandler` | Manages errors specific to Websocket connections. |
| `WithSSEErrorHandler` | Manages errors related to Server-Sent Events (SSE). |
