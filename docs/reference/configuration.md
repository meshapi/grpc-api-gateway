# Configuration

Significant effort has gone into designing a flexible configuration system with sensible defaults.

Both the _gRPC API Gateway_ and _OpenAPI v3.1_ plug-ins offer extensive customization options through configurations embedded directly in `.proto` files or through separate configuration files in either `YAML` or `JSON` format.

Throughout this documentation, you will find examples provided in both proto annotations and `YAML` configuration file formats.

## Mix and Match

You can mix and match these two methods, setting some configurations in `proto` files and using separate configuration files. However, there are some rules to consider:

!!! note
    When both proto files and configuration files set an option:

    * __If the value is an _object_ or an _array_ type, the result is a merge of both settings.__

        ??? example
            If `schema.description` is set in a proto file for a message type and `schema.summary` is set in a configuration file for the same message, the result would contain both `summary` and `description`.
    
    * __If the value is a simple type such as _string_, _boolean_ or _number_ the configuration file takes precedence.__

        ??? example
            _For instance_: if `schema.description` is set in a proto file for a message type and a configuration file also sets `schema.description` for the same message, the value from the configuration file is used.

## Using proto annotations

When using proto annotations, you will need to import the proto annotations and types for the `gRPC API Gateway`.

[Buf](https://buf.build/) is a tool that simplifies the development and consumption of the Protobuf APIs.
It manages dependencies and builds proto files efficiently.

All proto files and annotations are available on [buf.build](https://buf.build/meshapi/grpc-api-gateway).

If you decide to use Buf, follow the instructions below or you can visit the `protoc` tab for
instructions on using protoc.

=== "Using Buf"

    Let's create a `buf.gen.yaml` file if you do not already have one
    with the following content or add `buf.build/meshapi/grpc-api-gateway` in your dependencies if you have an existing one:

    ```yaml title="buf.yaml" linenums="1"
    version: v1
    deps:
      - "buf.build/meshapi/grpc-api-gateway"
    ``` 

    Update mods to download the proto files:

    ```sh
    $ buf mod update
    ```

=== "Using protoc"

    You will first need to download the proto files for `gRPC API Gateway`.
    File named `grpc_api_gateway_proto.tar.gz` in the [Releases](https://github.com/meshapi/grpc-api-gateway/releases) page contains all the necessary proto files.

    From now on, use the `-I` or `--proto-path` option to include these proto files if they reside outside of the proto search path.


In any proto file you wish to use annotations, use the import line below when wanting to use gateway or openapi options:

```proto
import "meshapi/gateway/annotations.proto";
```


## Using configuration files

Each proto file can have a single accompanying configuration file in _YAML_ or _JSON_ format
that gets loaded along with the proto file. As the plug-in processes each proto file, it will use the _config file pattern_
setting to determine a configuration file name and will try three file extensions `.yaml`, `.yml` and `.json` in that order.
If a file exists, the configuration file gets loaded and the search ends.

!!! note
    Both proto annotations and configurations files offer the same options in different formats. Thus there is no option
    that can be set using one method that cannot be set with the other.

### Search Path

Since _protoc_ compiler does not provide the full path to the proto files, it is __NOT__ possible for the plug-ins to
know the exact path of each proto file and cannot determine which directory to search for the configuration file.
To address this issue, you can set _search paths_ in the command line options for either plug-in to set
the root directory where these configurations live.

Search paths essentially set the root directory that will be used to search for configuration files.
The default value will work for the majority of the use cases. However, if you want to place configuration files
in a separate directory than your proto files or your file structures are more complex, you can use the _search paths_
to direct the plug-ins where they should search.

__Default Value__: default value is always `.` which means the current working directory of the `protoc` compiler.

??? example

    Imagine the following structure:

    ```
    project/
        proto/
            models.proto
    ```

    Assuming `protoc` is callled from the `project` directory:

    * If _search path_ is `.` (default value): directory `proto/` will be searched for the relative configuration file.
    * If _search path_ is `configs`, directory `configs/proto/` will be searched for the relative configuration file.

### Filename Pattern

We have discussed the search path and the directory that will be used to search for finding a
relative configuration file for proto files. What is the file name? This setting can be used using command line flags in
either plug-in to specify a convention based on the name of the proto file. To remain flexible, this value is a Go template
value.

!!! note
    This name must __NOT__ contain the file extension since the tool
    itself will append and look for `.yaml`, `yml` and `.json` extensions in that order.

__Default Value__: default value is always `{{ .Path }}_gateway` for both OpenAPI and API Gateway configuration files.

This is a default but it can be changed according to your preferred file organization pattern.

#### Go Template Values

The following values are available to use in the configuration filename pattern.

The value column shows the value for an example proto file `proto/myservice/v1/model.proto`.

| Expression      | Description                          | Value |
| ----------- | ------------------------------------ | ------- |
| `{{.Name}}` | is the base file name (no parent directorry) excluding `.proto` extension | `model` |
| `{{.Path}}` | is the relative path to the file but excluding the `.proto` extension | `proto/myservice/v1/model` |
| `{{.Dir}}`  | is the relative path to the parent directory of the related proto file | `proto/myservice/v1` |

!!! example
    If _search path_ is `/path/to/configs` and _file pattern_ is `{{.Name}}_gateway`. Associated configuration
    file name (omitting file extension) for proto file `cool_service/v1/service.proto` is
    `/path/to/configs/cool_service/v1/service_gateway`.

!!! warning
    Each `.proto` file can be associated with a single configuration file,
    meaning only one configuration will be loaded per .proto file.
    However, the same configuration file can be used by multiple `.proto` files if the configuration
    file pattern is customized (e.g. using `{{.Dir}}` expression).

It is truly a matter of personal preference which method you would like to use to customize
the gateway and/or the OpenAPI objects. It might be worth noting the following:

* A [JSON schema](https://json.schemastore.org/grpc-api-gateway.json) exists for YAML/JSON files so you benefit from autocompletion
if you have installed the proper YAML/JSON extension.

* With many customization, proto files can get bloated. Separating the proto definitions from the gateway and OpenAPI configurations
can help with the organization of files.
