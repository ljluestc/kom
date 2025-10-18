package example

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp"
	"k8s.io/klog/v2"
)

// TestMultiClusterSSE 测试多集群SSE模式
func TestMultiClusterSSE(t *testing.T) {
	// 创建测试用的kubeconfig文件
	testDir := "/tmp/kom-test-kubeconfigs"
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// 创建示例kubeconfig文件
	kubeconfig1 := createTestKubeconfig("cluster1", "https://cluster1.example.com:6443")
	kubeconfig2 := createTestKubeconfig("cluster2", "https://cluster2.example.com:6443")

	file1 := filepath.Join(testDir, "cluster1.yaml")
	file2 := filepath.Join(testDir, "cluster2.yaml")

	err = os.WriteFile(file1, []byte(kubeconfig1), 0644)
	if err != nil {
		t.Fatalf("Failed to write kubeconfig1: %v", err)
	}

	err = os.WriteFile(file2, []byte(kubeconfig2), 0644)
	if err != nil {
		t.Fatalf("Failed to write kubeconfig2: %v", err)
	}

	// 测试方式1：手动指定kubeconfig文件
	t.Run("ManualKubeconfigConfig", func(t *testing.T) {
		cfg := &mcp.ServerConfig{
			Name:    "kom multi-cluster test",
			Version: "0.0.1",
			Port:    9097, // 使用不同端口避免冲突
			Mode:    mcp.ServerModeSSE,
			Kubeconfigs: []mcp.KubeconfigConfig{
				{
					ID:        "cluster1",
					Path:      file1,
					IsDefault: true,
				},
				{
					ID:   "cluster2",
					Path: file2,
				},
			},
		}

		// 创建MCP服务器（不启动）
		server := mcp.GetMCPServerWithOption(cfg)
		if server == nil {
			t.Fatal("Failed to create MCP server")
		}

		// 验证集群已注册
		clusters := kom.Clusters().AllClusters()
		if len(clusters) < 2 {
			t.Fatalf("Expected at least 2 clusters, got %d", len(clusters))
		}

		// 验证集群信息
		cluster1 := kom.Clusters().GetClusterById("cluster1")
		if cluster1 == nil {
			t.Fatal("Cluster1 not found")
		}

		cluster2 := kom.Clusters().GetClusterById("cluster2")
		if cluster2 == nil {
			t.Fatal("Cluster2 not found")
		}

		t.Logf("Successfully registered clusters: %v", getClusterNames(clusters))
	})

	// 测试方式2：从目录自动加载
	t.Run("LoadFromDirectory", func(t *testing.T) {
		// 清理之前的集群
		kom.Clusters().RemoveClusterById("cluster1")
		kom.Clusters().RemoveClusterById("cluster2")

		// 从目录加载kubeconfig
		configs, err := mcp.LoadKubeconfigsFromDirectory(testDir)
		if err != nil {
			t.Fatalf("Failed to load kubeconfigs from directory: %v", err)
		}

		if len(configs) != 2 {
			t.Fatalf("Expected 2 kubeconfig configs, got %d", len(configs))
		}

		// 验证配置
		for _, config := range configs {
			if config.ID == "" {
				t.Error("Cluster ID should not be empty")
			}
			if config.Path == "" {
				t.Error("Kubeconfig path should not be empty")
			}
		}

		t.Logf("Successfully loaded kubeconfig configs: %v", configs)
	})

	// 测试方式3：使用kubeconfig内容
	t.Run("KubeconfigContent", func(t *testing.T) {
		// 清理之前的集群
		kom.Clusters().RemoveClusterById("cluster1")
		kom.Clusters().RemoveClusterById("cluster2")

		cfg := &mcp.ServerConfig{
			Name:    "kom multi-cluster content test",
			Version: "0.0.1",
			Port:    9098,
			Mode:    mcp.ServerModeSSE,
			Kubeconfigs: []mcp.KubeconfigConfig{
				{
					ID:        "cluster1",
					Content:   kubeconfig1,
					IsDefault: true,
				},
				{
					ID:      "cluster2",
					Content: kubeconfig2,
				},
			},
		}

		// 创建MCP服务器
		server := mcp.GetMCPServerWithOption(cfg)
		if server == nil {
			t.Fatal("Failed to create MCP server with kubeconfig content")
		}

		// 验证集群已注册
		clusters := kom.Clusters().AllClusters()
		if len(clusters) < 2 {
			t.Fatalf("Expected at least 2 clusters, got %d", len(clusters))
		}

		t.Logf("Successfully registered clusters from content: %v", getClusterNames(clusters))
	})
}

// TestDynamicClusterManagement 测试动态集群管理
func TestDynamicClusterManagement(t *testing.T) {
	// 清理现有集群
	allClusters := kom.Clusters().AllClusters()
	for clusterID := range allClusters {
		kom.Clusters().RemoveClusterById(clusterID)
	}

	// 创建测试kubeconfig
	kubeconfig := createTestKubeconfig("test-cluster", "https://test.example.com:6443")

	// 测试动态注册集群
	t.Run("DynamicRegister", func(t *testing.T) {
		// 直接测试集群注册功能
		_, err := kom.Clusters().RegisterByStringWithID(kubeconfig, "test-cluster")
		if err != nil {
			t.Fatalf("Failed to register cluster: %v", err)
		}

		// 验证集群已注册
		cluster := kom.Clusters().GetClusterById("test-cluster")
		if cluster == nil {
			t.Fatal("Cluster not found after registration")
		}

		t.Logf("Successfully registered cluster: %s", cluster.ID)
	})

	// 测试动态注销集群
	t.Run("DynamicUnregister", func(t *testing.T) {
		// 验证集群存在
		cluster := kom.Clusters().GetClusterById("test-cluster")
		if cluster == nil {
			t.Fatal("Cluster should exist before unregistration")
		}

		// 注销集群
		kom.Clusters().RemoveClusterById("test-cluster")

		// 验证集群已注销
		cluster = kom.Clusters().GetClusterById("test-cluster")
		if cluster != nil {
			t.Fatal("Cluster should not exist after unregistration")
		}

		t.Log("Successfully unregistered cluster")
	})
}

// TestMultiClusterToolUsage 测试多集群工具使用
func TestMultiClusterToolUsage(t *testing.T) {
	// 注册测试集群
	kubeconfig1 := createTestKubeconfig("cluster1", "https://cluster1.example.com:6443")
	kubeconfig2 := createTestKubeconfig("cluster2", "https://cluster2.example.com:6443")

	_, err := kom.Clusters().RegisterByStringWithID(kubeconfig1, "cluster1")
	if err != nil {
		t.Fatalf("Failed to register cluster1: %v", err)
	}

	_, err = kom.Clusters().RegisterByStringWithID(kubeconfig2, "cluster2")
	if err != nil {
		t.Fatalf("Failed to register cluster2: %v", err)
	}

	// 测试集群列表
	t.Run("ListClusters", func(t *testing.T) {
		clusters := kom.Clusters().AllClusters()
		if len(clusters) < 2 {
			t.Fatalf("Expected at least 2 clusters, got %d", len(clusters))
		}

		// 验证集群信息
		for clusterID, cluster := range clusters {
			if cluster.Config == nil {
				t.Errorf("Cluster %s config should not be nil", clusterID)
			}
			t.Logf("Cluster %s: %s", clusterID, cluster.Config.Host)
		}
	})

	// 测试默认集群选择
	t.Run("DefaultCluster", func(t *testing.T) {
		defaultCluster := kom.Clusters().DefaultCluster()
		if defaultCluster == nil {
			t.Fatal("Default cluster should not be nil")
		}

		t.Logf("Default cluster: %s", defaultCluster.ID)
	})

	// 清理
	kom.Clusters().RemoveClusterById("cluster1")
	kom.Clusters().RemoveClusterById("cluster2")
}

// 辅助函数

func createTestKubeconfig(clusterName, server string) string {
	return fmt.Sprintf(`apiVersion: v1
clusters:
- cluster:
    server: %s
    insecure-skip-tls-verify: true
  name: %s
contexts:
- context:
    cluster: %s
    user: admin
  name: %s
current-context: %s
kind: Config
users:
- name: admin
  user:
    token: test-token-%s`, server, clusterName, clusterName, clusterName, clusterName, clusterName)
}

func getClusterNames(clusters map[string]*kom.ClusterInst) []string {
	var names []string
	for name := range clusters {
		names = append(names, name)
	}
	return names
}

// mockCallToolRequest 模拟MCP工具请求
type mockCallToolRequest struct {
	params map[string]interface{}
}

func (m *mockCallToolRequest) GetString(key, defaultValue string) string {
	if val, ok := m.params[key].(string); ok {
		return val
	}
	return defaultValue
}

func (m *mockCallToolRequest) GetBoolean(key string, defaultValue bool) bool {
	if val, ok := m.params[key].(bool); ok {
		return val
	}
	return defaultValue
}

func (m *mockCallToolRequest) GetStringSlice(key string, defaultValue []string) []string {
	if val, ok := m.params[key].([]string); ok {
		return val
	}
	return defaultValue
}

// 基准测试
func BenchmarkMultiClusterRegistration(b *testing.B) {
	kubeconfig := createTestKubeconfig("benchmark-cluster", "https://benchmark.example.com:6443")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clusterID := fmt.Sprintf("cluster-%d", i)
		_, err := kom.Clusters().RegisterByStringWithID(kubeconfig, clusterID)
		if err != nil {
			b.Fatalf("Failed to register cluster: %v", err)
		}
		kom.Clusters().RemoveClusterById(clusterID)
	}
}

// 示例：如何在生产环境中使用多集群配置
func TestMultiClusterConfig(t *testing.T) {
	// 方式1：从环境变量读取kubeconfig目录
	kubeconfigDir := os.Getenv("KUBECONFIG_DIR")
	if kubeconfigDir == "" {
		kubeconfigDir = "/etc/kom/kubeconfigs"
	}

	// 从目录加载所有kubeconfig文件
	configs, err := mcp.LoadKubeconfigsFromDirectory(kubeconfigDir)
	if err != nil {
		klog.Errorf("Failed to load kubeconfigs: %v", err)
		return
	}

	// 创建服务器配置
	cfg := &mcp.ServerConfig{
		Name:        "kom multi-cluster server",
		Version:     "1.0.0",
		Port:        9096,
		Mode:        mcp.ServerModeSSE,
		Kubeconfigs: configs,
		AuthKey:     "username",
	}

	// 创建服务器（不启动，避免阻塞测试）
	server := mcp.GetMCPServerWithOption(cfg)
	if server == nil {
		t.Fatal("Failed to create MCP server")
	}
	
	t.Logf("Successfully created MCP server with %d kubeconfig configurations", len(configs))
}

// 示例：如何动态管理集群
func TestDynamicClusterMgmt(t *testing.T) {
	// 动态注册新集群
	newKubeconfig := `apiVersion: v1
clusters:
- cluster:
    server: https://new-cluster.example.com:6443
  name: new-cluster
contexts:
- context:
    cluster: new-cluster
    user: admin
  name: new-cluster
current-context: new-cluster
kind: Config
users:
- name: admin
  user:
    token: new-cluster-token`

	// 注册集群
	_, err := kom.Clusters().RegisterByStringWithID(newKubeconfig, "new-cluster")
	if err != nil {
		klog.Errorf("Failed to register new cluster: %v", err)
		return
	}

	// 使用集群
	kubectl := kom.Cluster("new-cluster")
	if kubectl != nil {
		// 获取集群信息
		cluster := kom.Clusters().GetClusterById("new-cluster")
		if cluster != nil && cluster.GetServerVersion() != nil {
			klog.Infof("Cluster version: %s", cluster.GetServerVersion().GitVersion)
		}
	}

	// 清理
	kom.Clusters().RemoveClusterById("new-cluster")
}
