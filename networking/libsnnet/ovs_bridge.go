package libsnnet

import (
	"os/exec"

	"github.com/golang/glog"
)
func createOvsBridge(bridgeId string) error {
	// Example: ovs-vsctl add-br ovs-br1
	args := []string{"add-br", bridgeId, "--", "set", "bridge", bridgeId, "datapath_type=netdev"}

	// Execute command
	if err := vsctlCmd(args); err != nil {
		return err
	}

	return nil
}

func destroyOvsBridge(bridgeId string) error {
	// Example: ovs-vsctl del-br ovs-br1
	args := []string{"del-br", bridgeId}

	if err := vsctlCmd(args); err != nil {
		return err
	}

	return nil
}

func addOvsPort(v *Vnic) error {
	args := []string{"add-port", v.BridgeID, v.Attrs.LinkName}

	if err := vsctlCmd(args); err != nil {
		return err
	}

	return nil
}

func delOvsPort(v *Vnic) error {
	args := []string{"del-port", v.BridgeID, v.Attrs.LinkName}

	if err := vsctlCmd(args); err != nil {
		return err
	}

	return nil
}

func vsctlCmd(args []string) error {
	_, err := exec.Command("ovs-vsctl", args...).Output()

	if err != nil {
		glog.Error("vsctlCmd failed: " + err.Error())
		return err
	}

	return nil
}
