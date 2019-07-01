package annotations

const (
	// DrainSafeMaintenance key for maintenance
	DrainSafeMaintenance string = "drainsafe.azure.com/maintenancestate"
	// Scheduled maintenance is scheduled  on virtual machine
	Scheduled string = "Scheduled"
	// Cordoning workload scheduling will be disabled on virtual machine
	Cordoning string = "Cordoning"
	// Cordoned workload scheduling is disabled on virtual machine
	Cordoned string = "Cordoned"
	// Draining workload will be drained on virtual machine
	Draining string = "Draining"
	// Drained workload is drained on virtual machine
	Drained string = "Drained"
	// Started maintenance is started on virtual machine
	Started string = "Started"
	// Running maintenance is completed on virtual machine
	Running string = "Running"
	// Uncordoned workload scheduling is enabled on virtual machine
	Uncordoned string = "Uncordoned"
)
