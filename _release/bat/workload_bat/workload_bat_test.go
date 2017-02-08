// Copyright (c) 2017 Intel Corporation
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

package workloadbat

import (
	"context"
	"testing"
	"time"

	"github.com/01org/ciao/bat"
)

const standardTimeout = time.Second * 300

const vmCloudInit = `---
#cloud-config
users:
  - name: demouser
    geocos: CIAO Demo User
    lock-passwd: false
    passwd: %s
    sudo: ALL=(ALL) NOPASSWD:ALL
    ssh-authorized-keys:
      - %s
...
`

const vmWorkloadImageName = "Fedora Cloud Base 24-1.2"

func testCreateWorkload(t *testing.T, public bool) {
	// we'll use empty string for now
	tenant := ""

	ctx, cancelFunc := context.WithTimeout(context.Background(), standardTimeout)
	defer cancelFunc()

	// generate ssh test keys?

	// get the Image ID to use.
	// TBD: where does ctx and tenant come from?
	source := bat.Source{
		Type: "image",
	}

	// if we pass in "" for tenant, we get whatever the CIAO_USERNAME value
	// is set to.
	images, err := bat.GetImages(ctx, tenant)
	if err != nil {
		t.Fatal(err)
	}

	for ID, image := range images {
		if image.Name != vmWorkloadImageName {
			continue
		}
		source.ID = ID
	}

	if source.ID == "" {
		t.Fatalf("vm Image %s not available", vmWorkloadImageName)
	}

	// fill out the opt structure for this workload.
	defaults := bat.DefaultResources{
		VCPUs: 2,
		MemMB: 128,
	}

	disk := bat.Disk{
		Bootable:  true,
		Source:    &source,
		Ephemeral: true,
	}

	opt := bat.WorkloadOptions{
		Description: "BAT VM Test",
		VMType:      "qemu",
		FWType:      "legacy",
		Defaults:    defaults,
		Disks:       []bat.Disk{disk},
	}

	var ID string
	if public {
		ID, err = bat.CreatePublicWorkload(ctx, tenant, opt, vmCloudInit)
	} else {
		ID, err = bat.CreateWorkload(ctx, tenant, opt, vmCloudInit)
	}

	if err != nil {
		t.Fatal(err)
	}

	// now get all the workloads
	wls, err := bat.GetAllWorkloads(ctx, "")
	if err != nil {
		t.Fatal(err)
	}

	for _, w := range wls {
		if w.ID == ID {
			if w.Name != opt.Description || w.CPUs != opt.Defaults.VCPUs || w.Mem != opt.Defaults.MemMB {
				t.Fatalf("Workload not defined correctly")
			}
			return
		}
	}

	t.Fatalf("New workload did not appear in workload list")
}

// TestCreateTenantWorkload will create a new private workload, then attempt to
// retrieve the workload information from the tenant's workload list.
func TestCreateTenantWorkload(t *testing.T) {
	testCreateWorkload(t, false)
}

// TestCreatePublicWorkload will create a new public workload, then attempt to
// retrieve the workload information from a normal tenant's workload list.
func TestCreatePublicWorkload(t *testing.T) {
	testCreateWorkload(t, true)
}
