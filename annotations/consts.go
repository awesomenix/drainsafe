// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT license.

package annotations

const (
	// DrainSafeMaintenance key for maintenance
	DrainSafeMaintenance string = "drainsafe.azure.com/maintenancestate"
	// DrainSafeMaintenanceType key for maintenancetype
	DrainSafeMaintenanceType string = "drainsafe.azure.com/maintenancetype"
	// Scheduled maintenance is scheduled  on virtual machine
	Scheduled string = "MaintenanceScheduled"
	// MaintenancePending gets maintenance approval from repairman to coordinate repairs
	MaintenancePending string = "MaintenancePending"
	// MaintenanceApproved marked when maintenance is approved by repairman
	MaintenanceApproved string = "MaintenanceApproved"
	// Cordoning workload scheduling will be disabled on virtual machine
	Cordoning string = "NodeCordoning"
	// Cordoned workload scheduling is disabled on virtual machine
	Cordoned string = "NodeCordoned"
	// Draining workload will be drained on virtual machine
	Draining string = "NodeDraining"
	// Drained workload is drained on virtual machine
	Drained string = "NodeDrained"
	// Started maintenance is started on virtual machine
	Started string = "MaintenanceStarted"
	// Running maintenance is completed on virtual machine
	Running string = "NodeRunning"
	// Uncordoned workload scheduling is enabled on virtual machine
	Uncordoned string = "NodeUncordoned"
)
