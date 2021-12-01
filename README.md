# Mini_Project3

## How to use (Target: TA)

Limitations: Maximum 20 servers.

1. Start server by typing: "go run server.go" in a terminal while in the server folder. You can create up to 20 servers, remove and add as many times as you want. However please wait for the change to be processed before adding or removeing new servers. (detect time can be improved by lovering the poke action time out, but that will spam the log/terminal with alot of calls.)

1. Start client by typing: "go run client.go" in a terminal while in the client folder. You can now summit a bid by writing a number in the terminal. You can also see the result by typing "Result" in the terminal. 

## Description of Submited LogFile (before TA runs)
The program includes 2 logs. A log for auctionclient and leading server communication called AuctionLog and a log for internal servercommunication called ServerComsLog.
In the logs, we created 3 servers, then killed a backup (ID 2), a new server join and got the lowest avaible ID. (ID 2)
We created one auction client, and bidded. We created another client and over bid. And so on...
Atlast we slowly killed server leader and then next server leader until non were left. 

## == A Distributed Auction System ==
You must implement a **distributed auction system** using replication: a distributed component which handles auctions, and provides operations for bidding and querying the state of an auction. The component must faithfully implement the semantics of the system described below, and must at least be resilient to one (1) crash failure.

#### :: API ::
Your system must be implemented as some number of nodes, possibly running on distinct hosts. Clients direct API requests to any node they happen to know (it is up to you to decide how many nodes can be known). Nodes must respond to the following API:

**Method: bid**
``` sh
Inputs: amount (an int)
Outputs: ack
Comment: given a bid, returns an outcome among {fail, success or exception}
```

**Method: result**
```
Inputs: void amount (an int)
Ouputs: outcome
Comment: if over, it returns the result, else highest bid. 
```

#### :: Semantics ::
Your component must have the following behaviour, for any reasonable sequentialisation/interleaving of requests to it:

- The first call to "bid" registers the bidder.
- Bidders can bid several times, but a bid must be higher than the previous one(s).
- after a specified timeframe, the highest bidder ends up as the winner of the auction.
- bidders can query the system in order to know the state of the auction.

#### :: Faults :: 
- Assume a network that has reliable, ordered message transport, where transmissions to non-failed nodes complete within a known time-limit.
- Your component must be resilient to the failure-stop failure of one (1) node.
- You may assume that crashes only happen “one at a time”; e.g., between a particular client request and the system’s subsequent response, you may assume that at most one crash occurs. However, a second crash may still happen during subsequent requests. For example, the node receiving a request might crash. On the next request, another node in the system might crash.

#### :: Implementation :: 

- Implement your system in GoLang. We strongly recommend that you reuse the the frameworks and libraries used in the previous mandatory activites.
- You may submit a log (as a separate file) documenting a correct system run under failures. Your log can be a collection of relevant print statements, that demonstrates the control flow trough the system. It must be clear from the log where crashes occur.

#### :: Report ::
Write a report of at most 5 pages containing the following structure (exactly create four sections as below):

- Introduction. A short introduction to what you have done.
- Protocol. A description of your protocol, including any protocols used internally between nodes of the system.
- Correctness 1. An argument that your protocol is correct in the absence of failures.
- Correctness 2. An argument that your protocol is correct in the presence of failures.

#### ::Submit::
- a **single** zip-compressed file containing: a folder src containing the source code. You are only allowed to submit source code files in this folder.
- A file report.pdf containing a report (in PDF) with your answers; the file can be **at most** 5 A4 pages (it can be less), font cannot be smaller than 9pt. The report has to contain 4 sections (see above for a detailed specification of the report format). The report cannot contain source code (except possibly for illustrative code snippets).
- (Optional) a text file log.txt containing log(s).
