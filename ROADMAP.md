## Swan Roadmap

## About this document

Here we provide high-level description of desired features which are gathered from the community and planned by the Swan team. This should serve as a reference point for Swan users and contributors to understand the direction that the project is heading to, and help determined if a contribution could be conflicting with the long term plan or not.

## How to help?

Discussion on the roadmap can take place in threads under [Issues](https://github.com/Dataman-Cloud/swan/issues). Please open or comment on an issue if you want to provide suggestions and feedback to an item in the roadmap. You'd better review the roadmap before file a issue to avoid potential duplicated effort.

## How to add an item to the roadmap?

Please open an issue to track any initiative on the roadmap of Swan. We will consider the action on the issue comments.

### Application CRUD
Implement application CRUD and lifecycle management and failover.

### Rolling Update
Supports specifying the number of updates and auto rollbak when updating failed.

### Health Check
Supports two kinds of health check methods: HTTP and TCP.

### Event Stream Record
Can retrive and query application event and event history.

### Service Discovery
Supports layer 7 and layer 4 discovery.

### Unique Ip for each container
Swan user could assign a unique IP to any container.

### Mesos Health Check
Implement Mesos based health check.

### Resources contraints
Give any app the ability to select desired resources to run.

### High Availability
Use Raft to implement HA, any node lose doesn't affect overall stablility.

### Load Balance
Load balance HTTP requests to some task serving the requests.

### Priority-Based Preemptive
Priority based task preemption.

### Resource Quota

### CLI Client

### DashBoard
