# DistriMiniP3
Distributed Systems 2021 -- Miniproject 3

What to submit on learnit:



- a **single** zip-compressed file containing: a folder src containing

the source code. You are only allowed to submit source code files in

this folder.

- A file report.pdf containing a report (in PDF) with your answers;

the file can be **at most** 5 A4 pages (it can be less), font cannot

be smaller than 9pt. The report has to contain 4 sections (see below

for a detailed specification of the report format). The report cannot

contain source code (except possibly for illustrative code snippets).

- (Optional) a text file log.txt containing log(s).



== A Distributed Auction System ==

You must implement a **distributed auction system** using replication:

a distributed component which handles auctions, and provides

operations for bidding and querying the state of an auction. The

component must faithfully implement the semantics of the system

described below, and must at least be resilient to one (1) crash

failure.


:: API ::

Your system must be implemented as some number of nodes, possibly

running on distinct hosts. Clients direct API requests to any node

they happen to know (it is up to you to decide how many nodes can be

known). Nodes must respond to the following API:



Method: bid

Inputs: amount (an int)

Outputs: ack

Comment: given a bid, returns an outcome among {fail, success or exception}



Method: result

Inputs: void amount (an int)

Ouputs: outcome

Comment: if over, it returns the result, else highest bid. 



:: Semantics ::

Your component must have the following behaviour, for any reasonable

sequentialisation/interleaving of requests to it:

  

- The first call to "bid" registers the bidder.

- Bidders can bid several times, but a bid must be higher than the

  previous one(s).

- after a specified timeframe, the highest bidder ends up as the

  winner of the auction.

- bidders can query the system in order to know the state of the

  auction.



:: Faults :: 

- Assume a network that has reliable, ordered message transport, where

  transmissions to non-failed nodes complete within a known

  time-limit.



- Your component must be resilient to the failure-stop failure of one

   (1) node.



- You may assume that crashes only happen “one at a time”; e.g.,

  between a particular client request and the system’s subsequent

  response, you may assume that at most one crash occurs. However, a

  second crash may still happen during subsequent requests. For

  example, the node receiving a request might crash. On the next

  request, another node in the system might crash.





:: Report ::

Write a report of at most 5 pages containing the following structure

(exactly create four sections as below):


- Introduction. A short introduction to what you have done.

- Protocol. A description of your protocol, including any protocols

  used internally between nodes of the system.

- Correctness 1. An argument that your protocol is correct

  in the absence of failures.

- Correctness 2. An argument that your protocol is correct in the

  presence of failures.



:: Implementation :: 

- Implement your system in GoLang. We strongly recommend that you

  reuse the the frameworks and libraries used in the previous

  mandatory activites.

- You may submit a log (as a separate file) documenting a correct

  system run under failures. Your log can be a collection of relevant

  print statements, that demonstrates the control flow trough the

  system. It must be clear from the log where crashes occur.
