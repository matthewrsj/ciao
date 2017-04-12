package libsnnet

import (
	"fmt"
	"os/exec"

	"github.com/golang/glog"
)

func addPortInternal(bridgeId string, g *GreTunEP) error {
	// Example: ovs-vsctl add-port ovs-br1 endpoint1
	// Usage: ovs-vsctl add-port <bridge> <port-name>
	var err error
	if g.LinkName == "" {
		if g.LinkName, err = genIface(g, false); err != nil {
			return netError(g, "create geniface %v, %v", g.GlobalID, err)
		}
	}
	args := []string{"add-port", bridgeId, g.LinkName, "--", "set", "interface", g.LinkName, "type=internal"}

	if err := vsctlCmd(args); err != nil {
		return err
	}

	return nil
}

func ifconfigInterface(portID string, localIP string) error {
	args := []string{portID, localIP}
	glog.Warning(localIP)
	out, err := exec.Command("ifconfig", args...).Output()
	if err != nil {
		glog.Warning(out)
		return err
	}

	return nil
}

func createGrePort(bridgeId string, portId string, remoteIp string) error {
	args := []string{"add-port", bridgeId, portId, "--", "set", "interface", portId, "type=gre", fmt.Sprintf("options:remote_ip=%s", remoteIp)}
	glog.Warning(args)

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
