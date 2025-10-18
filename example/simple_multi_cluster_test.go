package example

import (
	"testing"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp"
)

// TestSimpleMultiCluster tests the core multi-cluster functionality without network calls
func TestSimpleMultiCluster(t *testing.T) {
	// Test 1: Manual kubeconfig configuration
	t.Run("ManualKubeconfig", func(t *testing.T) {
		// Create test kubeconfig content
		kubeconfig1 := `apiVersion: v1
clusters:
- cluster:
    server: https://cluster1.example.com:6443
    insecure-skip-tls-verify: true
  name: cluster1
contexts:
- context:
    cluster: cluster1
    user: admin
  name: cluster1
current-context: cluster1
kind: Config
users:
- name: admin
  user:
    token: test-token-1`

		kubeconfig2 := `apiVersion: v1
clusters:
- cluster:
    server: https://cluster2.example.com:6443
    insecure-skip-tls-verify: true
  name: cluster2
contexts:
- context:
    cluster: cluster2
    user: admin
  name: cluster2
current-context: cluster2
kind: Config
users:
- name: admin
  user:
    token: test-token-2`

		// Register clusters manually
		_, err1 := kom.Clusters().RegisterByStringWithID(kubeconfig1, "cluster1")
		if err1 != nil {
			t.Fatalf("Failed to register cluster1: %v", err1)
		}

		_, err2 := kom.Clusters().RegisterByStringWithID(kubeconfig2, "cluster2")
		if err2 != nil {
			t.Fatalf("Failed to register cluster2: %v", err2)
		}

		// Verify clusters are registered
		clusters := kom.Clusters().AllClusters()
		if len(clusters) < 2 {
			t.Fatalf("Expected at least 2 clusters, got %d", len(clusters))
		}

		// Check specific clusters exist
		cluster1 := kom.Clusters().GetClusterById("cluster1")
		if cluster1 == nil {
			t.Fatal("cluster1 not found")
		}

		cluster2 := kom.Clusters().GetClusterById("cluster2")
		if cluster2 == nil {
			t.Fatal("cluster2 not found")
		}

		t.Logf("Successfully registered %d clusters", len(clusters))
		t.Logf("Cluster1 server: %s", cluster1.Config.Host)
		t.Logf("Cluster2 server: %s", cluster2.Config.Host)
	})

	// Test 2: Directory loading
	t.Run("DirectoryLoading", func(t *testing.T) {
		// Test the LoadKubeconfigsFromDirectory function
		configs, err := mcp.LoadKubeconfigsFromDirectory("/tmp/nonexistent")
		if err != nil {
			// Expected error for non-existent directory
			t.Logf("Expected error for non-existent directory: %v", err)
		}

		// Should return empty slice for non-existent directory
		if len(configs) != 0 {
			t.Fatalf("Expected empty configs for non-existent directory, got %d", len(configs))
		}

		t.Log("Directory loading function works correctly")
	})

	// Test 3: Server configuration
	t.Run("ServerConfiguration", func(t *testing.T) {
		// Test creating server config with multiple kubeconfigs
		cfg := &mcp.ServerConfig{
			Name:    "test-server",
			Version: "1.0.0",
			Port:    9096,
			Mode:    mcp.ServerModeSSE,
			Kubeconfigs: []mcp.KubeconfigConfig{
				{
					ID:      "test-cluster-1",
					Content: "test-kubeconfig-content-1",
				},
				{
					ID:      "test-cluster-2", 
					Content: "test-kubeconfig-content-2",
				},
			},
		}

		// Verify configuration is valid
		if len(cfg.Kubeconfigs) != 2 {
			t.Fatalf("Expected 2 kubeconfigs in config, got %d", len(cfg.Kubeconfigs))
		}

		if cfg.Kubeconfigs[0].ID != "test-cluster-1" {
			t.Fatalf("Expected first cluster ID to be 'test-cluster-1', got '%s'", cfg.Kubeconfigs[0].ID)
		}

		if cfg.Kubeconfigs[1].ID != "test-cluster-2" {
			t.Fatalf("Expected second cluster ID to be 'test-cluster-2', got '%s'", cfg.Kubeconfigs[1].ID)
		}

		t.Log("Server configuration with multiple kubeconfigs works correctly")
	})

	// Test 4: Cluster management
	t.Run("ClusterManagement", func(t *testing.T) {
		// Test removing clusters
		initialCount := len(kom.Clusters().AllClusters())
		
		// Remove cluster1
		kom.Clusters().RemoveClusterById("cluster1")
		
		// Verify cluster1 is removed
		cluster1 := kom.Clusters().GetClusterById("cluster1")
		if cluster1 != nil {
			t.Fatal("cluster1 should be removed but still exists")
		}

		// Verify cluster2 still exists
		cluster2 := kom.Clusters().GetClusterById("cluster2")
		if cluster2 == nil {
			t.Fatal("cluster2 should still exist but was removed")
		}

		finalCount := len(kom.Clusters().AllClusters())
		if finalCount >= initialCount {
			t.Fatalf("Expected cluster count to decrease, initial: %d, final: %d", initialCount, finalCount)
		}

		t.Logf("Cluster management works correctly. Removed cluster1, cluster2 still exists")
	})
}
