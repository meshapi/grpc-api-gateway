You can directly download the binaries from the [Releases](https://github.com/meshapi/grpc-api-gateway/releases) page, which also includes the proto files containing the annotations.

Additionally, you can find the binaries in the package repositories listed below:

#### Arch Linux

This plug-in is not part of the official repository, but you can use a package from the user repository (AUR):

[protoc-gen-grpc-api-gateway-bin](https://aur.archlinux.org/packages/protoc-gen-grpc-api-gateway-bin)

To install it using `makepkg`, follow these steps:

```sh
$ git clone https://aur.archlinux.org/protoc-gen-grpc-api-gateway-bin.git
$ cd protoc-gen-grpc-api-gateway-bin && makepkg -si
```

#### Alpine Linux

This is an ongoing effort and should be available soon. In the meantime, you can download the binaries from the [Releases](https://github.com/meshapi/grpc-api-gateway/releases) page or install from source.

#### Install from Source

If binaries are not available for your operating system or architecture, you can install from the source using Go:

```sh
$ go install github.com/meshapi/grpc-api-gateway/codegen/cmd/protoc-gen-openapiv3@<version>
$ go install github.com/meshapi/grpc-api-gateway/codegen/cmd/protoc-gen-grpc-api-gateway@<version>
```

To retrieve the latest version, replace `<version>` with `latest`.

### Docker

To install this tool inside a Docker container, you can use `wget` or `curl` to download the binaries for your intended architecture. Refer to the [Releases](https://github.com/meshapi/grpc-api-gateway/releases) page for more details.
