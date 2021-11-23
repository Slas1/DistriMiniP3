package main

import (
	"context"
	"flag"
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
	ServerID        	int32
	HighestServerID 	int32
	AllServerIDs   		[]bool
	LamportTime 		lamportTime
	Clients				[]replicatepb.ServerCommunicationClient

	ServerLogFile os.File
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
//RPC Server to Server Calls - Server Funktions
func (s *Server) JoinCommunication(ctx context.Context, request *replicatepb.JoinRequest) (*replicatepb.JoinResponse, error) {
	log.SetOutput(&s.ServerLogFile)
	s.LamportTime.update(int(request.LamportTime))

	s.HighestServerID++
	s.AllServerIDs[s.HighestServerID] = true
	s.LamportTime.increment()

	fmt.Printf("Server with ID: %s - Time: %s - Joined the ServerCommunication.\n", strconv.Itoa(int(s.HighestServerID)), strconv.Itoa(s.LamportTime.time))
	log.Printf("Server with ID: %s - Time: %s - Joined the ServerCommunication.\n", strconv.Itoa(int(s.HighestServerID)), strconv.Itoa(s.LamportTime.time))

	for i := 1; i < int(s.HighestServerID); i++ {
		if s.AllServerIDs[i] {
			s.serverInformationUpdate(s.Clients[i], replicatepb.SIUpdateRequest{LamportTime: int32(s.LamportTime.time), ID: s.ServerID, ChangeName: "ServerJoined", NewValue: []int32{s.HighestServerID}})
		}	
	}

	return &replicatepb.JoinResponse{LamportTime: int32(s.LamportTime.time), ID: s.ServerID, HighestServerID: s.HighestServerID, AllServerIDs: s.AllServerIDs}, nil
}

func (s *Server) ServerInformationUpdate(ctx context.Context, request *replicatepb.SIUpdateRequest) (*replicatepb.SIUpdateResponse, error) {
	log.SetOutput(&s.ServerLogFile)
	s.LamportTime.update(int(request.LamportTime))

	switch request.ChangeName {
		case "ChangeYourServerID":
			s.ServerID = request.NewValue[0]
		case "ServerJoined":
			s.AllServerIDs[request.NewValue[0]] = true
		case "ServerLeft":
			s.AllServerIDs[request.NewValue[0]] = false
		case "NewAllServerList":
			for i := 0; i < len(s.AllServerIDs) ; i++ {
				if request.NewValue[i] == 1 {
					s.AllServerIDs[i] = true
				}else {
					s.AllServerIDs[i] = false
				}				
			}
		case "CreateNewClient":
			s.createNewServerClient(int(request.ID))
	}
	
	fmt.Printf("Server with Id: %s - Time: %s - Parsed the %s change.\n", strconv.Itoa(int(s.ServerID)), strconv.Itoa(s.LamportTime.time), request.ChangeName)
	log.Printf("Server with Id: %s - Time: %s - Parsed the %s change.\n", strconv.Itoa(int(s.ServerID)), strconv.Itoa(s.LamportTime.time), request.ChangeName)
	s.LamportTime.increment()
	return &replicatepb.SIUpdateResponse{LamportTime: int32(s.LamportTime.time), ID: s.ServerID}, nil
}

func (s *Server) AuctionInformationUpdate(ctx context.Context, request *replicatepb.AIUpdateRequest) (*replicatepb.AIUpdateResponse, error) {
	log.SetOutput(&s.ServerLogFile)
	s.LamportTime.update(int(request.LamportTime))
	
	//To-do

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
	
	s.ServerID = response.HighestServerID
	s.HighestServerID = response.HighestServerID
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

	response, err := client.AuctionInformationUpdate(context.Background(), &request)
	if err != nil {
	}

	s.LamportTime.update(int(response.LamportTime))
}
func (s *Server) isAlive(client replicatepb.ServerCommunicationClient, request replicatepb.Poke) {
	s.LamportTime.increment()
	request.LamportTime = int32(s.LamportTime.time)

	fmt.Printf("Server with Id: %s - Time: %s - Pokes to check, if target is alive.\n", strconv.Itoa(int(s.ServerID)), strconv.Itoa(s.LamportTime.time))
	log.Printf("Server with Id: %s - Time: %s - Pokes to check, if target is alive.\n", strconv.Itoa(int(s.ServerID)), strconv.Itoa(s.LamportTime.time))


	response, err := client.IsAlive(context.Background(), &request)
	if err != nil {
	}

	s.LamportTime.update(int(response.LamportTime))
}
//
//
//
//Helping functions
func (s *Server) changeServerID(NewServerID int32){
	s.ServerID = NewServerID
	s.changeHostPort()
}

func (s *Server) changeHostPort() {
	grpcServer.GracefulStop()
	listener.Close()

	listener, err := net.Listen("tcp", "localhost:" + strconv.Itoa(8080 + int(s.ServerID)))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcS := grpc.NewServer(opts...)
	grpcServer := grpcS
	replicatepb.RegisterServerCommunicationServer(grpcServer, s)
	grpcServer.Serve(listener)

}

func (s *Server) createNewServerClient(TargetServerID int){
	var hostString = ":" + strconv.Itoa(8080+TargetServerID)
	var key = "server" + strconv.Itoa(TargetServerID)
	var tcpServer = flag.String(key, hostString , "TCP server")
	var callerOpts []grpc.DialOption
	callerOpts = append(callerOpts, grpc.WithBlock(), grpc.WithInsecure())
	conn, err := grpc.Dial(*tcpServer, callerOpts...)
	if err != nil {
		log.Fatalf("Invalid Target Server. Not posible to dial.\n")
	}
	client := replicatepb.NewServerCommunicationClient(conn)
	s.Clients[TargetServerID] = client

	
	go func()  {
		for{
			time.Sleep(20 * time.Second)
			pokeAction(*s, TargetServerID)
		}
	}()
	
}
//
//
//
//Initialisation
func newServer() *Server {
	s := &Server{
		LamportTime: lamportTime{0, new(sync.Mutex)},
		ServerID: -1,
		HighestServerID: -1,
		AllServerIDs: make([]bool, 20),
		Clients: make([]replicatepb.ServerCommunicationClient, 20),
	}
	return s
}
func newClient(s Server) string{
	s.serverInformationUpdate(s.Clients[0], replicatepb.SIUpdateRequest{LamportTime: int32(s.LamportTime.time), ID: s.ServerID, ChangeName: "CreateNewClient"})
	return "true"
}
func pokeAction(s Server, targetServerID int) string{
	s.isAlive(s.Clients[targetServerID], replicatepb.Poke{LamportTime: int32(s.LamportTime.time),ID: s.ServerID})
	return "true"
}

func main(){
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
	var tcpServer = flag.String("server", ":8080", "TCP server")
	var callerOpts []grpc.DialOption
	callerOpts = append(callerOpts, grpc.WithBlock(), grpc.WithTimeout(time.Duration(2)*time.Second), grpc.WithInsecure())
	conn, err := grpc.Dial(*tcpServer, callerOpts...)
	if err != nil {
		log.Printf("Fail to dial leader when joining, setting myself to leader\n")
		fmt.Printf("Fail to dial leader when joining, setting myself to leader\n")
		s.ServerID = 0
		s.HighestServerID = 0
		s.AllServerIDs[0] = true
	}else {
		client := replicatepb.NewServerCommunicationClient(conn)
		s.Clients[0] = client
		s.joinCommunication(s.Clients[0], replicatepb.JoinRequest{LamportTime: int32(s.LamportTime.time), ID: s.ServerID})
		parallel := make(chan string)
		go func() {
			parallel <- newClient(*s)
		}()
		
		
		go func()  {
			for{
				time.Sleep(30 * time.Second)
				pokeAction(*s, 0)
			}
		}()
		
		
	}

	//Setting up Listener to the correct port in regards to ServerID
	var localhost = "localhost:" + strconv.Itoa(8080 + int(s.ServerID))
	lis, err := net.Listen("tcp", localhost)
	listener := lis
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcS := grpc.NewServer(opts...)
	grpcServer := grpcS
	replicatepb.RegisterServerCommunicationServer(grpcServer, s)
	fmt.Printf("Server with ID: %s - Is up and listening.\n", strconv.Itoa(int(s.ServerID)))
	log.Printf("Server with ID: %s - Is up and listening.\n", strconv.Itoa(int(s.ServerID)))
	grpcServer.Serve(listener)
}