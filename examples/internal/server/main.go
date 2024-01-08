package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/meshapi/grpc-rest-gateway/examples/api/echo"
	"github.com/meshapi/grpc-rest-gateway/gateway"
	"github.com/meshapi/grpc-rest-gateway/protoconvert"
	"github.com/meshapi/grpc-rest-gateway/utilities"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

var (
	filter_EchoService_Echo_0 = &utilities.DoubleArray{Encoding: map[string]int{"id": 0}, Base: []int{1, 2, 0, 0}, Check: []int{0, 1, 2, 2}}
)

func request_EchoService_Echo_0(ctx context.Context, marshaler gateway.Marshaler, client echo.EchoServiceClient, req *http.Request, params httprouter.Params) (proto.Message, gateway.ServerMetadata, error) {
	var protoReq echo.SimpleMessage
	var metadata gateway.ServerMetadata

	var (
		val string
		err error
		_   = err
	)

	val = params.ByName("id")
	if val == "" {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "id")
	}

	protoReq.Id, err = protoconvert.String(val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "id", err)
	}

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := gateway.PopulateQueryParameters(&protoReq, req.Form, filter_EchoService_Echo_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.Echo(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err
}

func RegisterEchoServiceHandlerClient(ctx context.Context, mux *gateway.ServeMux, client echo.EchoServiceClient) {
	mux.HandleWithParams(http.MethodPost, "/v1/examples/echo/:id", func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := mux.MarshalerForRequest(req)
		var err error
		var annotatedContext context.Context
		annotatedContext, err = gateway.AnnotateContext(ctx, mux, req, "/grpc.gateway.examples.internal.proto.examplepb.EchoService/Echo", gateway.WithHTTPPathPattern("/v1/example/echo/{id}"))
		if err != nil {
			gateway.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_EchoService_Echo_0(annotatedContext, inboundMarshaler, client, req, params)
		annotatedContext = gateway.NewServerMetadataContext(annotatedContext, md)
		if err != nil {
			gateway.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
			return
		}

		mux.ForwardResponseMessage(annotatedContext, outboundMarshaler, w, req, resp)
	})
}

func main() {
	listener, err := net.Listen("tcp", ":40000")
	if err != nil {
		log.Fatalf("failed to bind: %s", err)
	}

	server := grpc.NewServer()
	echo.RegisterEchoServiceServer(server, &EchoService{})
	reflection.Register(server)

	connection, err := grpc.Dial(":40000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to dial: %s", err)
	}

	restGateway := gateway.NewServeMux()
	RegisterEchoServiceHandlerClient(context.Background(), restGateway, echo.NewEchoServiceClient(connection))

	go func() {
		log.Printf("starting HTTP on port 4000...")
		if err := http.ListenAndServe(":4000", restGateway); err != nil {
			log.Printf("failed to start HTTP Rest Gateway service: %s", err)
		}
	}()

	log.Printf("starting gRPC on port 40000...")
	server.Serve(listener)
}
