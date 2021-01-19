package annotation

import (
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"
)

const (
	instanceIsRootServer = "instance.hetzner.cloud/is-root-server"
)

func HasRootServerLabel(node *corev1.Node) (bool, error) {
	value, ok := node.Labels[instanceIsRootServer]
	if !ok {
		return false, nil
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("node %s has invalid label '%s': %v", node.Name, instanceIsRootServer, err)
	}

	return boolValue, nil
}
