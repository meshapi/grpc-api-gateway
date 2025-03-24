## Introduction

The gRPC API Gateway patch feature allows you to update resources using HTTP PATCH requests. This is particularly useful for making partial updates to a resource without needing to send the entire resource representation. The PATCH feature can automatically generate a field mask that includes all the fields present in the request body.

#### Why is using an update mask necessary?

Without an update mask, null fields in the request body could either indicate that the API user wants to unset them or that they should remain unchanged. Using an update mask clarifies this ambiguity.

#### How does the PATCH feature help with this?

The PATCH feature simplifies the update process by automatically generating a field mask based on the fields present in the request body. This ensures that only the specified fields are updated, reducing ambiguity and making partial updates more efficient.

## How to Use

#### 1. Define Update Mask

In your proto request type, include a field named `update_mask` of type `google.protobuf.FieldMask`. This is crucial because the gateway requires this exact field name and type to populate them automatically.

#### 2. Define PATCH Method Binding

Only the `PATCH` HTTP method will trigger the automatic mapping of the field mask. Define the `PATCH` method in your Protobuf file or gateway configuration. While you can have additional bindings using other HTTP methods, the automatic mapping of the field mask will not occur with those methods.

Another important aspect that is needed to enable the automatic population of the field mask is the `body` annotation.
You must select the field subfield that contains the updates.

## Example

Below is an illustrative example demonstrating the use of the PATCH feature.

```proto title="service.proto" linenums="1" hl_lines="11 20"
import "meshapi/gateway/annotations.proto";
import "google/protobuf/field_mask.proto";

message Entity {
  string id = 1;
  string name = 2;
}

message Request {
  Entity entity = 1;
  google.protobuf.FieldMask update_mask = 2; // (1)!
}

message Response {}

service MyService {
    rpc MyMethod(Request) returns (Response) {
        option (meshapi.gateway.http) = {
            patch: "/my-endpoint/{entity.id}",
            body: "entity" // (2)!
        };
    }
}
```

1. The update mask must be named `update_mask` to be recognized by the gateway.
2. To enable automatic population of the `update_mask`, use the subfield that contains the partial updates, in this case, `entity`.

**Test the PATCH Endpoint**

Once you generate and implement your method, you can test the PATCH endpoint using tools like `curl` or Postman. Here is an example using `curl`.

```sh
curl -X PATCH -d '{"name": "New Name"}' http://localhost/my-endpoint/1
```

Notice that the `update_mask` is automatically set to `["name"]`, indicating that the `"name"` field in the request body should be updated.

!!! note
    The camelCase JSON name of the field is used to populate the update mask list.

!!! note
    Only the values bound to the body annotation will appear in the update mask. For example, in the example above, the entity ID would not be included in the list.


## Best Practices

- **Validation:** Ensure that you validate the incoming data to prevent invalid updates.
- **Partial Updates:** Only update the fields that are provided in the request to avoid overwriting existing data.
- **Error Handling:** Implement proper error handling to return meaningful error messages to the client.

By following these steps, you can effectively use the gRPC API Gateway patch feature to perform partial updates on your resources.
