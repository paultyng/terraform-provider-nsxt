// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/vmware/terraform-provider-nsxt/nsxt/util"
)

type resourceFunc func() *schema.Resource

func commonVersionCheck(m interface{}) bool {
	initNSXVersion(getPolicyConnector(m))
	return !util.NsxVersionHigherOrEqual("9.0.0")
}

func readWrapper(originalFunc schema.ReadFunc, name string, fail bool) schema.ReadFunc {
	wrappedFunc := func(d *schema.ResourceData, m interface{}) error {
		if commonVersionCheck(m) {
			return originalFunc(d, m)

		}
		if fail {
			// We fail for data source only
			log.Printf("[INFO] Failing read for data source %s: removed from NSX 9.0.0", name)
			return mpDataSourceRemovedError(name)
		}
		log.Printf("[INFO] Skipping read for resource %s: removed from NSX 9.0.0", name)
		return nil
	}
	return wrappedFunc
}

func createWrapper(originalFunc schema.CreateFunc, name string) schema.CreateFunc {
	wrappedFunc := func(d *schema.ResourceData, m interface{}) error {
		if commonVersionCheck(m) {
			return originalFunc(d, m)
		}
		log.Printf("[INFO] Failing write for resource %s: removed from NSX 9.0.0", name)
		return mpResourceRemovedError(name)
	}
	return wrappedFunc
}

func updateWrapper(originalFunc schema.UpdateFunc, name string) schema.UpdateFunc {
	wrappedFunc := func(d *schema.ResourceData, m interface{}) error {
		if commonVersionCheck(m) {
			return originalFunc(d, m)
		}
		log.Printf("[INFO] Failing update for resource %s: removed from NSX 9.0.0", name)
		return mpResourceRemovedError(name)
	}
	return wrappedFunc
}

func deleteWrapper(originalFunc schema.DeleteFunc, name string) schema.DeleteFunc {
	wrappedFunc := func(d *schema.ResourceData, m interface{}) error {
		if commonVersionCheck(m) {
			return originalFunc(d, m)
		}
		log.Printf("[INFO] Skipping delete for resource %s: removed from NSX 9.0.0", name)
		return nil
	}
	return wrappedFunc
}

func importerWrapper(originalImporter *schema.ResourceImporter, name string) *schema.ResourceImporter {
	newImporter := new(schema.ResourceImporter)
	wrappedFunc := func(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
		if commonVersionCheck(m) {
			stateFunc := originalImporter.State
			return stateFunc(d, m)
		}
		log.Printf("[INFO] Failing import for resource %s: removed from NSX 9.0.0", name)
		return nil, mpResourceRemovedError(name)
	}
	newImporter.State = wrappedFunc
	newImporter.StateContext = originalImporter.StateContext
	return newImporter
}

func removedResourceWrapper(realResourceFunc resourceFunc, name string) *schema.Resource {
	resource := realResourceFunc()
	resource.Read = readWrapper(resource.Read, name, false)
	if resource.Delete != nil {
		resource.Delete = deleteWrapper(resource.Delete, name)
	}
	if resource.Create != nil {
		resource.Create = createWrapper(resource.Create, name)
	}
	if resource.Update != nil {
		resource.Update = updateWrapper(resource.Update, name)
	}
	if resource.Importer != nil {
		resource.Importer = importerWrapper(resource.Importer, name)
	}
	return resource
}

func removedDataSourceWrapper(realDataSourceFunc resourceFunc, name string) *schema.Resource {
	resource := realDataSourceFunc()
	resource.Read = readWrapper(resource.Read, name, true)
	return resource
}
