package libsnnet

import (
	"os/exec"
	"errors"
	"syscall"
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
	cmd := exec.Command("ovs-vsctl", args...)

	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0
			if _, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return errors.New("ovs-vsctl command exited with non-zero exit code")
			}
		} else {
			return err
		}
	}

	return nil
}
