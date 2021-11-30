package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"replicatepb/replicatepb"

	"google.golang.org/grpc"
)

//
//
//
//DataStructures
var listener net.Listener
var grpcServer *grpc.Server

type Server struct {
	replicatepb.UnimplementedServerCommunicationServer
	ServerID        int32
	HighestServerID int32
	AllServerIDs    []bool
	LamportTime     lamportTime
	Clients         []replicatepb.ServerCommunicationClient

	HighestBidderID  int32
	HigestBid        int32
	isAuctionOnGoing bool
	HistestClientID  int32

	ServerLogFile  os.File
	AuctionLogFile os.File
}

type lamportTime struct {
	time int
	*sync.Mutex
}

//
//
//
//Helper functions LampartTime
func (lt *lamportTime) max(otherValue int) int {
	if lt.time > otherValue {
		return lt.time
	}
	return otherValue
}

func (lt *lamportTime) update(otherValue int) {
	lt.Lock()

	lt.time = lt.max(otherValue) + 1

	lt.Unlock()
}

func (lt *lamportTime) increment() {
	lt.Lock()

	lt.time++

	lt.Unlock()
}

//
//
//
// Auction Server Functions
func (s *Server) JoinAuction(ctx context.Context, request *replicatepb.JoinRequest) (*replicatepb.JoinResponse, error) {
	log.SetOutput(&s.AuctionLogFile)
	s.LamportTime.update(int(request.LamportTime))

	s.HistestClientID++
	fmt.Printf("AuctionClient with ID: %s - Time: %s - Joined the AuctionCommunication.\n", strconv.Itoa(int(s.HistestClientID)), strconv.Itoa(s.LamportTime.time))
	log.Printf("AuctionClient with ID: %s - Time: %s - Joined the AuctionCommunication.\n", strconv.Itoa(int(s.HistestClientID)), strconv.Itoa(s.LamportTime.time))

	s.LamportTime.increment()
	return &replicatepb.JoinResponse{LamportTime: int32(s.LamportTime.time), YourNewID: s.HistestClientID}, nil
}
func (s *Server) NewBid(ctx context.Context, request *replicatepb.BidRequest) (*replicatepb.BidResponse, error) {
	log.SetOutput(&s.AuctionLogFile)
	s.LamportTime.update(int(request.LamportTime))

	s.LamportTime.increment()
	if request.BidValue > s.HigestBid {
		s.HigestBid = request.BidValue
		s.HighestBidderID = request.ID
		fmt.Printf("AuctionClient with ID: %s - Time: %s - Bid is higher than current, and is set to the highest.\n", strconv.Itoa(int(s.HistestClientID)), strconv.Itoa(s.LamportTime.time))
		log.Printf("AuctionClient with ID: %s - Time: %s - Bid is higher than current, and is set to the highest.\n", strconv.Itoa(int(s.HistestClientID)), strconv.Itoa(s.LamportTime.time))
		for i := 1; i < len(s.AllServerIDs); i++ {
			if s.AllServerIDs[i] {
				s.auctionInformationUpdate(s.Clients[i], replicatepb.AIUpdateRequest{LamportTime: int32(s.LamportTime.time), ID: s.ServerID, AuctionStatus: s.isAuctionOnGoing, BidderID: request.ID, BidValue: request.BidValue})
			}
		}
		return &replicatepb.BidResponse{LamportTime: int32(s.LamportTime.time), YourBidStatus: "Success"}, nil

	} else if request.BidValue < s.HigestBid+1 {
		fmt.Printf("AuctionClient with ID: %s - Time: %s - Bid was to low, not recorded.\n", strconv.Itoa(int(s.HistestClientID)), strconv.Itoa(s.LamportTime.time))
		log.Printf("AuctionClient with ID: %s - Time: %s - Bid was to low, not recorded.\n", strconv.Itoa(int(s.HistestClientID)), strconv.Itoa(s.LamportTime.time))
		return &replicatepb.BidResponse{LamportTime: int32(s.LamportTime.time), YourBidStatus: "Fail"}, nil
	} else {
		fmt.Printf("Error when dealing with bid from AuctionClient with ID: %s - Time: %s\n", strconv.Itoa(int(s.HistestClientID)), strconv.Itoa(s.LamportTime.time))
		log.Printf("Error when dealing with bid from AuctionClient with ID: %s - Time: %s\n", strconv.Itoa(int(s.HistestClientID)), strconv.Itoa(s.LamportTime.time))
		return &replicatepb.BidResponse{LamportTime: int32(s.LamportTime.time), YourBidStatus: "Exception"}, nil
	}
}

func (s *Server) Result(ctx context.Context, request replicatepb.ResultRequest) (*replicatepb.ResultResponse, error) {
	log.SetOutput(&s.AuctionLogFile)
	s.LamportTime.update(int(request.LamportTime))

	fmt.Printf("AuctionClient with ID: %s - Time: %s - Asked for auction Result.\n", strconv.Itoa(int(s.HistestClientID)), strconv.Itoa(s.LamportTime.time))
	log.Printf("AuctionClient with ID: %s - Time: %s - Asked for auction Result.\n", strconv.Itoa(int(s.HistestClientID)), strconv.Itoa(s.LamportTime.time))
		
	s.LamportTime.increment()
	return &replicatepb.ResultResponse{LamportTime: int32(s.LamportTime.time), AuctionOngoing: s.isAuctionOnGoing, HighestBid: s.HigestBid}, nil
}

//
//
//
//RPC Server to Server Calls - Server Funktions
func (s *Server) JoinCommunication(ctx context.Context, request *replicatepb.JoinRequest) (*replicatepb.JoinResponse, error) {
	log.SetOutput(&s.ServerLogFile)
	s.LamportTime.update(int(request.LamportTime))

	var lowestFreeID int
	for i := 1; i < len(s.AllServerIDs); i++ {
		if !s.AllServerIDs[i] {
			lowestFreeID = i
			break
		}
	}

	if lowestFreeID > int(s.HighestServerID) {
		s.HighestServerID = int32(lowestFreeID)
	}
	s.AllServerIDs[lowestFreeID] = true
	s.LamportTime.increment()

	fmt.Printf("Server with ID: %s - Time: %s - Joined the ServerCommunication.\n", strconv.Itoa(lowestFreeID), strconv.Itoa(s.LamportTime.time))
	log.Printf("Server with ID: %s - Time: %s - Joined the ServerCommunication.\n", strconv.Itoa(lowestFreeID), strconv.Itoa(s.LamportTime.time))

	for i := 1; i < len(s.AllServerIDs); i++ {
		if s.AllServerIDs[i] && i != lowestFreeID {
			s.serverInformationUpdate(s.Clients[i], replicatepb.SIUpdateRequest{LamportTime: int32(s.LamportTime.time), ID: s.ServerID, ChangeName: "ServerJoined", NewValue: []int32{int32(lowestFreeID)}})
		}
	}

	return &replicatepb.JoinResponse{LamportTime: int32(s.LamportTime.time), ID: s.ServerID, YourNewID: int32(lowestFreeID), AllServerIDs: s.AllServerIDs}, nil
}

func (s *Server) ServerInformationUpdate(ctx context.Context, request *replicatepb.SIUpdateRequest) (*replicatepb.SIUpdateResponse, error) {
	log.SetOutput(&s.ServerLogFile)
	s.LamportTime.update(int(request.LamportTime))

	switch request.ChangeName {
	case "ChangeYourServerID":
		s.ServerID = request.NewValue[0]
	case "ServerJoined":
		s.AllServerIDs[request.NewValue[0]] = true
		if request.NewValue[0] > s.HighestServerID {
			s.HighestServerID = request.NewValue[0]
		}
	case "ServerLeft":
		s.AllServerIDs[request.NewValue[0]] = false
		if request.NewValue[0] == s.HighestServerID {
			for i, v := range s.AllServerIDs {
				if v {
					s.HighestServerID = int32(i)
				}
			}
		}
	case "NewAllServerList":
		for i := 0; i < len(s.AllServerIDs); i++ {
			if request.NewValue[i] == 1 {
				s.AllServerIDs[i] = true
			} else {
				s.AllServerIDs[i] = false
			}
		}
	case "CreateNewClient":
		s.createNewServerClient(int(request.ID))
	case "ImTheNewKing":
		s.AllServerIDs[0] = true
		s.AllServerIDs[request.ID] = false
		s.createNewServerClient(0)
	}

	fmt.Printf("Server with Id: %s - Time: %s - Parsed the %s change.\n", strconv.Itoa(int(s.ServerID)), strconv.Itoa(s.LamportTime.time), request.ChangeName)
	log.Printf("Server with Id: %s - Time: %s - Parsed the %s change.\n", strconv.Itoa(int(s.ServerID)), strconv.Itoa(s.LamportTime.time), request.ChangeName)
	s.LamportTime.increment()
	return &replicatepb.SIUpdateResponse{LamportTime: int32(s.LamportTime.time), ID: s.ServerID}, nil
}

func (s *Server) AuctionInformationUpdate(ctx context.Context, request *replicatepb.AIUpdateRequest) (*replicatepb.AIUpdateResponse, error) {
	log.SetOutput(&s.ServerLogFile)
	s.LamportTime.update(int(request.LamportTime))

	s.isAuctionOnGoing = request.AuctionStatus
	s.HighestBidderID = request.BidderID
	s.HigestBid = request.BidValue

	fmt.Printf("Server with Id: %s - Time: %s - Parsed the AuctionInformation change.\n", strconv.Itoa(int(s.ServerID)), strconv.Itoa(s.LamportTime.time))
	log.Printf("Server with Id: %s - Time: %s - Parsed the AuctionInformation change.\n", strconv.Itoa(int(s.ServerID)), strconv.Itoa(s.LamportTime.time))

	s.LamportTime.increment()
	return &replicatepb.AIUpdateResponse{LamportTime: int32(s.LamportTime.time), ID: s.ServerID}, nil
}

func (s *Server) IsAlive(ctx context.Context, request *replicatepb.Poke) (*replicatepb.HandSign, error) {
	log.SetOutput(&s.ServerLogFile)

	s.LamportTime.update(int(request.LamportTime))
	s.LamportTime.increment()

	fmt.Printf("Server with Id: %s - Time: %s - Throws Hand Sign and is still alive.\n", strconv.Itoa(int(s.ServerID)), strconv.Itoa(s.LamportTime.time))
	log.Printf("Server with Id: %s - Time: %s - Throws Hand Sign and is still alive.\n", strconv.Itoa(int(s.ServerID)), strconv.Itoa(s.LamportTime.time))

	return &replicatepb.HandSign{LamportTime: int32(s.LamportTime.time), ID: s.ServerID}, nil
}

//
//
//
//RPC Server to Server - Client funktions
func (s *Server) joinCommunication(client replicatepb.ServerCommunicationClient, request replicatepb.JoinRequest) {
	s.LamportTime.increment()
	request.LamportTime = int32(s.LamportTime.time)

	fmt.Printf("New Server - Sending Joining-Request.\n")
	log.Printf("New Server - Sending Joining-Request.\n")

	response, err := client.JoinCommunication(context.Background(), &request)
	if err != nil {
	}

	s.ServerID = response.YourNewID
	for i := 1; i < len(s.AllServerIDs); i++ {
		if s.AllServerIDs[i] {
			s.HighestServerID = int32(i)

		}
	}
	s.AllServerIDs = response.AllServerIDs
	s.LamportTime.update(int(response.LamportTime))
	fmt.Printf("Server with Id: %s - Time: %s - Initial values set based on JoinResponse.\n", strconv.Itoa(int(s.ServerID)), strconv.Itoa(s.LamportTime.time))
	log.Printf("Server with Id: %s - Time: %s - Initial values set based on JoinResponse.\n", strconv.Itoa(int(s.ServerID)), strconv.Itoa(s.LamportTime.time))
}
func (s *Server) serverInformationUpdate(client replicatepb.ServerCommunicationClient, request replicatepb.SIUpdateRequest) {
	s.LamportTime.increment()
	request.LamportTime = int32(s.LamportTime.time)

	fmt.Printf("Server with Id: %s - Time: %s - Sends update-request with the change: %s.\n", strconv.Itoa(int(s.ServerID)), strconv.Itoa(s.LamportTime.time), request.ChangeName)
	log.Printf("Server with Id: %s - Time: %s - Sends update-request with the change: %s.\n", strconv.Itoa(int(s.ServerID)), strconv.Itoa(s.LamportTime.time), request.ChangeName)

	response, err := client.ServerInformationUpdate(context.Background(), &request)
	if err != nil {
	}

	s.LamportTime.update(int(response.LamportTime))
}
func (s *Server) auctionInformationUpdate(client replicatepb.ServerCommunicationClient, request replicatepb.AIUpdateRequest) {
	s.LamportTime.increment()
	request.LamportTime = int32(s.LamportTime.time)

	fmt.Printf("Server with Id: %s - Time: %s - Sends update-request with the change in Auction.\n", strconv.Itoa(int(s.ServerID)), strconv.Itoa(s.LamportTime.time))
	log.Printf("Server with Id: %s - Time: %s - Sends update-request with the change in Auction.\n", strconv.Itoa(int(s.ServerID)), strconv.Itoa(s.LamportTime.time))


	response, err := client.AuctionInformationUpdate(context.Background(), &request)
	if err != nil {
	}

	s.LamportTime.update(int(response.LamportTime))
}
func (s *Server) isAlive(client replicatepb.ServerCommunicationClient, request replicatepb.Poke) bool {
	s.LamportTime.increment()
	request.LamportTime = int32(s.LamportTime.time)

	fmt.Printf("Server with Id: %s - Time: %s - Pokes to check, if target is alive.\n", strconv.Itoa(int(s.ServerID)), strconv.Itoa(s.LamportTime.time))
	log.Printf("Server with Id: %s - Time: %s - Pokes to check, if target is alive.\n", strconv.Itoa(int(s.ServerID)), strconv.Itoa(s.LamportTime.time))

	if client == nil {
		return false
	}
	response, err := client.IsAlive(context.Background(), &request)
	if err != nil {
		fmt.Println("Error i poke. The Target is Dead.")
		log.Printf("Error i poke. The Target is Dead.")
		return false
	} else if response != nil {
		s.LamportTime.update(int(response.LamportTime))
		return true
	}
	return false
}

//
//
//
//Helping functions
func (s *Server) changeServerID(NewServerID int32) {
	s.AllServerIDs[s.ServerID] = false

	s.changeHostPort(int(NewServerID))

}

func (s *Server) changeHostPort(NewServerID int) {
	lis, err := net.Listen("tcp", "localhost:"+strconv.Itoa(8080+NewServerID))
	listener = lis
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcS := grpc.NewServer(opts...)
	grpcServer = grpcS
	replicatepb.RegisterServerCommunicationServer(grpcServer, s)

	go func() {
		time.Sleep(time.Second * 2)
		s.AllServerIDs[0] = true

		for i := 1; i < len(s.AllServerIDs); i++ {
			if s.AllServerIDs[i] {
				s.createNewServerClient(i)
			}
		}
		for i := 1; i < len(s.AllServerIDs); i++ {
			if s.AllServerIDs[i] {
				s.serverInformationUpdate(s.Clients[i], replicatepb.SIUpdateRequest{LamportTime: int32(s.LamportTime.time), ID: s.ServerID, ChangeName: "ImTheNewKing"})
			}
		}
		fmt.Printf("Server with Id: %s - Time: %s - Changes ID to 0.\n", strconv.Itoa(int(s.ServerID)), strconv.Itoa(s.LamportTime.time))
		log.Printf("Server with Id: %s - Time: %s - Changes ID to 0.\n", strconv.Itoa(int(s.ServerID)), strconv.Itoa(s.LamportTime.time))
		s.ServerID = int32(NewServerID)
	}()
	grpcServer.Serve(listener)
}

func (s *Server) createNewServerClient(TargetServerID int) {
	var hostString = ":" + strconv.Itoa(8080+TargetServerID)
	var callerOpts []grpc.DialOption
	callerOpts = append(callerOpts, grpc.WithBlock(), grpc.WithInsecure())
	conn, err := grpc.Dial(hostString, callerOpts...)
	if err != nil {
		log.Fatalf("Invalid Target Server. Not posible to dial.\n")
	}
	client := replicatepb.NewServerCommunicationClient(conn)
	s.Clients[TargetServerID] = client

	go func() {
		for {
			time.Sleep(10 * time.Second)
			result := pokeAction(*s, TargetServerID)
			if !result {
				break
			}
		}
	}()

}

func (s *Server) TheKingIsDead() {
	var lowestID int
	for i := 1; i < len(s.AllServerIDs); i++ {
		if s.AllServerIDs[i] {
			lowestID = i
			break
		}
	}
	if lowestID == int(s.ServerID) {
		s.changeServerID(0)
	}
}

//
//
//
//Initialisation
func newServer() *Server {
	s := &Server{
		LamportTime:     lamportTime{0, new(sync.Mutex)},
		ServerID:        -1,
		HighestServerID: -1,
		AllServerIDs:    make([]bool, 20),
		Clients:         make([]replicatepb.ServerCommunicationClient, 20),
	}
	return s
}
func newClient(s Server) string {
	s.serverInformationUpdate(s.Clients[0], replicatepb.SIUpdateRequest{LamportTime: int32(s.LamportTime.time), ID: s.ServerID, ChangeName: "CreateNewClient"})
	return "true"
}
func pokeAction(s Server, targetServerID int) bool {
	result := s.isAlive(s.Clients[targetServerID], replicatepb.Poke{LamportTime: int32(s.LamportTime.time), ID: s.ServerID})
	if !result {
		s.Clients[targetServerID] = nil
		s.AllServerIDs[targetServerID] = false
		if targetServerID == 0 {
			s.AllServerIDs[0] = false
			s.Clients[0] = nil
			s.TheKingIsDead()
		} else {
			for i := 1; i < len(s.AllServerIDs); i++ {
				if s.ServerID == 0 && s.AllServerIDs[i] {
					s.serverInformationUpdate(s.Clients[i], replicatepb.SIUpdateRequest{LamportTime: int32(s.LamportTime.time), ID: s.ServerID, ChangeName: "ServerLeft", NewValue: []int32{int32(targetServerID)}})
				}
			}
		}
	}
	return result
}

func main() {
	s := newServer()

	//Log Files
	LOG_FILE_SERVERCOM := "./ServerComsLog"
	LOG_FILE_AUCTION := "./AuctionLog"
	logFileServerCom, err := os.OpenFile(LOG_FILE_SERVERCOM, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	logFileAuction, err := os.OpenFile(LOG_FILE_AUCTION, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}
	s.ServerLogFile = *logFileServerCom
	s.AuctionLogFile = *logFileAuction
	defer logFileServerCom.Close()
	defer logFileAuction.Close()
	log.SetFlags(log.Lmicroseconds)
	log.SetOutput(&s.ServerLogFile)

	//Initial dial to port 8080. If there is no Server, the caller becomes the Leader.
	var tcpServer = ":8080"
	var callerOpts []grpc.DialOption
	callerOpts = append(callerOpts, grpc.WithBlock(), grpc.WithTimeout(time.Duration(2)*time.Second), grpc.WithInsecure())
	conn, err := grpc.Dial(tcpServer, callerOpts...)
	if err != nil {
		log.Printf("Fail to dial leader when joining, setting myself to leader\n")
		fmt.Printf("Fail to dial leader when joining, setting myself to leader\n")
		s.ServerID = 0
		s.HighestServerID = 0
		s.AllServerIDs[0] = true
	} else {
		client := replicatepb.NewServerCommunicationClient(conn)
		s.Clients[0] = client
		s.AllServerIDs[0] = true
		s.joinCommunication(s.Clients[0], replicatepb.JoinRequest{LamportTime: int32(s.LamportTime.time), ID: s.ServerID})
		parallel := make(chan string)
		go func() {
			parallel <- newClient(*s)
		}()

		go func() {
			for {
				time.Sleep(10 * time.Second)
				result := pokeAction(*s, 0)
				if !result {
					break
				}
			}
		}()

	}

	//Setting up Listener to the correct port in regards to ServerID
	var localhost = "localhost:" + strconv.Itoa(8080+int(s.ServerID))
	lis, err := net.Listen("tcp", localhost)
	listener = lis
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcS := grpc.NewServer(opts...)
	grpcServer = grpcS
	replicatepb.RegisterServerCommunicationServer(grpcServer, s)
	fmt.Printf("Server with ID: %s - Is up and listening.\n", strconv.Itoa(int(s.ServerID)))
	log.Printf("Server with ID: %s - Is up and listening.\n", strconv.Itoa(int(s.ServerID)))
	grpcServer.Serve(listener)
}
