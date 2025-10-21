# MCP (Microservice Control Panel) User Guide

[English](README_en.md) | [中文](README.md)

## Getting Started

### Requirements
- Go 1.16 or higher
- Configured Kubernetes cluster
- Default Kubeconfig file

### Basic Start Command
```go
mcp.RunMCPServer("kom mcp server", "0.0.1", 3619)
```

### Multi-Cluster Configuration
```go
// Method 1: Manual kubeconfig specification
kubeconfigs := []mcp.KubeconfigConfig{
    {
        ID:        "production",
        Path:      "/path/to/production-kubeconfig.yaml",
        IsDefault: true,
    },
    {
        ID:   "staging",
        Path: "/path/to/staging-kubeconfig.yaml",
    },
}

cfg := mcp.ServerConfig{
    Name:        "kom mcp server",
    Version:     "0.0.1",
    Port:        9096,
    Mode:        mcp.ServerModeSSE,
    Kubeconfigs: kubeconfigs,
}
mcp.RunMCPServerWithOption(&cfg)
```

### Directory Auto-Discovery
```go
// Method 2: Load all kubeconfig files from directory
kubeconfigs, err := mcp.LoadKubeconfigsFromDirectory("/path/to/kubeconfigs")
if err != nil {
    log.Fatal(err)
}

cfg := mcp.ServerConfig{
    Mode:        mcp.ServerModeSSE,
    Kubeconfigs: kubeconfigs,
}
mcp.RunMCPServerWithOption(&cfg)
```

## Features

- **Multi-cluster Management**: Support managing multiple Kubernetes clusters simultaneously in SSE mode
- **Dynamic Cluster Registration**: Register and unregister clusters at runtime without restart
- **Flexible Configuration**: Support kubeconfig file paths, content, and directory auto-discovery
- **Dynamic Resource Operations**: Support CRUD operations on various Kubernetes resources
- **Event Monitoring**: Real-time cluster event viewing
- **Resource Description**: Get detailed resource information
- **SSE Mode**: Full Server-Sent Events support for real-time communication

## API Interfaces

### Cluster Management

#### List Clusters
- API Name: `list_k8s_clusters`
- Description: List all registered Kubernetes clusters with detailed information
- Parameters: None
- Returns: Cluster list containing cluster names, hosts, and versions

#### Register Cluster
- API Name: `register_k8s_cluster`
- Description: Dynamically register a new Kubernetes cluster
- Parameters:
  - cluster_id: Unique identifier for the cluster
  - kubeconfig_path: Path to the kubeconfig file (optional)
  - kubeconfig_content: Kubeconfig content (optional, alternative to path)
  - is_default: Whether to set as default cluster
- Returns: Registration status and cluster information

#### Unregister Cluster
- API Name: `unregister_k8s_cluster`
- Description: Unregister a Kubernetes cluster
- Parameters:
  - cluster_id: Cluster identifier to unregister
- Returns: Unregistration status

### Dynamic Resource Operations

#### List Resources
- API Name: `list_k8s_resource`
- Description: List Kubernetes resources by cluster and resource type
- Parameters:
  - cluster: Cluster where the resources are running (use empty string for default cluster)
  - namespace: Namespace of the resources (optional for cluster-scoped resources)
  - group: API group of the resource
  - version: API version of the resource
  - kind: Kind of the resource
  - label: Label selector to filter resources (e.g. app=k8m)

#### Get Resource
- API Name: `get_k8s_resource`
- Description: Get a specific Kubernetes resource
- Parameters:
  - cluster: Cluster name
  - namespace: Namespace
  - group: API group
  - version: API version
  - kind: Resource kind
  - name: Resource name

#### Delete Resource
- API Name: `delete_k8s_resource`
- Description: Delete a specific Kubernetes resource
- Parameters:
  - cluster: Cluster name
  - namespace: Namespace
  - group: API group
  - version: API version
  - kind: Resource kind
  - name: Resource name

### Event Monitoring

#### List Events
- API Name: `list_k8s_event`
- Description: List Kubernetes events by cluster and namespace
- Parameters:
  - cluster: Cluster where the events are running (use empty string for default cluster)
  - namespace: Namespace of the events (optional)
  - involvedObjectName: Filter events by involved object name

## Usage Examples

### List All Clusters
```json
{
  "tool": "list_k8s_clusters"
}
```

### Register New Cluster
```json
{
  "tool": "register_k8s_cluster",
  "params": {
    "cluster_id": "new-cluster",
    "kubeconfig_path": "/path/to/kubeconfig.yaml",
    "is_default": false
  }
}
```

### Unregister Cluster
```json
{
  "tool": "unregister_k8s_cluster",
  "params": {
    "cluster_id": "old-cluster"
  }
}
```

### List All Pods in Default Namespace
```json
{
  "tool": "list_k8s_resource",
  "params": {
    "cluster": "",
    "namespace": "default",
    "group": "",
    "version": "v1",
    "kind": "Pod"
  }
}
```

### View Events for a Specific Pod
```json
{
  "tool": "list_k8s_event",
  "params": {
    "cluster": "",
    "namespace": "default",
    "involvedObjectName": "my-pod"
  }
}
```

## Multi-Cluster Configuration Options

### Method 1: Manual Configuration
Specify each cluster individually with file paths or content.

### Method 2: Directory Auto-Discovery
Automatically load all kubeconfig files from a directory.

### Method 3: Mixed Approach
Combine manual configuration with directory loading.

## Notes

1. **Multi-cluster Support**: All existing MCP tools work seamlessly with multi-cluster setup
2. **Cluster Identification**: Clusters are identified by their unique ID
3. **Default Cluster**: When no specific cluster is specified, operations use the default cluster
4. **Dynamic Management**: Clusters can be registered/unregistered at runtime without server restart
5. **Error Handling**: Network errors and invalid configurations are handled gracefully
6. **Resource Operations**: When using dynamic resource operations, ensure to provide correct resource group, version, and kind information
7. **Cluster-scoped Resources**: For cluster-scoped resources, the namespace parameter can be omitted
8. **Label Selectors**: When using label selectors, ensure to use the correct label format

## AI Tool Integration

### Claude Desktop
1. Open Claude Desktop settings panel
2. Add MCP Server address in the API configuration area
3. Enable SSE event listening function
4. Verify connection status

### Cursor
1. Enter Cursor settings interface
2. Find extension service configuration option
3. Add MCP Server URL (e.g., http://localhost:3619/sse)
4. Enable real-time event notifications

### Windsurf
1. Access configuration center
2. Set API server address
3. Enable real-time event notifications
4. Test connection

### Common Issues
1. Ensure MCP Server is running and port is accessible
2. Check if network connection is normal
3. Verify if SSE connection is established successfully
4. Check tool logs to troubleshoot connection issues

## Contributing

Issues and Pull Requests are welcome to help improve MCP.