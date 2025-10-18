package example

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/weibaohui/kom/mcp"
	"k8s.io/klog/v2"
)

func TestMultiClusterDemo(t *testing.T) {
	// 设置日志级别
	klog.InitFlags(nil)
	flag.Set("v", "2")

	// 示例1：手动指定kubeconfig文件
	fmt.Println("=== 示例1：手动指定kubeconfig文件 ===")
	demoManualKubeconfig(t)

	// 示例2：从目录自动加载
	fmt.Println("\n=== 示例2：从目录自动加载 ===")
	demoLoadFromDirectory(t)

	// 示例3：使用kubeconfig内容
	fmt.Println("\n=== 示例3：使用kubeconfig内容 ===")
	demoKubeconfigContent(t)
}

// demoManualKubeconfig 演示手动指定kubeconfig文件
func demoManualKubeconfig(t *testing.T) {
	// 创建测试kubeconfig文件
	testDir := "/tmp/kom-demo-kubeconfigs"
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// 创建示例kubeconfig文件
	kubeconfig1 := createDemoKubeconfig("production", "https://prod.example.com:6443")
	kubeconfig2 := createDemoKubeconfig("staging", "https://staging.example.com:6443")

	file1 := testDir + "/production.yaml"
	file2 := testDir + "/staging.yaml"

	err = os.WriteFile(file1, []byte(kubeconfig1), 0644)
	if err != nil {
		t.Fatalf("Failed to write kubeconfig1: %v", err)
	}

	err = os.WriteFile(file2, []byte(kubeconfig2), 0644)
	if err != nil {
		t.Fatalf("Failed to write kubeconfig2: %v", err)
	}

	// 配置多集群
	cfg := &mcp.ServerConfig{
		Name:    "kom multi-cluster demo",
		Version: "1.0.0",
		Port:    9096,
		Mode:    mcp.ServerModeSSE,
		Kubeconfigs: []mcp.KubeconfigConfig{
			{
				ID:        "production",
				Path:      file1,
				IsDefault: true,
			},
			{
				ID:   "staging",
				Path: file2,
			},
		},
	}

	// 创建MCP服务器（不启动）
	server := mcp.GetMCPServerWithOption(cfg)
	if server == nil {
		t.Fatal("Failed to create MCP server")
	}

	fmt.Printf("✅ 成功配置了 %d 个集群\n", len(cfg.Kubeconfigs))
	for _, kubeconfig := range cfg.Kubeconfigs {
		fmt.Printf("  - 集群ID: %s, 文件: %s, 默认: %v\n", 
			kubeconfig.ID, kubeconfig.Path, kubeconfig.IsDefault)
	}
}

// demoLoadFromDirectory 演示从目录自动加载
func demoLoadFromDirectory(t *testing.T) {
	// 创建测试目录和文件
	testDir := "/tmp/kom-demo-auto"
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// 创建多个kubeconfig文件
	clusters := []string{"dev", "test", "prod"}
	for _, cluster := range clusters {
		kubeconfig := createDemoKubeconfig(cluster, fmt.Sprintf("https://%s.example.com:6443", cluster))
		file := fmt.Sprintf("%s/%s.yaml", testDir, cluster)
		err = os.WriteFile(file, []byte(kubeconfig), 0644)
		if err != nil {
			t.Fatalf("Failed to write kubeconfig for %s: %v", cluster, err)
		}
	}

	// 从目录加载kubeconfig
	configs, err := mcp.LoadKubeconfigsFromDirectory(testDir)
	if err != nil {
		t.Fatalf("Failed to load kubeconfigs from directory: %v", err)
	}

	fmt.Printf("✅ 从目录 %s 自动加载了 %d 个kubeconfig文件\n", testDir, len(configs))
	for _, config := range configs {
		fmt.Printf("  - 集群ID: %s, 文件: %s\n", config.ID, config.Path)
	}
}

// demoKubeconfigContent 演示使用kubeconfig内容
func demoKubeconfigContent(t *testing.T) {
	kubeconfigContent := `apiVersion: v1
clusters:
- cluster:
    server: https://demo.example.com:6443
    insecure-skip-tls-verify: true
  name: demo-cluster
contexts:
- context:
    cluster: demo-cluster
    user: admin
  name: demo-cluster
current-context: demo-cluster
kind: Config
users:
- name: admin
  user:
    token: demo-token-12345`

	cfg := &mcp.ServerConfig{
		Name:    "kom kubeconfig content demo",
		Version: "1.0.0",
		Port:    9097,
		Mode:    mcp.ServerModeSSE,
		Kubeconfigs: []mcp.KubeconfigConfig{
			{
				ID:        "demo-cluster",
				Content:   kubeconfigContent,
				IsDefault: true,
			},
		},
	}

	// 创建MCP服务器
	server := mcp.GetMCPServerWithOption(cfg)
	if server == nil {
		t.Fatal("Failed to create MCP server with kubeconfig content")
	}

	fmt.Printf("✅ 成功使用kubeconfig内容配置了集群: %s\n", cfg.Kubeconfigs[0].ID)
	fmt.Printf("  - 服务器地址: %s\n", "https://demo.example.com:6443")
	fmt.Printf("  - 是否为默认集群: %v\n", cfg.Kubeconfigs[0].IsDefault)
}

// createDemoKubeconfig 创建演示用的kubeconfig内容
func createDemoKubeconfig(clusterName, server string) string {
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
    token: %s-token-12345`, server, clusterName, clusterName, clusterName, clusterName, clusterName)
}

// 演示如何使用MCP工具管理多集群
func demoMCPTools() {
	fmt.Println("\n=== MCP工具使用示例 ===")
	
	// 列出所有集群
	fmt.Println("1. 列出所有集群:")
	fmt.Println(`   {
     "method": "tools/call",
     "params": {
       "name": "list_k8s_clusters"
     }
   }`)

	// 注册新集群
	fmt.Println("\n2. 注册新集群:")
	fmt.Println(`   {
     "method": "tools/call",
     "params": {
       "name": "register_k8s_cluster",
       "arguments": {
         "cluster_id": "new-cluster",
         "kubeconfig_path": "/path/to/kubeconfig.yaml",
         "is_default": false
       }
     }
   }`)

	// 获取指定集群的Pod列表
	fmt.Println("\n3. 获取指定集群的Pod列表:")
	fmt.Println(`   {
     "method": "tools/call",
     "params": {
       "name": "list_k8s_pods",
       "arguments": {
         "cluster": "production",
         "namespace": "default"
       }
     }
   }`)

	// 使用默认集群
	fmt.Println("\n4. 使用默认集群:")
	fmt.Println(`   {
     "method": "tools/call",
     "params": {
       "name": "list_k8s_pods",
       "arguments": {
         "namespace": "default"
       }
     }
   }`)

	// 注销集群
	fmt.Println("\n5. 注销集群:")
	fmt.Println(`   {
     "method": "tools/call",
     "params": {
       "name": "unregister_k8s_cluster",
       "arguments": {
         "cluster_id": "old-cluster"
       }
     }
   }`)
}
