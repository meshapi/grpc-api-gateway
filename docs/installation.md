You can directly download the binaries from the [Releases](https://github.com/meshapi/grpc-api-gateway/releases) page.

The link above includes the proto files containing the annotations as well.

Package repositories below also contain the binaries:

#### Arch

This plug-in is not part of the official repository but there is a package in the user repository that can be used:

[protoc-gen-grpc-api-gateway-bin](https://aur.archlinux.org/packages/protoc-gen-grpc-api-gateway-bin)

This can be installed using `makepkg`:

```sh
$ git clone https://aur.archlinux.org/protoc-gen-grpc-api-gateway-bin.git
$ cd protoc-gen-grpc-api-gateway-bin && makepkg -si
```

#### Alpine

This is an on-going effort and should be available soon, at the moment, you can download the binaries from
[Releases](https://github.com/meshapi/grpc-api-gateway/releases) page or install from source:

#### Install from source

If the binaries are not available for your operating system or architecture,
you can install from the source using Go

```sh
$ go install github.com/meshapi/grpc-api-gateway@<version>
```

### Docker

If you wish to install this tool inside docker, you can use `wget` or `curl` to download the binaries for the intended architecture. Take a look at the [Releases](https://github.com/meshapi/grpc-api-gateway/releases).
