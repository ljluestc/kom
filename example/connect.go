package example

import (
	"github.com/weibaohui/kom/callbacks"
	"github.com/weibaohui/kom/kom"
)

func Connect() {
	callbacks.RegisterInit()

	// For testing, use a test kubeconfig instead of real kubeconfig
	testKubeconfig := `apiVersion: v1
clusters:
- cluster:
    server: https://test-cluster.example.com:6443
    insecure-skip-tls-verify: true
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: admin
  name: test-cluster
current-context: test-cluster
kind: Config
users:
- name: admin
  user:
    token: test-token-12345`

	// Register test cluster as default
	_, _ = kom.Clusters().RegisterByStringWithID(testKubeconfig, "default")
	kom.Clusters().Show()
}
