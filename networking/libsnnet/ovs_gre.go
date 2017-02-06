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

func (g *GreTunEP) createBridge(ovs *libovsdb.OvsdbClient) {
	namedUUID := "ciaoee-wowee"
	// bridge row to insert
	bridge := make(map[string]interface{})
	bridge["name"] = g.GlobalId

	// simple insert operation
	insertOp := libovsdb.Operation{
		Op:       "insert",
		Table:    "Bridge",
		Row:      bridge,
		UUIDName: namedUUID,
	}

	// Inserting a Bridge row in Bridge table requires mutating the open_vswitch table.
	uuidParameter := libovsdb.UUID{GoUUID: getRootUUID()}
	mutateUUID := []libovsdb.UUID{{namedUUID}}
	mutateSet, _ := libovsdb.NewOvsSet(mutateUUID)
	mutation := libovsdb.NewMutation("bridges", "insert", mutateSet)
	condition := libovsdb.NewCondition("_uuid", "==", uuidParameter)

	// simple mutate operation
	mutateOp := libovsdb.Operation{
		Op:        "mutate",
		Table:     "Open_vSwitch",
		Mutations: []interface{}{mutation},
		Where:     []interface{}{condition},
	}

	operations := []libovsdb.Operation{insertOp, mutateOp}
	reply, _ := ovs.Transact("Open_vSwitch", operations...)

	if len(reply) < len(operations) {
		fmt.Println("Number of Replies should be atleast equal to number of Operations")
	}
	ok := true
	for i, o := range reply {
		if o.Error != "" && i < len(operations) {
			fmt.Println("Transaction Failed due to an error :", o.Error, " details:", o.Details, " in ", operations[i])
			ok = false
		} else if o.Error != "" {
			fmt.Println("Transaction Failed due to an error :", o.Error)
			ok = false
		}
	}
	if ok {
		fmt.Println("Bridge Addition Successful : ", reply[0].UUID.GoUUID)
	}
}
