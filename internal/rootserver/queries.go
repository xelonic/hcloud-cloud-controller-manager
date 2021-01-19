package rootserver

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	cloudprovider "k8s.io/cloud-provider"
)

type Queries interface {
	// IsRootServer finds out if the given node is one looking at the k8s node
	// label instance.hetzner.cloud/is-root-server
	IsRootServer(ctx context.Context, nodeName types.NodeName) (bool, error)
	IsRootServerByNode(node *corev1.Node) bool
	GetInstanceMetadata(node *corev1.Node) (*cloudprovider.InstanceMetadata, error)
	GetRootServerRoutes(ctx context.Context) ([]*cloudprovider.Route, error)
}

type queriesMock struct {
}

func (m *queriesMock) IsRootServer(context.Context, types.NodeName) (bool, error) {
	return false, nil
}

func (m *queriesMock) IsRootServerByNode(*corev1.Node) bool {
	return false
}

func (m *queriesMock) NodeAddresses(context.Context, types.NodeName) ([]corev1.NodeAddress, error) {
	return []corev1.NodeAddress{}, nil
}

func (m *queriesMock) GetInstanceMetadata(*corev1.Node) (*cloudprovider.InstanceMetadata, error) {
	return &cloudprovider.InstanceMetadata{}, nil
}

func (m *queriesMock) GetRootServerRoutes(context.Context) ([]*cloudprovider.Route, error) {
	return []*cloudprovider.Route{}, nil
}

func NewQueriesMock() Queries {
	return &queriesMock{}
}
