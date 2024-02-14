package plugin

import (
	"context"
	"fmt"

	"github.com/meshapi/grpc-rest-gateway/api/codegen"
	"google.golang.org/protobuf/types/descriptorpb"
)

// GetGatewayConfigFile is a shortcut method to get gateway config file name.
func (c Client) GetGatewayConfigFile(
	ctx context.Context, file *descriptorpb.FileDescriptorProto) (string, error) {
	response, err := c.Gateway.GetGatewayConfigFile(ctx, &codegen.GetGatewayConfigFileRequest{ProtoFile: file})
	if err != nil {
		return "", fmt.Errorf("plugin failed processing GetGatewayConfigFile call: %w", err)
	}

	return response.FilePath, nil
}
