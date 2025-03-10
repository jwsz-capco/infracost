package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getComputeInstanceGroupManagerRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "google_compute_instance_group_manager",
		RFunc:               newComputeInstanceGroupManager,
		Notes:               []string{"Multiple versions are not supported."},
		ReferenceAttributes: []string{"version.0.instance_template"},
	}
}

func newComputeInstanceGroupManager(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var region string
	zone := d.Get("zone").String()
	if zone != "" {
		region = zoneToRegion(zone)
	}

	targetSize := int64(1)
	if d.Get("target_size").Exists() {
		targetSize = d.Get("target_size").Int()
	}

	var machineType string
	purchaseOption := "on_demand"
	disks := []*google.ComputeDisk{}
	guestAccelerators := []*google.ComputeGuestAccelerator{}

	if len(d.References("version.0.instance_template")) > 0 {
		instanceTemplate := d.References("version.0.instance_template")[0]

		machineType = instanceTemplate.Get("machine_type").String()
		purchaseOption = getComputePurchaseOption(instanceTemplate.RawValues)

		if len(instanceTemplate.Get("disk").Array()) > 0 {
			for _, disk := range instanceTemplate.Get("disk").Array() {
				diskSize := int64(100)
				if size := disk.Get("disk_size_gb"); size.Exists() {
					diskSize = size.Int()
				}
				diskType := disk.Get("disk_type").String()

				disks = append(disks, &google.ComputeDisk{
					Type: diskType,
					Size: float64(diskSize),
				})
			}
		}

		guestAccelerators = collectComputeGuestAccelerators(instanceTemplate.Get("guest_accelerator"))
	}

	r := &google.ComputeInstanceGroupManager{
		Address:           d.Address,
		Region:            region,
		MachineType:       machineType,
		PurchaseOption:    purchaseOption,
		TargetSize:        targetSize,
		Disks:             disks,
		GuestAccelerators: guestAccelerators,
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}
