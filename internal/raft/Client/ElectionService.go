package raft

import (
	pb "cache/internal/election"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"math/rand"

	"context"
	"sync"
	"time"
)

type ElectionServiceInf interface {
	startGrpcServer() error
	RunElectionLoop(ctx context.Context)
	GetLeaderId() string
	SendHeartbeats(request pb.HeartbeatRequest)
}

type ElectionService struct {
	pb.UnimplementedLeaderElectionServer
	mu                 sync.Mutex
	currentTerm        int32
	votedFor           string
	leaderID           string
	electionInProgress bool
	client             *RaftClient
	electionTimer      *time.Timer
	electionTimeout    time.Duration
}

func NewElectionService(client *RaftClient) *ElectionService {
	duration := time.Duration(rand.Intn(5)+10) * time.Second
	return &ElectionService{
		mu:                 sync.Mutex{},
		currentTerm:        1,
		votedFor:           "",
		electionTimer:      time.NewTimer(duration),
		electionInProgress: false,
		leaderID:           "",
		client:             client,
		electionTimeout:    duration,
	}
}

func (es *ElectionService) RequestVote(ctx context.Context, req *pb.VoteRequest) (*pb.VoteResponse, error) {
	es.mu.Lock()
	defer es.mu.Unlock()
	fmt.Println("requesting vote at", es.client.NodeDetails.NodePort)
	if req.Term > es.currentTerm {
		es.currentTerm = req.Term
		es.votedFor = ""
	}

	voteGranted := false
	if (es.votedFor == "" || es.votedFor == req.CandidateId) && req.Term >= es.currentTerm {
		es.votedFor = req.CandidateId
		voteGranted = true
	}

	return &pb.VoteResponse{VoteGranted: voteGranted, Term: es.currentTerm}, nil
}

func (es *ElectionService) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	//es.mu.Lock()
	//defer es.mu.Unlock()
	fmt.Println("At node :- ", es.client.NodeDetails.NodePort)
	log.Println("HeartBeat from Leader ")
	if req.Term >= es.currentTerm {
		es.currentTerm = req.Term
		es.leaderID = req.LeaderId
		es.electionInProgress = false
		es.resetElectionTimer(ctx)
		return &pb.HeartbeatResponse{Success: true, Term: es.currentTerm}, nil
	}

	return &pb.HeartbeatResponse{Success: false, Term: es.currentTerm}, nil
}

func (es *ElectionService) resetElectionTimer(Ctx context.Context) {
	if es.electionTimer != nil {
		es.electionTimer.Stop()
	}
	es.electionTimer = time.AfterFunc(es.electionTimeout, func() {
		es.startElection()
	})
}

func (es *ElectionService) startElection() {
	if es.electionInProgress {
		fmt.Println("Election already in progress")
		return
	}
	es.mu.Lock()
	es.currentTerm++
	es.votedFor = "self"
	es.electionInProgress = true
	es.mu.Unlock()
	// Adding logic for RequestVote
	req := &pb.VoteRequest{
		Term:        es.currentTerm,
		CandidateId: "self", // Adjust according to your implementation
	}
	nodesLen := len(es.client.ClusterMembers)
	gatheredVotes := 0
	for _, node := range es.client.ClusterMembers {
		if node.NodePort != es.client.NodeDetails.NodePort {
			voting, err := es.initiatingVoting(node, req)
			if err != nil {
				return
			}
			if voting.VoteGranted == true {
				gatheredVotes = gatheredVotes + 1
			}
		}
	}
	b := gatheredVotes >= (nodesLen / 2)
	if b {
		es.leaderID = es.client.NodeDetails.NodePort
		log.Println("Election success , Congratulations !!")
		go es.SendHeartbeats()
	}
	es.mu.Lock()
	es.electionInProgress = false
	es.mu.Unlock()
}

func (es *ElectionService) initiatingVoting(node *ClusterPeer, vr *pb.VoteRequest) (*pb.VoteResponse, error) {
	dial, err := grpc.NewClient(node.NodeAddr+":"+node.NodePort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer dial.Close()

	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	pb.NewLeaderElectionClient(dial)
	defer cancel()
	return es.RequestVote(ctx, vr)
}

// Example service that handles client requests
type serviceServer struct {
	mu                 sync.Mutex
	leaderID           string
	electionInProgress bool
}

//func (s *serviceServer) HandleRequest(ctx context.Context, req *ClientRequest) (ClientResponse, error) {
//	s.mu.Lock()
//	defer s.mu.Unlock()
//
//	if s.electionInProgress {
//		return &ClientResponse{Message: "Election in progress, please retry"}, nil
//	}
//
//	if s.leaderID != "self" {
//		return &ClientResponse{Message: "Not the leader, please contact the leader"}, nil
//	}
//
//	// Handle the request
//	return &ClientResponse{Message: "Request handled by leader"}, nil
//}

func (es *ElectionService) RunElectionLoop() error {
	for {
		time.Sleep(es.electionTimeout)
		if es.electionInProgress || es.GetLeaderId() == es.client.NodeDetails.NodePort {
			continue
		}
		log.Println("Leader Ping timedOut , Election will happen at :", es.client.NodeDetails.NodePort)
		es.client.RMu.Lock()
		if !es.client.IsLeader() || !es.electionInProgress {
			log.Println("Starting election")
			es.startElection()
		}
		es.client.RMu.Unlock()
	}
}

func (es *ElectionService) GetLeaderId() string {
	return es.leaderID
}

func (es *ElectionService) SendHeartbeats() {
	timer := time.NewTicker(5 * time.Second)
	intialTick := time.NewTimer(200 * time.Nanosecond)
	for {
		select {
		case <-timer.C:
			if es.client.IsLeader() {
				sendHeartBeats(es)
			} else {
				break
			}
		case <-intialTick.C:
			if es.client.IsLeader() {
				sendHeartBeats(es)
			} else {
				break
			}
		}
	}
}

func sendHeartBeats(es *ElectionService) {
	es.mu.Lock()
	defer es.mu.Unlock()
	for _, node := range es.client.ClusterMembers {
		if node.NodePort == es.client.NodeDetails.NodePort {
			continue
		}
		url := node.NodeAddr + ":" + node.NodePort
		dial, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Println("Cannot create grpc client", err)
		}
		defer dial.Close()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		pb.NewLeaderElectionClient(dial)
		defer cancel()
		hb := pb.HeartbeatRequest{
			Term:     es.currentTerm,
			LeaderId: es.leaderID,
		}
		log.Println("Sending heartbeat to ", url)
		heartbeat, err := es.Heartbeat(ctx, &hb)
		if err != nil {
			fmt.Println("Cannot send heartbeat to ", url, err)
		}
		fmt.Println(heartbeat)
	}
}
