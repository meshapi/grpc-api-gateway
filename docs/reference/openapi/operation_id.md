# Operation IDs & Service Tags

OpenAPI operations are defined within the paths of the API specification and represent the various endpoints that can be called. Each operation must have a unique operation ID, which serves as an identifier for that specific operation. This uniqueness is crucial as it allows for precise referencing and avoids conflicts within the API documentation and client code generation.

Tags can also be assigned to operation objects to group them logically. These tags help in organizing the operations, making it easier to navigate and understand the API structure. Tags are typically used to categorize operations by their functionality or resource type, enhancing the readability and maintainability of the API documentation.

## Operation IDs

You have the flexibility to customize the format of automatically generated operation IDs using the OpenAPI plug-in, or you can choose to manually assign them for greater control and specificity.

### Manual Assignment

Similar to many other configurations, you can use configuration files or directly set this in the proto file:

=== "Proto Annotations"
    ```proto linenums="1" hl_lines="6"
    import "meshapi/gateway/annotations.proto";

    service EchoService {
      rpc Echo(EchoRequest) returns (EchoResponse) {
        option (meshapi.gateway.openapi_operation) = {
          operation_id: "custom_operation_id"
        };
      };
    }
    ```

=== "Configuration"
    ```yaml linenums="1" hl_lines="6"
    openapi:
      services:
        - selector: "EchoService"
          methods:
            Echo:
              operation_id: "custom_operation_id"
    ```

### Automatic Assignment

You can utilize the `operation_id_mode` plug-in option to define the format of the operation IDs.

For instance, if you have a method named `Echo` within a service called `EchoService` in the package `meshapi.examples.echo.v1`, the following formats can be generated:

| Operation ID Mode | Operation Name |
| --- | --- |
| `service+method` | `EchoService_Echo` |
| `method` | `Echo` |
| `fqn` | `meshapi.examples.echo.v1.EchoService.Echo` |

!!! info
    Multiple HTTP endpoints (operations) can be associated with a single gRPC method. In such cases, the additional bindings are appended with an incremental numerical suffix to distinguish them.

!!! warning
    When using the `method` operation id mode, be aware that there is a potential risk of generating duplicate operation IDs if multiple services share the same gRPC method name.

## Tags

By default, each operation is tagged with the service name, providing a clear and organized way to identify the operations. However, you have the option to override these tags to better categorize and group operations according to your specific needs. If you override the tags, this automatic generation will be skipped.

=== "Proto Annotations"
    ```proto linenums="1" hl_lines="6"
    import "meshapi/gateway/annotations.proto";

    service EchoService {
      rpc Echo(EchoRequest) returns (EchoResponse) {
        option (meshapi.gateway.openapi_operation) = {
          tags: ["tag1", "tag2"]
        };
      };
    }
    ```

    1. No automatic tag will be appended to this list.

=== "Configuration"
    ```yaml linenums="1" hl_lines="7-8"
    openapi:
      services:
        - selector: "EchoService"
          methods:
            Echo:
              tags:
                - "tag1"
                - "tag2"
    ```

#### Disable Automatic Service Tags

To prevent the automatic generation of tags based on the service name, use the `disable_service_tags` plug-in option.

#### Include Package Name in Tags

For services with identical names or if you prefer to include the full package name in the tags, utilize the `include_package_in_tags` option.
