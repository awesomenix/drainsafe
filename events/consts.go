package events

const (
	// Scheduled maintenance is scheduled  on virtual machine
	Scheduled string = "Scheduled"
	// Cordoned workload scheduling is disabled on virtual machine
	Cordoned string = "Cordoned"
	// Drained workload is drained on virtual machine
	Drained string = "Drained"
	// Started maintenance is started on virtual machine
	Started string = "Started"
	// Running maintenance is completed on virtual machine
	Running string = "Running"
	// Uncordoned workload scheduling is enabled on virtual machine
	Uncordoned string = "Uncordoned"
)
