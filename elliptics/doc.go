/*
* 2013+ Copyright (c) Anton Tyurin <noxiouz@yandex.ru>
* All rights reserved.
*
* This program is free software; you can redistribute it and/or modify
* it under the terms of the GNU General Public License as published by
* the Free Software Foundation; either version 2 of the License, or
* (at your option) any later version.
*
* This program is distributed in the hope that it will be useful,
* but WITHOUT ANY WARRANTY; without even the implied warranty of
* MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
* GNU General Public License for more details. */

/*
Package provides interface to work with Elliptics. Elliptics is distributed fault-tolerant
key-value storage system, it also supports secondary indexes.

More information about Elliptics here: http://reverbrain.com/.

Motivaiting example:

	// Create filelogger
	EllLog, err := elliptics.NewFileLogger("/tmp/elliptics-go.log", elliptics.LOGERROR)
	if err != nil {
		log.Fatalln("NewFileLogger: ", err)
	}
	defer EllLog.Free()

	EllLog.Log(elliptics.LOGINFO, "started: %v, level: %d", time.Now(), level)

	// Create elliptics node
	node, err := elliptics.NewNode(EllLog)
	if err != nil {
		log.Println(err)
	}
	defer node.Free()

	if err = node.AddRemote("example.some.host:1025:2"); err != nil {
		log.Fatalln("AddRemote: ", err)
	}

	session, err := elliptics.NewSession(node)
	if err != nil {
		log.Fatal("Error", err)
	}

	session.SetGroups([]int32{1, 2, 3})
	session.SetNamespace("example")

	for rd := range session.ReadData("key") {
		log.Printf("%s \n", rd.Data())
	}

	for rw := range session.WriteData("key", "testdata") {
		log.Println(rw)
	}

	lookuped_key, _ := elliptics.NewKey("key")
	defer lookuped_key.Free()

	for lookUp := range session.Lookup(lookuped_key) {
		log.Println(lookUp)
	}

	indexes := map[string]string{
		"F": "indexF",
		"A": "indexA",
	}
	if si, ok := <-session.SetIndexes("key", indexes); !ok {
		log.Println("SetIndexes successfully")
	} else {
		log.Println("SetIndexes error: ", si.Error())
	}

	for li := range session.ListIndexes("key") {
		log.Println("Index: ", li.Data)
	}

	indexes["TTT"] = "IndexTTT"
	if ui, ok := <-session.UpdateIndexes(KEY, indexes); !ok {
		log.Println("UpdateIndexes successfully")
	} else {
		log.Println("UpdateIndexes error: ", ui.Error())
	}

	log.Println("List indexes for key ", "key")
	for li := range session.ListIndexes(KEY) {
		log.Println("Index: ", li.Data)
	}

	//KEY exists
	if rm, ok := <-session.Remove("key"); !ok {
		log.Println("Remove successfully")
	} else {
		log.Println("Removing error: ", rm.Error())
	}

	//KEY doesn't exist
	if rm, ok := <-session.Remove("key"); !ok {
		log.Println("Remove successfully")
	} else {
		log.Println("Removing error: ", rm.Error())
	}
*/
package elliptics
