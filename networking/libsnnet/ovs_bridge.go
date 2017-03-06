package libsnnet

import (
	"os/exec"
)

const (
	ovs-vsctl := "ovs-vsctl"
)

func initBridgeCli(id string) error {
	// Example: ovs-vsctl add-br ovs-br1
	args := []string{"add-br", id}

	// Execute command
	if out, err := exec.Command(ovs-vsctl, args...).Output(); err != nil {
		return err
	}

	return nil
}

func destroyBridgeCli(id string) error {
	// Example: ovs-vsctl del-br ovs-br1
	args := []string("del-br", id}

	if out, err := exec.Command(ovs-vsctl, args...).Output(); err != nil {
		return err
	}

	return nil
}
