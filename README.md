


- [DrainSafe](#DrainSafe)
- [Design](#Design)
  - [Events](#Events)
  - [Scheduled Events Controller](#Scheduled-Events-Controller)
  - [Safe drain Controller](#Safe-drain-Controller)

## DrainSafe

Azure has a [scheduled events feature](https://docs.microsoft.com/en-us/azure/virtual-machines/linux/scheduled-events) which lets safely drain the workload based on planned/unplanned maintenance events. 

## Design

### Events

Following events are defined based on which controllers perform certain actions
- **Scheduled** - Maintenance is scheduled  on virtual machine
- **Cordoned** - Scheduling is disabled on virtual machine
- **Drained** - Workload is drained on virtual machine
- **Started** - Maintenance is started on virtual machine
- **Running** - Maintenance is completed on virtual machine
- **Uncordoned** - Scheduling is enabled on virtual machine

### Scheduled Events Controller

- Runs as a daemonset which watches [scheduled events](https://docs.microsoft.com/en-us/azure/virtual-machines/linux/scheduled-events) for virtual machine its running on.
- Emits a **Scheduled** Event when a maintenance is scheduled.
- Emits a **Started** Event when a maintenance is started.
- Emits a **Running** Event at daemonset startup.

### Safe drain Controller

- Runs as a controller watches pre defined [events](#Events).
- Emits a **Cordoned** Event when node has been corded based on **Scheduled** event.
- Emits a **Drained** Event when a node has been drained based on **Cordoned** event.
- Emits a **Uncordoned** Event when node has been uncordened based on **Running** event.