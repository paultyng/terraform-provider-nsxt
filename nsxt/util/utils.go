// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package util

import (
	"fmt"
	"hash/fnv"
	"log"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/model"
)

var NsxVersion = ""

func NsxVersionLower(ver string) bool {
	return VersionLower(NsxVersion, ver)
}

func VersionLower(base, ver string) bool {
	requestedVersion, err1 := version.NewVersion(ver)
	currentVersion, err2 := version.NewVersion(base)
	if err1 != nil || err2 != nil {
		log.Printf("[ERROR] Failed perform version check for version %s", ver)
		return true
	}
	return currentVersion.LessThan(requestedVersion)
}

func NsxVersionHigherOrEqual(ver string) bool {
	return VersionHigherOrEqual(NsxVersion, ver)
}

func VersionHigherOrEqual(base, ver string) bool {

	requestedVersion, err1 := version.NewVersion(ver)
	currentVersion, err2 := version.NewVersion(base)
	if err1 != nil || err2 != nil {
		log.Printf("[ERROR] Failed perform version check for version %s", ver)
		return false
	}
	return currentVersion.Compare(requestedVersion) >= 0
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func GetVerifiableID(id, extra string) string {
	h := hash(id + extra)
	return fmt.Sprintf("%s:%x", id, h)
}

func VerifyVerifiableID(id, extra string) bool {
	s := strings.Split(id, ":")
	if len(s) != 2 {
		return false
	}
	h := hash(s[0] + extra)
	return s[1] == fmt.Sprintf("%x", h)
}

func KeyValuePairsReplaceOrAppend(kvpSlice []model.KeyValuePair, k, v string) []model.KeyValuePair {
	for i := range kvpSlice {
		if *kvpSlice[i].Key == k {
			kvpSlice[i].Value = &v
			return kvpSlice
		}
	}
	return append(kvpSlice, model.KeyValuePair{Key: &k, Value: &v})
}
