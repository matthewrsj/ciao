package libsnnet

import (

)

func addPortCli(id string, portId string) error {
	// Example: ovs-vsctl add-port ovs-br1 endpoint1
	// Usage: ovs-vsctl add-port <bridge> <port-name>
	args := []string{"add-port", id, portId}

	if out, err := exec.Command(ovs-vsctl, args...).Output(); err != nil {
		return err
	}

	return nil
}

// TODO: need to call ifconfig to add ip address to port
