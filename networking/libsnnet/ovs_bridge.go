package libsnnet

import (
	"os/exec"
	"errors"
)

func createOvsBridge(bridgeId string) error {
	// Example: ovs-vsctl add-br ovs-br1
	args := []string{"add-br", bridgeId}

	// Execute command
	if err := vsctlCmd(args); err != nil {
		return err
	}

	return nil
}

func destroyBridgeCli(bridgeId string) error {
	// Example: ovs-vsctl del-br ovs-br1
	args := []string{"del-br", bridgeId}

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
/*
func getOvsDevice(bridgeId string) error {
	args := []string{"br-exists", bridgeId}
	if err := vsctlCmd(args); err != nil {
		return err
	}

	if out,_ := exec.Command("echo", "$?").Output(); string(out) == "0" {
		return errors.New("bridge already exists")
	}

	return nil
}
*/
