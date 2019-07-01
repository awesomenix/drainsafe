package events

const (
	// Scheduled maintenance is scheduled  on virtual machine
	Scheduled string = "MaintenanceScheduled"
	// Cordoned workload scheduling is disabled on virtual machine
	Cordoned string = "NodeCordoned"
	// Drained workload is drained on virtual machine
	Drained string = "NodeDrained"
	// Started maintenance is started on virtual machine
	Started string = "MaintenanceStarted"
	// Running maintenance is completed on virtual machine
	Running string = "NodeRunning"
	// Uncordoned workload scheduling is enabled on virtual machine
	Uncordoned string = "NodeUncordoned"
)
