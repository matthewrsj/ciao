//
// Copyright (c) 2016 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package libsnnet

import (
	"github.com/socketplane/libovsdb"
	"fmt"
	"os"
)

// Enable the GreTunnel
func (g *GreTunEP) connect() (*libovsdb.OvsdbClient, error) {

	ovs, err := libovsdb.ConnectUsingProtocol("gre", g.RemoteIP.String())

	if err != nil {
		fmt.Println("Unable to Connect ", err)
		os.Exit(1)
	}

	return ovs, nil
}

func (g *GreTunEP) addGrePort(bridgeName string) error {
	ovs = libovsdb.OvsdbClient()
	PortNamedUUID = "ciao-port"
	IntfNamedUUID = "ciao-intf"

	options := make(map[string]interface{})
	options["remote_ip"] = g.RemoteIP
	//options["ip"] = g.LocalIP // is this necessary?
	intf["name"] = g.GlobalID
	intf["type"] = `gre`
	intf["options"], _ = libovsdb.NewOvsMap(options)

	insertIntfOp := libovsdb.Operation{
		Op:        "insert",
		Table:     "Interface",
		Row:       intf,
		UUIDName:  IntfNamedUUID,
	}

	port := make(map[string]interface{})
	port["name"] = g.GlobalID
	port["interfaces"] = libovsdb.UUID{IntfNamedUUID}

	insertPortOp := libovsdb.Operation{
		Op:        "insert",
		Table:     "Port",
		Row:       port,
		UUIDName:  PortNamedUUID,
	}

	operations := []libovsdb.Operation{insertIntfOp, insertPortOp}
	reply, _ := ovs.ovsdb.Transact("Open_vSwitch", operations...)
	if len(reply) < len(operations) {
		return netError("Number of replies should at least equal number of operations")
	}

	for i, o := range reply {
		if o.Error != "" && i < len(operations) {
			msg := fmt.Sprintf("Transaction failed due to an error : %v details: %v in %v", o.Error, o.Details, operations[i])
			return netError(g, msg)
		} else if o.Error != "" {
			msg := fmt.Sprintf("Transaction failed due to an error : %v", o.Error)
			return netError(g, msg)
		}
	}

	return nil
}
