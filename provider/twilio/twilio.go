// Package twilio is a placeholder for the Twilio Memory provider.
//
// The actual Twilio Memory provider implementation is in a separate module:
//
//	github.com/plexusone/omni-twilio/omnimemory
//
// This separation keeps the core omnimemory module lightweight and avoids
// pulling in the Twilio SDK as a dependency for all users.
//
// # Usage
//
// To use the Twilio Memory provider, import the omni-twilio package:
//
//	import (
//	    "github.com/plexusone/omnimemory"
//	    "github.com/plexusone/omnimemory/core"
//	    _ "github.com/plexusone/omni-twilio/omnimemory" // Register Twilio provider
//	)
//
//	func main() {
//	    client, err := omnimemory.NewClient(core.ClientConfig{
//	        Providers: []core.ProviderConfig{
//	            {
//	                Name: core.ProviderNameTwilio,
//	                Options: map[string]any{
//	                    "account_sid": "ACxxx",
//	                    "auth_token":  "xxx",
//	                },
//	            },
//	        },
//	    })
//	    // ...
//	}
//
// # Configuration
//
// The provider can be configured via options or environment variables:
//
//   - account_sid: Twilio Account SID (or TWILIO_ACCOUNT_SID env)
//   - auth_token: Twilio Auth Token (or TWILIO_AUTH_TOKEN env)
//
// # Concept Mapping
//
// Omnimemory concepts map to Twilio Memory API as follows:
//
//   - TenantID → Twilio Store ID
//   - SubjectID → Twilio Profile ID
//   - Memory → Twilio Observation
//   - Search/Recall → Twilio Recall API with semantic search
package twilio
