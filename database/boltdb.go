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

package database

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/boltdb/bolt"
	"github.com/golang/glog"
)

type boltDB struct {
	Name  string
	Cache map[string]map[string]interface{}
	DB    *bolt.DB
}

type dbProvider boltDB

func newBoltDb() *boltDB {
	return &boltDB{
		Name: "bolt.DB",
	}
}

//NewBoltDBProvider returns a bolt based database that conforms
//to the DBProvider interface
func NewBoltDBProvider() DbProvider {
	return (*dbProvider)(newBoltDb())
}

func (db *dbProvider) DbInit(dbDir, dbFile string) error {

	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("Unable to create db directory (%s) %v", dbDir, err)
	}

	dbPath := path.Join(dbDir, dbFile)

	options := bolt.Options{
		Timeout: 3 * time.Second,
	}

	var err error
	db.DB, err = bolt.Open(dbPath, 0644, &options)
	if err != nil {
		return fmt.Errorf("initDb failed %v", err)
	}

	return err
}

func (db *dbProvider) DbClose() error {
	return db.DB.Close()
}

func (db *dbProvider) DbTablesInit(tables []string) (err error) {

	glog.Infof("dbInit Tables := %v", tables)
	db.Cache = make(map[string]map[string]interface{})
	for i, table := range tables {
		glog.Infof("table[%v] := %v, %v", i, table, []byte(table))
		db.Cache[table] = make(map[string]interface{})
	}

	err = db.DB.Update(func(tx *bolt.Tx) error {
		for _, table := range tables {
			_, err := tx.CreateBucketIfNotExists([]byte(table))
			if err != nil {
				return fmt.Errorf("Bucket creation error: %v %v", table, err)
			}
		}
		return nil
	})

	if err != nil {
		glog.Errorf("Table creation error %v", err)
	}

	return err
}

func (db *dbProvider) DbAdd(table string, key string, value interface{}) (err error) {

	db.Cache[table][key] = value
	err = db.DB.Update(func(tx *bolt.Tx) error {

		var v bytes.Buffer

		if err := gob.NewEncoder(&v).Encode(value); err != nil {
			glog.Errorf("Encode Error: %v %v", err, value)
			return err
		}

		bucket := tx.Bucket([]byte(table))
		if bucket == nil {
			return fmt.Errorf("Bucket %v not found", table)
		}

		err = bucket.Put([]byte(key), v.Bytes())
		if err != nil {
			return fmt.Errorf("Key Store error: %v %v %v %v", table, key, value, err)
		}
		return nil
	})

	return err
}

func (db *dbProvider) DbDelete(table string, key string) (err error) {

	delete(db.Cache[table], key)
	err = db.DB.Update(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte(table))
		if bucket == nil {
			return fmt.Errorf("Bucket %v not found", table)
		}

		err = bucket.Delete([]byte(key))
		if err != nil {
			return fmt.Errorf("Key Delete error: %v %v ", key, err)
		}
		return nil
	})

	return err
}

func (db *dbProvider) DbGet(table string, key string, dbTable DbTable) (interface{}, error) {

	elem := db.Cache[table][key]
	if elem != nil {
		return elem, nil
	}

	err := db.DB.View(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte(table))
		if bucket == nil {
			return fmt.Errorf("Bucket %v not found", table)
		}
		data := bucket.Get([]byte(key))
		vr := bytes.NewReader(data)

		elem = dbTable.NewElement()
		if err := gob.NewDecoder(vr).Decode(elem); err != nil {
			return err
		}
		return nil
	})

	db.Cache[table][key] = elem

	return elem, err
}

func (db *dbProvider) DbGetAll(table string, dbTable DbTable) (elements []interface{}, err error) {
	err = db.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(table))

		err := b.ForEach(func(key, value []byte) error {
			vr := bytes.NewReader(value)
			elem := dbTable.NewElement()
			if err := gob.NewDecoder(vr).Decode(elem); err != nil {
				return err
			}
			db.Cache[table][bytes.NewBuffer(key).String()] = elem
			elements = append(elements, elem)
			return nil
		})
		return err
	})

	return elements, err
}
