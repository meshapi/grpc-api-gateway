# Query Parameter Binding

It is important to understand how query parameters get processed by the gRPC Gateway
so you are aware of the logic and can customize your API based on your needs.

### Default Behavior

In each [EndpointConfig](/grpc-api-gateway/reference/grpc/config),
any proto field from the request message that is not already bound to path parameters or HTTP body
automatically gets bound to query parameters.


#### Naming

By default, the name of these query parameters are the proto names
and the nested fields contain the parent fields name followed by a dot.

!!! example
    Consider the following proto messages where `Request` is the request message:

    ```proto
    message Options {
        bool case_sensitive = 1;
    }

    message Request {
        Options options = 1;
        string some_input = 2;
    }
    ```

    Assuming neither are captured by path parameters or the HTTP body,
    query parameters `some_input` and `options.case_sensitive` get bound to the proto message.


### Customization

There are a number of customizations available.

#### Disable Automatic Discovery

In [EndpointConfig](/grpc-api-gateway/reference/grpc/config),
you can disable the automatic discovery and binding of query parameters.
Doing so means for that specific endpoint binding, only explicitly defined query parameter bindings are considered.


=== "Configuration"
    ```yaml title="service_gateway.yaml" linenums="1" hl_lines="5"
    gateway:
      endpoints:
        - post: "/my-endpoint"
          selector: "~.MyService.MyMethod"
          disable_query_param_discovery: true
    ```

=== "Proto Annotations"
    ```proto title="service.proto" linenums="1" hl_lines="5"
    service MyService {
        rpc MyMethod(Request) returns (Response) {
            option (meshapi.gateway.http) = {
                post: "/my-endpoint",
                disable_query_param_discovery: true
            };
        }
    }
    ```

Some applications of using this setting:

* Use custom name for all query parameters.
* Only expose parts of the proto message to the HTTP request.


#### Additional Query Parameter Binding and Aliases

If you would like to use a custom name for a query parameter or allow multiple names (aliases) for the same message field, you can explicitly define query parameter to request proto message field binding.

Using `query_params` in [EndpointConfig](/grpc-api-gateway/reference/grpc/config), you can add multiple [QueryParameterBinding](/grpc-api-gateway/reference/grpc/config/#queryparameterbinding) objects.

!!! example
    Consider request proto message below:
    ```proto
    message PageOptions {
        uint32 per_page = 1;
    }

    message QueryRequest {
        string term = 1;
        string language = 2;
        PageOptions pagination = 3;
    }
    ```
    Assume we would like to:

    1. Accept either `language` or `lang` (in order of priority)
    2. Use `per_page` instead of `pagination.per_page` which is the default name.

    === "Configuration"
        ```yaml title="query_gateway.yaml" linenums="1" hl_lines="5-11"
        gateway:
          endpoints:
            - get: "/query"
              selector: "~.QueryService.Query"
              query_params:
                - selector: 'language'
                  name: 'lang'
                - selector: 'language'
                  name: 'language'
                - selector: 'pagination.per_page'
                  name: 'per_page'
        ```

    === "Proto Annotations"
        ```proto title="query.proto" linenums="1" hl_lines="5-9"
        service QueryService {
            rpc Query(QueryRequest) returns (QueryResponse) {
                option (meshapi.gateway.http) = {
                    get: "/query",
                    query_params: [
                        {selector: 'language', name: 'lang'},
                        {selector: 'language', name: 'language'},
                        {selector: 'pagination.per_page', name: 'per_page'}
                    ]
                };
            }
        }
        ```

    With the configuration above:

    * `language` and `lang` both are accepted and bound to `language` in the proto request.
    * `per_page` is accepted and is bound to `pagination.per_page`.
    * `term` does not have an explicit binding but gets automatically discovered and added.

!!! warning
    Once you add an explicit binding for a proto field, that specific field no longer receives automatic bindings.
    Only the explicitly added bindings will be applied.
    For instance, in the example above, `pagination.per_page` is not longer accepted because an
    explicit binding for `pagination.per_page` is defined.

##### Prioritization

If multiple names are defined for the same proto field, the later bindings take priority over previously defined ones.
In the example above, `language` takes priority over `lang`.

#### Ignoring Parameters

You may want to avoid binding any query parameter to some proto fields.
You can use `ignore` in [QueryParameterBinding](/grpc-api-gateway/reference/grpc/config/#queryparameterbinding) to remove
some proto fields from query parameters.

!!! example

    === "Configuration"
        ```yaml title="query_gateway.yaml" linenums="1" hl_lines="7"
        gateway:
          endpoints:
            - get: "/query"
              selector: "~.QueryService.Query"
              query_params:
                - selector: 'language'
                  ignore: true
        ```

    === "Proto Annotations"
        ```proto title="query.proto" linenums="1" hl_lines="6"
        service QueryService {
            rpc Query(QueryRequest) returns (QueryResponse) {
                option (meshapi.gateway.http) = {
                    get: "/query",
                    query_params: [
                        {selector: 'language', ignore: true},
                    ]
                };
            }
        }
        ```

    Proto field `language` gets ignored during automatic discovery phase and
    thus does not get bound to any query parameter.

### Data Types

This section addresses the mapping of query parameter data to proto types.
While mapping scalar types is straightforward, handling more complex data types
like repeated values and maps can be less intuitive.
Here, we explain how various types are parsed from query parameters.

Throughout this section, references to _scalar types_ include all numerical types, boolean, strings and enums.

#### Repeated Fields

Repeated fields can hold multiple values. Query parameters _do_ support repeated fields __only__ for the _scalar types_.

!!! example
    ```proto
    message Request {
        repeated string names = 1;
    }
    ```

    `?names=value1&names=value2&names=value3` gets mapped to `["value1", "value2", "value3"]`.

!!! info
    Commas in the value do not act as a separator and is instead read as part of the value.

    In the example above, `?names=value1,value2` gets mapped to `["value1,value2"]`.

#### Maps

Map types are also supported if both the key and the value are _scalar types_.
HTTP query parameter is `field_name[key]=value`.

!!! example
    ```proto
    message Request {
        map<string,string> metadata = 1;
    }
    ```

    `?metadata[key1]=value1&metadata[key2]=value2` gets mapped to `{"key1": "value1", "key2": "value2"}`.

#### Unbound Query Parameters

All unbound query parameters get ignored without generating any error.
