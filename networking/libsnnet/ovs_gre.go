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

func (ovs *ovsdber) addGrePort(bridgeName string, gre *GreTunEP) error {
	PortNamedUUID = "ciao-port"
	IntfNamedUUID = "ciao-intf"

	options := make(map[string]interface{})
	options["remote_ip"] = gre.RemoteIP
	//options["ip"] = gre.LocalIP // is this necessary?
	intf["name"] = gre.GlobalID
	intf["type"] = `gre`
	intf["options"], _ = libovsdb.NewOvsMap(options)

	insertIntfOp := libovsdb.Operation{
		Op:        "insert",
		Table:     "Interface",
		Row:       intf,
		UUIDName:  IntfNamedUUID,
	}

	port := make(map[string]interface{})
	port["name"] = gre.GlobalID
	port["interfaces"] = libovsdb.UUID{IntfNamedUUID}

	insertPortOp := libovsdb.Operation{
		Op:        "insert",
		Table:     "Port",
		Row:       port,
		UUIDName:  PortNamedUUID,
	}

	// Inserting a row in Port table requires mutating the bridge table.
	mutateUUID := []libovsdb.UUID{libovsdb.UUID{PortNamedUUID}}
	mutateSet, _ := libovsdb.NewOvsSet(mutateUUID)
	mutation := libovsdb.NewMutation("ports", "insert", mutateSet)
	condition := libovsdb.NewCondition("name", "==", bridgeName)

	// simple mutate operation
	mutateOp := libovsdb.Operation{
		Op:        "mutate",
		Table:     "Bridge",
		Mutations: []interface{}{mutation},
		Where:     []interface{}{condition},
	}

	operations := []libovsdb.Operation{insertIntfOp, insertPortOp, mutateOp}
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

func (ovsdber *ovsdber) deletePort(bridgeName string, portName string) error {
	condition := libovsdb.NewCondition("name", "==", portName)
	deleteOp := libovsdb.Operation{
		Op:    "delete",
		Table: "Port",
		Where: []interface{}{condition},
	}

	portUUID := portUUIDForName(portName)
	if portUUID == "" {
		log.Error("Unable to find a matching Port : ", portName)
		return fmt.Errorf("Unable to find a matching Port : [ %s ]", portName)
	}

	// Deleting a Bridge row in Bridge table requires mutating the open_vswitch table.
	mutateUUID := []libovsdb.UUID{libovsdb.UUID{portUUID}}
	mutateSet, _ := libovsdb.NewOvsSet(mutateUUID)
	mutation := libovsdb.NewMutation("ports", "delete", mutateSet)
	condition = libovsdb.NewCondition("name", "==", bridgeName)

	// simple mutate operation
	mutateOp := libovsdb.Operation{
		Op:        "mutate",
		Table:     "Bridge",
		Mutations: []interface{}{mutation},
		Where:     []interface{}{condition},
	}

	operations := []libovsdb.Operation{deleteOp, mutateOp}
	reply, _ := ovsdber.ovsdb.Transact("Open_vSwitch", operations...)

	if len(reply) < len(operations) {
		log.Error("Number of Replies should be atleast equal to number of Operations")
		return fmt.Errorf("Number of Replies should be atleast equal to number of Operations")
	}
	for i, o := range reply {
		if o.Error != "" && i < len(operations) {
			log.Error("Transaction Failed due to an error :", o.Error, " in ", operations[i])
			return fmt.Errorf("Transaction Failed due to an error: %s in %v", o.Error, operations[i])
		} else if o.Error != "" {
			log.Error("Transaction Failed due to an error :", o.Error)
			return fmt.Errorf("Transaction Failed due to an error %s", o.Error)
		}
	}
	return nil
}
