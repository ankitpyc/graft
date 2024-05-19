package raft

import "time"

type Discovery struct {
}

func (discovery *Discovery) HealthCheck(peer *Client) {
	if peer.NodeDetails.NodeStatus == "Leader" {
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

func healthCheckEntry(entry *ClusterPeer) {
	panic("implement me")
}
