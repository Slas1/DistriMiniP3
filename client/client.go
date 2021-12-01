package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"replicatepb/replicatepb"
	"strconv"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
)

type Client struct {
	ClientId         int32
	AreYouWinningSon bool
	MyHighestBid     int32
	isAuctionOnGoing bool

	LamportTime lamportTime

	Client replicatepb.AuctionCommunicationClient
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

func (c *Client) joinAuction(request replicatepb.JoinRequest) {
	c.LamportTime.increment()
	request.LamportTime = int32(c.LamportTime.time)

	c.crashTest()
	response, err := c.Client.JoinAuction(context.Background(), &request)
	if err != nil {
		log.Fatalf("Client with id: %s - Error when calling JoinAuction: %s", strconv.Itoa(int(request.ID)), err)
		fmt.Printf("Client with id: %s - Error when calling JoinAuction: %s\n", strconv.Itoa(int(request.ID)), err)
	}
	c.ClientId = response.YourNewID
	log.Printf("Client with id: %s - Got ID from server", strconv.Itoa(int(request.ID)))
	fmt.Printf("Client with id: %s - Got ID from server\n", strconv.Itoa(int(request.ID)))

	c.LamportTime.update(int(response.LamportTime))
}

func (c *Client) newBid(request replicatepb.BidRequest) {
	c.LamportTime.increment()
	request.LamportTime = int32(c.LamportTime.time)

	c.crashTest()
	response, err := c.Client.NewBid(context.Background(), &request)
	if err != nil {
		log.Fatalf("Client with id: %s - Error when calling newBid: %s", strconv.Itoa(int(request.ID)), err)
		fmt.Printf("Client with id: %s - Error when calling newBid: %s\n", strconv.Itoa(int(request.ID)), err)
	}

	log.Printf("Client with id: %s - Submitted a bid and got status: %v\n", strconv.Itoa(int(request.ID)), response.YourBidStatus)
	fmt.Printf("Client with id: %s - Submitted a bid and got status: %v\n", strconv.Itoa(int(c.ClientId)), response.YourBidStatus)

	if response.YourBidStatus == "Success" {
		c.AreYouWinningSon = true
		c.MyHighestBid = request.BidValue
	} else if response.YourBidStatus == "Fail" {
		c.AreYouWinningSon = false
		c.MyHighestBid = request.BidValue
	}

	c.LamportTime.update(int(response.LamportTime))
}

func (c *Client) result(request replicatepb.ResultRequest) {
	c.LamportTime.increment()
	request.LamportTime = int32(c.LamportTime.time)

	c.crashTest()
	response, err := c.Client.Result(context.Background(), &request)
	if err != nil {
		log.Fatalf("Client with id: %s - Error when calling result: %s", strconv.Itoa(int(c.ClientId)), err)
		fmt.Printf("Client with id: %s - Error when calling result: %s\n", strconv.Itoa(int(c.ClientId)), err)
	}

	log.Printf("Client with id: %s - Asked for status - IsActionOngion: %v, HighestBid: %v\n", strconv.Itoa(int(c.ClientId)), response.AuctionOngoing, response.HighestBid)
	fmt.Printf("Client with id: %s - Asked for status - IsActionOngion: %v, HighestBid: %v\n", strconv.Itoa(int(c.ClientId)), response.AuctionOngoing, response.HighestBid)

	c.isAuctionOnGoing = response.AuctionOngoing
	if c.MyHighestBid == response.HighestBid {
		c.AreYouWinningSon = true
	}
	c.LamportTime.update(int(response.LamportTime))
}

//
//
//
//Helper function for crash save communication
func (c *Client) crashTest() {
	if c.Client == nil {
		client, err := createClient(0)
		if err != nil {
		}

		c.Client = client
	}
}

//
//
//
//Initialisation helper functions:
func createClient(TimesTriedConnecting int) (client replicatepb.AuctionCommunicationClient, err error) {
	var tcpServer = ":8080"
	var callerOpts []grpc.DialOption
	callerOpts = append(callerOpts, grpc.WithBlock(), grpc.WithTimeout(time.Duration(2)*time.Second), grpc.WithInsecure())
	conn, err := grpc.Dial(tcpServer, callerOpts...)
	if err != nil {
		if TimesTriedConnecting > 4 {
			log.Printf("Fail to dial Auction when joining, after 5 tries, Stopping\n")
			fmt.Printf("Fail to dial Auction when joining, after 5 tries, Stopping\n")
			return nil, err
		} else {
			log.Printf("Fail to Auction when joining, trying again later.\n")
			fmt.Printf("Fail to Auction when joining, trying again later.\n")
			time.Sleep(time.Second * 2)
			return createClient(TimesTriedConnecting + 1)
		}

	} else {
		return replicatepb.NewAuctionCommunicationClient(conn), nil
	}
}
func createStruct() *Client {
	s := &Client{
		LamportTime:      lamportTime{0, new(sync.Mutex)},
		ClientId:         0,
		AreYouWinningSon: false,
	}
	return s
}

func main() {
	clientStruct := createStruct()
	LOG_FILE_AUCTION := "./AuctionLog"
	logFileAuction, err := os.OpenFile(LOG_FILE_AUCTION, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer logFileAuction.Close()
	log.SetOutput(logFileAuction)
	log.SetFlags(log.Lmicroseconds)

	client, err := createClient(0)
	if err != nil {
	}

	clientStruct.Client = client
	clientStruct.joinAuction(replicatepb.JoinRequest{LamportTime: int32(clientStruct.LamportTime.time)})
	clientStruct.result(replicatepb.ResultRequest{LamportTime: int32(clientStruct.LamportTime.time), ID: clientStruct.ClientId})

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		s, err := strconv.Atoi(input)
		if err == nil {
			go clientStruct.newBid(replicatepb.BidRequest{LamportTime: int32(clientStruct.LamportTime.time), ID: clientStruct.ClientId, BidValue: int32(s)})
		} else {
			if strings.Contains(input, "Result") {
				clientStruct.result(replicatepb.ResultRequest{LamportTime: int32(clientStruct.LamportTime.time), ID: clientStruct.ClientId})
			}
		}
	}

}
