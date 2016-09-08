/*
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
*/
// @APIVersion v2.1
// @APITitle CIAO Controller API
// @APIDescription Ciao controller is responsible for policy choices around tenant workloads. It provides compute API endpoints for access from ciao-cli and ciao-webui over HTTPS.
// @Contact https://github.com/01org/ciao/wiki/Package-maintainers
// @License Apache License, Version 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0
// @BasePath http://<ciao-controller-server>:8774/v2.1/

package main

import (
	"flag"
	"os"
	"strconv"
	"sync"

	datastore "github.com/01org/ciao/ciao-controller/internal/datastore"
	image "github.com/01org/ciao/ciao-image/client"
	storage "github.com/01org/ciao/ciao-storage"
	"github.com/01org/ciao/osprepare"
	"github.com/01org/ciao/ssntp"
	"github.com/01org/ciao/testutil"
	"github.com/golang/glog"
)

type controller struct {
	storage.BlockDriver

	client *ssntpClient
	ds     *datastore.Datastore
	id     *identity
	image  image.Client
}

var singleMachine = flag.Bool("single", false, "Enable single machine test")
var cert = flag.String("cert", "", "Client certificate")
var caCert = flag.String("cacert", "", "CA certificate")
var serverURL = flag.String("url", "", "Server URL")
var identityURL = "identity:35357"
var serviceUser = "csr"
var servicePassword = ""
var computeAPIPort = 8774
var httpsCAcert = "/etc/pki/ciao/ciao-controller-cacert.pem"
var httpsKey = "/etc/pki/ciao/ciao-controller-key.pem"
var tablesInitPath = flag.String("tables_init_path", "./tables", "path to csv files")
var workloadsPath = flag.String("workloads_path", "./workloads", "path to yaml files")
var noNetwork = flag.Bool("nonetwork", false, "Debug with no networking")
var persistentDatastoreLocation = flag.String("database_path", "./ciao-controller.db", "path to persistent database")
var transientDatastoreLocation = flag.String("stats_path", "/tmp/ciao-controller-stats.db", "path to stats database")
var logDir = "/var/lib/ciao/logs/controller"

var imagesPath = flag.String("images_path", "/var/lib/ciao/images", "path to ciao images")

var keyringPath = flag.String("ceph_keyring", "", "path to ceph client keyring")
var cephID = flag.String("ceph_id", "", "ceph client id")
var cephConfigPath = flag.String("ceph_config", "", "path to ceph config file")

func init() {
	flag.Parse()

	logDirFlag := flag.Lookup("log_dir")
	if logDirFlag == nil {
		glog.Errorf("log_dir does not exist")
		return
	}

	if logDirFlag.Value.String() == "" {
		logDirFlag.Value.Set(logDir)
	}

	if err := os.MkdirAll(logDirFlag.Value.String(), 0755); err != nil {
		glog.Errorf("Unable to create log directory (%s) %v", logDir, err)
		return
	}
}

func main() {
	var wg sync.WaitGroup
	var err error

	context := new(controller)
	context.ds = new(datastore.Datastore)

	context.image = image.Client{MountPoint: *imagesPath}

	dsConfig := datastore.Config{
		PersistentURI:     *persistentDatastoreLocation,
		TransientURI:      *transientDatastoreLocation,
		InitTablesPath:    *tablesInitPath,
		InitWorkloadsPath: *workloadsPath,
	}

	err = context.ds.Init(dsConfig)
	if err != nil {
		glog.Fatalf("unable to Init datastore: %s", err)
		return
	}

	config := &ssntp.Config{
		URI:    *serverURL,
		CAcert: *caCert,
		Cert:   *cert,
		Log:    ssntp.Log,
	}

	context.client, err = newSSNTPClient(context, config)
	if err != nil {
		// spawn some retry routine?
		glog.Fatalf("unable to connect to SSNTP server")
		return
	}

	clusterConfig, err := context.client.ssntp.ClusterConfiguration()
	if err != nil {
		glog.Fatalf("Unable to retrieve Cluster Configuration: %v", err)
		return
	}

	computeAPIPort = clusterConfig.Configure.Controller.ComputePort
	httpsCAcert = clusterConfig.Configure.Controller.HTTPSCACert
	httpsKey = clusterConfig.Configure.Controller.HTTPSKey
	identityURL = clusterConfig.Configure.IdentityService.URL
	serviceUser = clusterConfig.Configure.Controller.IdentityUser
	servicePassword = clusterConfig.Configure.Controller.IdentityPassword
	if *keyringPath == "" {
		*keyringPath = clusterConfig.Configure.Storage.SecretPath
	}
	if *cephID == "" {
		*cephID = clusterConfig.Configure.Storage.CephID
	}
	if *cephConfigPath == "" {
		*cephConfigPath = clusterConfig.Configure.Storage.ConfigPath
	}

	osprepare.InstallDeps(controllerDeps)

	if *singleMachine {
		hostname, _ := os.Hostname()
		computeURL := "https://" + hostname + ":" + strconv.Itoa(computeAPIPort)
		testIdentityConfig := testutil.IdentityConfig{
			ComputeURL: computeURL,
			ProjectID:  "f452bbc7-5076-44d5-922c-3b9d2ce1503f",
		}

		id := testutil.StartIdentityServer(testIdentityConfig)
		defer id.Close()
		identityURL = id.URL
		glog.Errorf("========================")
		glog.Errorf("Identity URL: %s", id.URL)
		glog.Errorf("Please")
		glog.Errorf("export CIAO_IDENTITY=%s", id.URL)
		glog.Errorf("========================")
		glog.Flush()
	}

	idConfig := identityConfig{
		endpoint:        identityURL,
		serviceUserName: serviceUser,
		servicePassword: servicePassword,
	}

	context.BlockDriver = func() storage.BlockDriver {
		driver := storage.CephDriver{
			SecretPath: *keyringPath,
			ID:         *cephID,
			ConfigPath: *cephConfigPath,
		}
		return driver
	}()

	context.id, err = newIdentityClient(idConfig)
	if err != nil {
		glog.Fatal("Unable to authenticate to Keystone: ", err)
		return
	}

	wg.Add(1)
	go createComputeAPI(context)

	wg.Add(1)
	go context.startVolumeService()

	wg.Wait()
	context.ds.Exit()
	context.client.Disconnect()
}
