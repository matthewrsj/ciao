package libsnnet

import (
	"fmt"
	"os/exec"
)

func addPortInternal(bridgeId string, portId string) error {
	// Example: ovs-vsctl add-port ovs-br1 endpoint1
	// Usage: ovs-vsctl add-port <bridge> <port-name>
	args := []string{"add-port", bridgeId, portId, "--", "set", "interface", portId, "type=internal"}

	if err := vsctlCmd(args); err != nil {
		return err
	}

	return nil
}

func addGrePort(bridgeId string, portId string, remoteIp string) error {
	args := []string{"add-port", bridgeId, portId, "--", "set", "interface", portId, "type=gre", fmt.Sprintf("options:remote_ip=%s", remoteIp)}

	if err := vsctlCmd(args); err != nil {
		return err
	}

	return nil
}

func delGrePort(bridgeId string, portId string) error {
	args := []string{"del-port", bridgeId, portId}

	if err := vsctlCmd(args); err != nil {
		return err
	}

	return nil
}

func vsctlCmd(args []string) error {
	if _, err := exec.Command("ovs-vsctl", args...).Output(); err != nil {
		return err
	}

	return nil
}

// TODO: need to call ifconfig to add ip address to port
