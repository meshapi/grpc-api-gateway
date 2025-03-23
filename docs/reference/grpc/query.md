# Query Parameter Binding

Understanding how query parameters are processed by the gRPC Gateway is crucial for customizing your API to meet specific needs.

### Default Behavior

In each [EndpointConfig](/grpc-api-gateway/reference/grpc/config), any proto field from the request message that is not already bound to path parameters or the HTTP body is automatically bound to query parameters.

#### Naming

By default, the names of these query parameters are derived from the proto field names. For nested fields, the name is constructed by concatenating the parent field name with the nested field name, separated by a dot.

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
    query parameters `some_input` and `options.case_sensitive` will be
    bound to the corresponding fields in the proto message.


### Customization

There are a number of customizations available.

#### Disable Automatic Discovery

In [EndpointConfig](/grpc-api-gateway/reference/grpc/config), you have the option to disable the automatic discovery and binding of query parameters. When this feature is disabled for a specific endpoint, only the query parameter bindings that are explicitly defined will be considered.


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

Some practical uses of this setting include:

* Assigning custom names to all query parameters.
* Restricting the exposure of certain parts of the proto message to the HTTP request.


#### Additional Query Parameter Binding and Aliases

To use custom names for query parameters or to allow multiple names (aliases) for the same message field, you can explicitly define the query parameter to request proto message field binding.

By utilizing `query_params` in [EndpointConfig](/grpc-api-gateway/reference/grpc/config), you can add multiple [QueryParameterBinding](/grpc-api-gateway/reference/grpc/config/#queryparameterbinding) objects.

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
    Assume we want to achieve the following:

    1. Allow both `language` and `lang` as query parameters, with `language` taking precedence.
    2. Use `per_page` as the query parameter name instead of the default `pagination.per_page`.

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

    * Both `language` and `lang` are accepted and bound to the `language` field in the proto request, with `language` taking precedence.
    * The `per_page` query parameter is accepted and bound to `pagination.per_page`.
    * The `term` field does not have an explicit binding but is automatically discovered and added.

!!! warning
    Once you add an explicit binding for a proto field, that specific field will no longer receive automatic bindings.
    Only the explicitly defined bindings will be applied.
    For instance, in the example above, `pagination.per_page` is no longer automatically bound because an
    explicit binding for `pagination.per_page` is defined.

##### Prioritization

If multiple names are defined for the same proto field, the most recently defined binding takes precedence over earlier ones.
In the example above, `language` takes precedence over `lang`.

#### Ignoring Parameters

You may want to exclude certain proto fields from being bound to any query parameters. To achieve this, you can use the `ignore` attribute in [QueryParameterBinding](/grpc-api-gateway/reference/grpc/config/#queryparameterbinding). This will prevent the specified proto fields from being included in the query parameters.

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

    Proto field `language` is excluded during the automatic discovery phase and, as a result, is not bound to any query parameter.

### Data Types

This section explains how query parameter data is mapped to proto types. While mapping scalar types is straightforward, handling more complex data types like repeated values and maps can be less intuitive. Here, we clarify how various types are parsed from query parameters.

Throughout this section, references to _scalar types_ include all numerical types, booleans, strings, and enums.

#### Repeated Fields

Repeated fields can hold multiple values. Query parameters support repeated fields only for scalar types.

!!! example
    ```proto
    message Request {
        repeated string names = 1;
    }
    ```

    `?names=value1&names=value2&names=value3` gets mapped to `["value1", "value2", "value3"]`.

!!! info
    Commas in the value do not act as a separator and are instead read as part of the value.

    In the example above, `?names=value1,value2` gets mapped to `["value1,value2"]`.

#### Maps

Map types are also supported if both the key and the value are scalar types. The HTTP query parameter format is `field_name[key]=value`.

!!! example
    ```proto
    message Request {
        map<string, string> metadata = 1;
    }
    ```

    `?metadata[key1]=value1&metadata[key2]=value2` gets mapped to `{"key1": "value1", "key2": "value2"}`.

#### Unbound Query Parameters

All unbound query parameters are ignored without generating any errors.
