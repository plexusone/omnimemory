// Package mem0 is a placeholder for the mem0 provider.
//
// The actual mem0 provider implementation is in a separate module:
//
//	github.com/plexusone/mem0-go/omnimemory
//
// This separation keeps the core omnimemory module lightweight and avoids
// pulling in the mem0 SDK as a dependency for all users.
//
// # Usage
//
// To use the mem0 provider, import the mem0-go package:
//
//	import (
//	    "github.com/plexusone/omnimemory"
//	    "github.com/plexusone/omnimemory/core"
//	    _ "github.com/plexusone/mem0-go/omnimemory" // Register mem0 provider
//	)
//
//	func main() {
//	    client, err := omnimemory.NewClient(core.ClientConfig{
//	        Providers: []core.ProviderConfig{
//	            {
//	                Name:   core.ProviderNameMem0,
//	                APIKey: os.Getenv("MEM0_API_KEY"),
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
//   - api_key: mem0 API key (or MEM0_API_KEY env)
//   - base_url: Custom base URL (optional, defaults to api.mem0.ai)
//
// # Concept Mapping
//
// Omnimemory concepts map to mem0 API as follows:
//
//   - TenantID → mem0 app_id
//   - SubjectID → mem0 user_id
//   - AgentID → mem0 agent_id
//   - SessionID → mem0 run_id
//   - Memory → mem0 MemoryItem
//   - Search/Recall → mem0 Search API
package mem0
