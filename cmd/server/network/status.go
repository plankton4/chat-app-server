package network

var ConnectedClients = make(map[uint32]*Client)
var ActiveHubs = make(map[HubID]*Hub)
