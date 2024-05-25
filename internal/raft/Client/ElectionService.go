package raft

import (
	pb "cache/internal/election"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
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

	duration := time.Duration(rangeIn(1, 3)) * time.Second
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
	fmt.Println("resetting election timer")
	if es.electionTimer != nil {
		es.electionTimer.Stop()
	}
	es.electionTimer = time.AfterFunc(es.electionTimeout, func() {
		fmt.Println("Starting election timer")
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
		if node.GrpcPort != es.client.NodeDetails.GrpcPort {
			voting, err := es.initiatingVoting(node, req)
			if err != nil {
				fmt.Println("error while initiating voting for ", node.NodePort)
				continue
			}
			if voting.VoteGranted == true {
				gatheredVotes = gatheredVotes + 1
			}
		}
	}
	b := gatheredVotes >= (nodesLen / 2)
	if b == true {
		es.leaderID = es.client.NodeDetails.NodePort
		log.Println("Election success , Congratulations !!")
		go es.SendHeartbeats()
	}
	es.mu.Lock()
	es.electionInProgress = false
	es.mu.Unlock()
}

func (es *ElectionService) initiatingVoting(node *ClusterPeer, vr *pb.VoteRequest) (*pb.VoteResponse, error) {
	dial, err := grpc.NewClient(node.NodeAddr+":"+node.GrpcPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer dial.Close()

	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	pb.NewLeaderElectionClient(dial)
	client := pb.NewLeaderElectionClient(dial)
	defer cancel()
	re, err := client.RequestVote(ctx, vr)
	if err != nil {
		fmt.Println("Error requesting vote", err)
	}
	return re, err
}

func (es *ElectionService) RunElectionLoop() error {
	for {
		select {
		case <-es.electionTimer.C:
			if es.electionInProgress || es.GetLeaderId() == es.client.NodeDetails.NodePort {
				continue
			}
			log.Println("Leader Ping TimedOut , Election will happen at :", es.client.NodeDetails.NodePort)
			es.client.RMu.Lock()
			if !es.client.IsLeader() || !es.electionInProgress {
				log.Println("Starting election")
				es.startElection()
			}
			es.client.RMu.Unlock()
		}

	}
}

func (es *ElectionService) GetLeaderId() string {
	return es.leaderID
}

func (es *ElectionService) SendHeartbeats() {
	timer := time.NewTicker(300 * time.Millisecond)
	for {
		select {
		case <-timer.C:
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
		if node.GrpcPort == es.client.NodeDetails.GrpcPort {
			continue
		}
		url := node.NodeAddr + ":" + node.GrpcPort
		log.Println("Sending heartbeat to ", url)
		dial, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
		defer dial.Close()
		if err != nil {
			log.Println("Cannot create grpc client", err)
		}
		client := pb.NewLeaderElectionClient(dial)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		hb := pb.HeartbeatRequest{
			Term:     es.currentTerm,
			LeaderId: es.leaderID,
		}
		heartbeat, err := client.Heartbeat(ctx, &hb)
		if err != nil {
			fmt.Println("Cannot send heartbeat to ", url, err)
		}
		fmt.Println(heartbeat)
	}
}
