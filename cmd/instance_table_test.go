package cmd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/cmd"
	boshdir "github.com/cloudfoundry/bosh-init/director"
)

var _ = Describe("InstanceTable", func() {
	Describe("ForVMInfo", func() {
		var (
			info boshdir.VMInfo
			tbl  InstanceTable
		)

		BeforeEach(func() {
			info = boshdir.VMInfo{}
			tbl = InstanceTable{Details: true, DNS: true, Vitals: true}
		})

		Describe("name, id, index, bootstrap", func() {
			It("returns ? name", func() {
				Expect(tbl.ForVMInfo(info).Name.String()).To(Equal("?/?"))
			})

			It("returns ? name with bootstrap", func() {
				info.Bootstrap = true
				Expect(tbl.ForVMInfo(info).Name.String()).To(Equal("?/?*"))
			})

			It("returns ? name", func() {
				info.JobName = "name"
				Expect(tbl.ForVMInfo(info).Name.String()).To(Equal("name/?"))
			})

			It("returns name with index", func() {
				idx := 1
				info.JobName = "name"
				info.Index = &idx
				Expect(tbl.ForVMInfo(info).Name.String()).To(Equal("name/1"))
			})

			It("returns name with index and bootstrap", func() {
				idx := 1
				info.JobName = "name"
				info.Index = &idx
				info.Bootstrap = true
				Expect(tbl.ForVMInfo(info).Name.String()).To(Equal("name/1*"))
			})

			It("returns name with id and index", func() {
				idx := 1
				info.JobName = "name"
				info.ID = "id"
				info.Index = &idx
				Expect(tbl.ForVMInfo(info).Name.String()).To(Equal("name/id (1)"))
			})

			It("returns name with id, bootstrap and index", func() {
				idx := 1
				info.JobName = "name"
				info.ID = "id"
				info.Index = &idx
				info.Bootstrap = true
				Expect(tbl.ForVMInfo(info).Name.String()).To(Equal("name/id* (1)"))
			})

			It("returns name with id, without index", func() {
				info.JobName = "name"
				info.ID = "id"
				Expect(tbl.ForVMInfo(info).Name.String()).To(Equal("name/id"))
			})

			It("returns ? name with id", func() {
				info.JobName = ""
				info.ID = "id"
				Expect(tbl.ForVMInfo(info).Name.String()).To(Equal("?/id"))
			})
		})

		Describe("vm type, resource pool", func() {
			It("returns RP if vm type is empty", func() {
				info.ResourcePool = "rp"
				Expect(tbl.ForVMInfo(info).VMType.String()).To(Equal("rp"))
			})

			It("returns vm type if vm type is non-empty", func() {
				info.ResourcePool = "rp"
				info.VMType = "vm-type"
				Expect(tbl.ForVMInfo(info).VMType.String()).To(Equal("vm-type"))
			})
		})

		Describe("disk cid", func() {
			It("returns empty if disk cid is empty", func() {
				Expect(tbl.ForVMInfo(info).DiskCID.String()).To(Equal(""))
			})

			It("returns disk cid if disk cid is non-empty", func() {
				info.DiskID = "disk-cid"
				Expect(tbl.ForVMInfo(info).DiskCID.String()).To(Equal("disk-cid"))
			})
		})
	})
})
