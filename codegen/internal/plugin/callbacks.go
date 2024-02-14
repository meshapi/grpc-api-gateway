package plugin

// Callback provides information about a callback, can be used to generate documentation.
type Callback struct {
	// Description holds a description for the callback.
	Description string
}

// Callback constants
const (
	CallbackGatewayConfigFile = "gateway:config_file"
)

// CallbackRegistry holds a table of callback constants to their info.
type CallbackRegistry map[string]Callback

// RestGatewayCallbacks returns all available rest gateway callbacks.
func RestGatewayCallbacks() CallbackRegistry {
	return CallbackRegistry{
		CallbackGatewayConfigFile: Callback{
			Description: "explicitly set gateway config file for each proto file that contains a service declaration.",
		},
	}
}
