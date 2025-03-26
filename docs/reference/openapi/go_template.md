# Go Template

You can use [Go Templates](https://golangdocs.com/templates-in-golang) to generate documentation or parts of the OpenAPI specification based on the types defined in your proto files.

This feature is not enabled by default. To enable it, use the plug-in option `use_go_templates`.

Additionally, you can define custom values and use them in your templates by utilizing the `go_template_args` option.

### Where Can You Use Templates?

You can incorporate Go Templates into various sections of your OpenAPI documentation:

* Title
* Summary
* Description
* External Docs Documentation

You can define these templates in different contexts:

* Service
* Message
* Field
* Enum
* Method


### Template Arguments

To effectively use templates, you need to understand the available arguments and how to incorporate them into your templates.

#### Functions

There are some functions you can use in your Go Template.

##### Import

You can import another file and bring it in your documentation.

```proto linenums="1" hl_lines="2"
// Some existing text here.
// {{ import "user.tpl" }}
message User {
  string name = 1;
}
```

This function reads and evaluates the specified file as a Go template, replacing the second line with the resulting content.

##### Field Comments

Field Comments can look up the documentation for a field.

```proto linenums="1" hl_lines="3"
// Some existing text here.
// Name:
//   * {{ fieldcomments index(.Fields 0) }}
message User {
  string name = 1;
}
```

##### Arg

This function allows you to retrieve an argument provided via `go_template_args`.

For instance, assuming we have used `go_template_args` with value `url=http://mydomain.com/`

```proto linenums="1" hl_lines="3"
// Some existing text here.
//
// Read more about this type at {{ arg "url" }}/ref?name=User
message User {
  string name = 1;
}
```

#### Contexts

Go Templates can be used in different contexts. In each context, the variables are different and relevant to the context.

##### Enum

Go template for enum will have the following fields:

| Name | Type | Description |
| --- | --- | --- |
| File | `File` | The proto file that contains this enum.  |
| Name | `string` | Name of the enum.  |
| Value | list of [EnumVariant](#enum-variant) | Slice of enum variants. |

###### EnumVariant ######

| Name | Type | Description |
| --- | --- | --- |
| Name | `string` | Name of the enum value.  |
| Number | `int32` | Number of this enum value. |

!!! example
    ```proto
    // Subscription type holds multiple values:
    //
    // The following are accepted:
    // {{ range .Value }}
    // {{ .Name }} => {{ .Number }}
    // {{ end }}
    enum SubscriptionType {
      SUBSCRIPTION_TYPE_UNSPECIFIED = 0;
      SUBSCRIPTION_TYPE_PREMIUM = 1;
      SUBSCRIPTION_TYPE_FREE = 2;
    }
    ```

##### Message

Go template for message will have the following fields:

| Name | Type | Description |
| --- | --- | --- |
| File | `File` | The proto file that contains this message.  |
| Name | `string` | Name of the message.  |
| Fields | list of [Field](#field) | Slice of fields. |

!!! example
    ```proto
    // User is the user in the system.
    //
    // Following fields are available on the user:
    // {{ range .Fields }}
    // * {{ .Name }} (@ {{ .Number }}):
    //   {{ fieldcomments . }}
    // {{ end }}
    message User {
      // Name is the full name of the user.
      string name = 1;
      // age is the age of this user.
      string age = 2;
    }
    ```

##### Field

Go template for field will have the following fields:

| Name | Type | Description |
| --- | --- | --- |
| Name | `string` | Name of the enum value.  |
| Number | `int32` | The field tag number. |
| TypeName | `*string` | The type name. |
| JsonName | `*string` | The JSON name for this field. |
| Proto3Optional | `*bool` | Whether or not this has proto3 optional label. |
| Message | [Message](#message) | The message that contains this field. |

##### Service

Go template for service will have the following fields:

| Name | Type | Description |
| --- | --- | --- |
| File | `File` | The proto file that contains this service.  |
| Name | `string` | Name of the enum value.  |
| Methods | list of [Method](#method) | The RPC methods for this service. |

##### Binding / OpenAPI Operation

Bindings are unique compared to other contexts because they are not native concepts in Protobuf. A binding defines the relationship between an HTTP endpoint and a gRPC method. Multiple HTTP endpoints can be mapped to the same gRPC method. Each endpoint in OpenAPI can have a `description` and a `summary`. By default, the summary is the first paragraph, and the description is the entire comment section for the corresponding method. However, this can be customized as needed.

Thus, you can use the comment for each method, The Go Template however will have the [Binding](#binding) as its argument.

!!! example
    ```proto
    service BookService {
      // Accepts: {{ .Method.RequestType.Name }}
      // Returns: {{ .Method.ResponseType.Name }}
      //
      // Path: {{ .HTTPMethod }} {{ .PathTemplate }}
      // Params:
      // {{ range .PathParameters }}
      //   - {{ .FieldPath }}
      // {{ end }}
      rpc UpdateBook(UpdateBookRequest) returns (UpdateBookResponse) {
        option (meshapi.gateway.http) = {
          put: "/books/{id}",
          body: "*",
          additional_bindings: [
            {patch: "/books/{id}", body: "*"}
          ]
        }
      };
    }
    ```

###### Binding

| Name | Type | Description |
| --- | --- | --- |
| Method | [Method](#method) | gRPC Method for this binding. |
| PathTemplate | `string` | PathTemplate |
| HTTPMethod | `string` | HTTP Method (uppercase) |
| PathParameters | list of [PathParameter](#pathparameter) | The type name. |
| QueryParameters | list of [QueryParameter](#queryparameter) | The type name. |
| Body | [Body](#body) | Field path to the subfield that includes the body. |
| ResponseBody | [Body](#body) | Field path to the subfield that is the response body. |
| StreamConfig | [StreamConfig](#streamconfig) | Field path to the subfield that is the response body. |

###### PathParameter

| Name | Type | Description |
| --- | --- | --- |
| FieldPath | `string` | Path to a proto field in the request message. |
| Target | [Field](#field) | Field related to this path parameter in the request message. |

###### QueryParameter

| Name | Type | Description |
| --- | --- | --- |
| FieldPath | `string` | Path to a proto field in the request message. |
| Name | `string` | Name of the query parameter. |
| NameIsAlias | `bool` | Whether or not the name is an alias. |

###### Body

| Name | Type | Description |
| --- | --- | --- |
| FieldPath | `string` | Path to a proto field in the request message. |

###### StreamConfig

| Name | Type | Description |
| --- | --- | --- |
| AllowWebsocket | `bool` | Whether or not websocket is allowed for this binding. |
| AllowSSE | `bool` | Whether or not Server-Sent-Events (SSE) is allowed for this binding. |
| AllowChunkedTransfer | `bool` | Whether or not chunked transfer is allowed for this binding. |

###### Method

| Name | Type | Description |
| --- | --- | --- |
| Name | `string` | Name of the enum value.  |
| Service | [Service](#service) | The service containing this method. |
| RequestType | [Message](#message) | The request type. |
| ResponseType | [Message](#message) | The response type. |
| ClientStreaming | `*bool` | Identifies if client streams multiple messages. |
| ServerStreaming | `*bool` | Identifies if server streams multiple messages. |
| Bindings | lits of [Binding](#binding) | List of bindings for this method. |
