package discovery

import (
	"cache/internal/raft/Client"
	"time"
)

type Discovery struct {
}

func (discovery *Discovery) HealthCheck(peer *raft.RaftClient) {
	if peer.NodeDetails.NodeType == 0 {
		timer := time.NewTicker(10 * time.Second)

		for {
			select {
			case <-timer.C:
				for _, entry := range peer.ClusterMembers {
					healthCheckEntry(entry)
				}

			}
		}
	}
}

func healthCheckEntry(entry *raft.ClusterPeer) {
	panic("implement me")
}
