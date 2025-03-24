# Field Optionality & Nullability

In OpenAPI documents and JSON schemas, a field can specify its nullability and requiredness. It is crucial to accurately map the proto labels to the JSON schema to ensure a correct representation of your API. Each API can have its own unique set of patterns. In this document, you will learn how to configure the OpenAPI plug-in to adjust the way these specifications are mapped to proto labels.

## Field Nullability

In OpenAPI and JSON schemas, a field can specify whether it accepts null values. In OpenAPI v3.1, the correct way to represent this is by using type arrays.

!!! example
    ```json
    {
      "nullable_field": {
        "type": ["string", "null"]
      },
      "string_field": {
        "type": "string"
      }
    }
    ```

How does this map to proto fields? Let's consider the proto message below:

```proto
message User {
  string id = 1;
  string name = 2;
  optional string email_address = 3;
  optional Address address = 4;
  PhoneNumber phone_number = 5;
}
```

Protobuf can handle null values, where a null value indicates an unspecified value. If you assigned null values to every key in a JSON object for the proto message above, the model would be mapped as follows:

```json linenums="1"
{
  "id": "", // (1)!
  "name": "",
  "email_address": null, // (2)!
  "address": null, // (3)!
  "phone_number": null // (4)!
}
```

1. `id` is a string, a primitive type, and is not marked as optional. Therefore, it will be assigned its default value, which is an empty string.
2. `email_address` is a string, a primitive type, and is explicitly marked as optional. This means it can differentiate between having no value and being blank. Therefore, it is assigned `null`.
3. `address` is an optional message type. Similar to `email_address`, because it is labeled as `optional`, it is nullable and thus receives `null`.
4. Unlike `address`, `phone_number` is not marked as an optional field. However, since it is a message type, it is inherently nullable and thus also receives `null`.

That is how protobuf handles unmarshaling bytes into in-memory structures. How should this be represented in OpenAPI? If `phone_number` is intended to be set and a null value is unacceptable, this requirement must be clearly documented in the OpenAPI specification.

This can be controlled via `field_nullable_mode` plug-in option.

### 1. Disabled

Value of `disabled` for `field_nullable_mode` would enable this mode.

Disabled means that nullability will not be automatically documented. For the proto message above, this results in a JSON schema where none of the fields are marked as nullable.

You can manually adjust the nullability settings as needed:

Below is an example of explicitly setting the OpenAPI types via configuration files.

```yaml
openapi:
  messages:
    - selector: "User"
      fields:
        email_address:
          types:
            - STRING
            - "NULL"
```

This can also be achieved directly within proto files as well:

```proto linenums="1" hl_lines="5-7"
message User {
  string id = 1;
  string name = 2;
  optional string email_address = 3 [
    (meshapi.gateway.openapi_field) = {
		types: [STRING, NULL]
    }
  ];
  optional Address address = 4;
  PhoneNumber phone_number = 5;
}
```

### 2. Nullable If Optional (Default)

Value of `optional` for `field_nullable_mode` would enable this mode.

This mode treats any proto field marked as `optional` as nullable.

In the proto message above, the `email_address` and `address` fields would be represented in the OpenAPI document as accepting `null` values, while the other fields would not include `null` as an accepted value.

### 3. Nullable Unless Required

Value of `non_required` for `field_nullable_mode` would enable this mode.

This mode treats all fields as nullable _unless_ they are explicitly marked as required.

There are several ways to mark a field as required:

**1. At the message level**:

This approach mirrors how OpenAPI specifies required fields. Each model explicitly lists all properties that are required:

=== "Proto Annotations"
    ```proto linenums="1" hl_lines="4-6"
    import "meshapi/gateway/annotations.proto";

    message User {
      (meshapi.gateway.openapi_schema) = {
        required: ["id", "name", "phone_number"] // (1)!
      }

      string id = 1;
      string name = 2;
      string email_address = 3;
      Address address = 4;
      PhoneNumber phone_number = 5;
    }
    ```

    1. Indicates that fields `id`, `name` and `phone_number` do not accept null values.

=== "Configuration"
    ```yaml linenums="1"
    openapi:
      messages:
        - selector: "User"
          schema:
            required:
              - "id"
              - "name"
              - "phone_number"
    ```

**2. At the field level**:

This approach sets the requiredness at the field.

=== "Proto Annotations"
    ```proto linenums="1" hl_lines="4 5 9"
    import "meshapi/gateway/annotations.proto";

    message User {
      string id = 1 [(meshapi.gateway.openapi_field).config.required=true];
      string name = 2 [(meshapi.gateway.openapi_field).config.required=true];
      string email_address = 3;
      Address address = 4;
      PhoneNumber phone_number = 5 [
        (meshapi.gateway.openapi_field).config.required = true
      ];
    }
    ```

=== "Configuration"
    ```yaml linenums="1" hl_lines="7 10 13"
    openapi:
      messages:
        - selector: "User"
          fields:
            id:
              config:
                required: true
            name:
              config:
                required: true
            phone_number:
              config:
                required: true
    ```

## Optional/Required Annotations
