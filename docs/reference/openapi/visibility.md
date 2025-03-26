# Visibility Selectors

## API Visibility Restrictions in gRPC and Protobuf

When designing APIs with gRPC and Protobuf, it's important to manage the visibility of your services and messages to ensure proper access control and versioning. Here are some key points to consider:

- **Public APIs**: These are accessible to all clients and should be stable and well-documented. Public APIs are typically used for external integrations and third-party developers.
- **Internal APIs**: These are intended for use within your organization and can be more flexible in terms of changes. Internal APIs should still be documented but may not require the same level of stability as public APIs.
- **Experimental APIs**: These are used for testing new features and may change frequently. Experimental APIs should be clearly marked and should not be relied upon for production use.

To enforce these visibility restrictions, you can use Protobuf options and annotations to specify the intended audience and stability level of your APIs.

```proto
import "google/api/visibility.proto";

service InternalService {
  option (google.api.api_visibility).restriction = "INTERNAL";
}

service Service {
  rpc NewExperimentalMethod(Request) returns (Response) {
    option (google.api.api_visibility).restriction = "EXPERIMENTAL";
  };
}
```

By default, all services and methods are considered. If the `visibility_selectors` plug-in option is left blank, visibility restrictions are ignored. However, you can provide a comma-separated list of labels to be considered. In this case, only the services and methods that either do not specify any visibility restriction or have at least one of the labels from the visibility selectors will be considered.
