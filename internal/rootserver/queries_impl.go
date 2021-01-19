package rootserver

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	cloudprovider "k8s.io/cloud-provider"
	"k8s.io/klog/v2"
)

const (
	instanceIsRootServer = "instance.hetzner.cloud/is-root-server"
)

type queries struct {
	k8sClientSet *kubernetes.Clientset
}

func NewQueries() Queries {
	var k8sClientSet *kubernetes.Clientset

	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		klog.ErrorS(err, "k8s cluster config can't be created")
	} else {
		k8sClientSet, err = kubernetes.NewForConfig(k8sConfig)
		if err != nil {
			klog.ErrorS(err, "k8s clients can't be initialized")
		}
	}

	return &queries{k8sClientSet}
}

func (q *queries) IsRootServer(ctx context.Context, nodeName types.NodeName) (bool, error) {
	if err := q.validateK8sConnection(); err != nil {
		return false, err
	}

	node, err := q.k8sClientSet.CoreV1().Nodes().Get(ctx, string(nodeName), metav1.GetOptions{})
	if err != nil {
		return false, errors.WithMessagef(err, "failed to retrieve k8s node info for node '%v'", nodeName)
	}

	isRootServer, err := hasRootServerLabel(node)
	if err != nil {
		return false, err
	}

	return isRootServer, nil
}

func (q *queries) IsRootServerByNode(node *corev1.Node) bool {
	if err := q.validateK8sConnection(); err != nil {
		klog.Error(err)
		return false
	}

	isRootServer, err := hasRootServerLabel(node)
	if err != nil {
		klog.ErrorS(err, "failed to query root server label; assuming it's a cloud node", "node", node.Name)
		return false
	}

	return isRootServer
}

func (q *queries) GetInstanceMetadata(node *corev1.Node) (*cloudprovider.InstanceMetadata, error) {
	if err := q.validateK8sConnection(); err != nil {
		return nil, err
	}

	return &cloudprovider.InstanceMetadata{
		ProviderID:    "",
		InstanceType:  "",
		NodeAddresses: node.Status.Addresses,
		Zone:          "",
		Region:        "",
	}, nil
}

func (q *queries) GetRootServerRoutes(ctx context.Context) ([]*cloudprovider.Route, error) {
	rootNodes, err := q.getRootServerNodes(ctx)
	if err != nil {
		return nil, err
	}

	rootServerRoutes := make([]*cloudprovider.Route, 0, len(rootNodes))

	for _, node := range rootNodes {
		if len(node.Spec.PodCIDR) < 1 {
			continue
		}

		destination := node.Spec.PodCIDR

		var gateway string

		for _, nodeAddress := range node.Status.Addresses {
			if nodeAddress.Type == corev1.NodeInternalIP && len(nodeAddress.Address) > 0 {
				gateway = nodeAddress.Address
				break
			}
		}

		if len(gateway) < 1 {
			continue
		}

		route := &cloudprovider.Route{
			DestinationCIDR: destination,
			Name:            fmt.Sprintf("%s-%s", gateway, destination),
			TargetNode:      types.NodeName(node.Name),
		}

		rootServerRoutes = append(rootServerRoutes, route)
	}

	return rootServerRoutes, nil
}

func hasRootServerLabel(node *corev1.Node) (bool, error) {
	value, ok := node.Labels[instanceIsRootServer]
	if !ok {
		return false, nil
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf(
			"node %s has invalid label '%s': %v",
			node.Name,
			instanceIsRootServer,
			err,
		)
	}

	return boolValue, nil
}

func (q *queries) getRootServerNodes(ctx context.Context) ([]*corev1.Node, error) {
	if err := q.validateK8sConnection(); err != nil {
		return []*corev1.Node{}, err
	}

	nodeList, err := q.k8sClientSet.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return []*corev1.Node{}, errors.WithMessage(err, "failed to list k8s nodes")
	}

	rootServerNodes := make([]*corev1.Node, 0, len(nodeList.Items))

	for _, node := range nodeList.Items {
		pNode := node
		isRootServer, err := hasRootServerLabel(&pNode)
		if err != nil || !isRootServer {
			continue
		}

		rootServerNodes = append(rootServerNodes, &pNode)
	}

	return rootServerNodes, nil
}

func (q *queries) validateK8sConnection() error {
	if q.k8sClientSet == nil {
		return errors.New("no connection to kubernetes API")
	}

	return nil
}
