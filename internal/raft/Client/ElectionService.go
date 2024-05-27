package raft

import (
	pb "cache/internal/election"
	"cache/internal/validation"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"log"
	"strconv"
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
	sync.Mutex
	currentTerm        int32
	votedFor           string
	leaderID           string
	electionInProgress bool
	client             *Client
	electionTimer      *time.Timer
	electionTimeout    time.Duration
}

func NewElectionService(client *Client) *ElectionService {

	duration := time.Duration(rangeIn(5, 10)) * time.Second
	return &ElectionService{
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
	es.Lock()
	defer es.Unlock()
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
	if err := validation.ValidateJWTFromContext(es.client.ServiceRegistry.Secret, ctx); err != nil {
		return &pb.HeartbeatResponse{Success: false, Term: es.currentTerm}, err
	}
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
		log.Printf("election timer expired . starting election %v at node %s", time.Now(), es.client.NodeDetails.NodePort)
		es.startElection()
	})
}

func (es *ElectionService) startElection() {
	if es.electionInProgress {
		fmt.Println("Election already in progress")
		return
	}
	es.Lock()
	defer es.Unlock()
	es.currentTerm++
	es.votedFor = "self"
	es.electionInProgress = true
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
	quorumAchieved := gatheredVotes >= (nodesLen / 2)
	if quorumAchieved == true {
		es.leaderID = es.client.NodeDetails.NodePort
		log.Println("Election success , Congratulations !!")
		go es.SendHeartbeats()
	}
	es.electionInProgress = false
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
			es.client.Lock()
			if !es.client.IsLeader() || !es.electionInProgress {
				log.Println("Starting election")
				es.startElection()
			}
			es.client.Unlock()
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
	es.Lock()
	defer es.Unlock()

	for _, node := range es.client.ClusterMembers {
		if node.GrpcPort == es.client.NodeDetails.GrpcPort {
			continue
		}
		token := getEncodedToken(es)
		url := node.NodeAddr + ":" + node.GrpcPort
		log.Println("Sending heartbeat to ", url)
		dial, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
		defer dial.Close()
		if err != nil {
			log.Println("Cannot create grpc client", err)
		}
		client := pb.NewLeaderElectionClient(dial)
		ctx, _ := context.WithTimeout(context.Background(), time.Second)
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
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

func getEncodedToken(es *ElectionService) string {
	secret := es.client.ServiceRegistry.Secret
	token := validation.GetToken(secret, strconv.FormatUint(es.client.ClusterID, 10), es.client.NodeDetails.GrpcPort)
	fmt.Println("Token: ", token)
	return token
}
