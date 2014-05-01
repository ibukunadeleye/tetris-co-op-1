package replicaserver

import (
	"centralserver"
	"container/list"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"rpc/replicarpc"
	"strconv"
	"sync"
	"time"
)

type replicaServer struct {
	hostport        string
	ID              int
	numNodes        int
	myN             float64
	Na              float64
	Va              []byte
	V               []byte  //current V bieng sent
	N               float64 //current N being sent
	Nh              float64
	commitLog       map[int]*replicarpc.CommitVal
	replyChan       chan replicarpc.Reply
	replicaList     []string
	httpInit        map[int]string
	httpHandlerMap  map[int]*rpc.Client
	logMutex        *sync.RWMutex
	inChan          chan []byte
	outChan         chan []byte
	testNodeDies_BC bool
	testNodeDies_AC bool
	FailureNode     int
}

// NewReplicaServer creates and starts a new ReplicaServer.
// numNodes is the number of members in the Paxos quorum.
// port is the number of the port that the replicaserver should listen on.
// memberID is the ID the replicaserver should have
// For now, the address of all replicaservers are "localhost".
// This function should return only once all members have joined the quorum
// and should return a non-nil error if the replicaserver could not be started.
func NewReplicaServer(masterServerHostPort string, port, numNodes int) (ReplicaServer, error) {
	fmt.Println("starting up replicaServer port:", port)
	//new replicaServer
	rs := &replicaServer{
		hostport:        ("localhost:" + strconv.Itoa(port)),
		myN:             0.0,
		Na:              0.0,
		Va:              nil,
		V:               nil,
		N:               0.0,
		Nh:              0.0,
		logMutex:        new(sync.RWMutex),
		commitLog:       make(map[int]*replicarpc.CommitVal),
		httpInit:        make(map[int]string),
		httpHandlerMap:  make(map[int]*rpc.Client),
		inChan:          make(chan []byte),
		outChan:         make(chan []byte),
		testNodeDies_BC: true,
		testNodeDies_AC: false,
		FailureNode:     2,
	}

	//start rpc for the replica
	replicaListener, err := net.Listen("tcp", rs.hostport)
	if err != nil {
		return nil, err
	}
	rpc.RegisterName("ReplicaServer", replicarpc.Wrap(rs))
	rpc.HandleHTTP()
	go http.Serve(replicaListener, nil)

	go handleBuffer(rs.inChan, rs.outChan)
	go rs.handleValues()

	//register the replica with the central
	masterRPC, err := rpc.DialHTTP("tcp", masterServerHostPort)
	args := &replicarpc.RegisterArgs{
		Hostport: rs.hostport,
	}
	reply := new(replicarpc.RegisterReply)
	masterRPC.Call("CentralServer.RegisterReplica", args, reply)
	rs.ID = reply.NodeID

	fmt.Println("finished making replicaServer port:", rs.hostport, "node:", rs.ID)
	return rs, nil
}

func (rs *replicaServer) handleValues() {
	for {
		select {
		case v := <-rs.outChan:
			rs.paxosRoutine(v)
		}
	}
}

func (rs *replicaServer) Get(args *replicarpc.GetArgs, reply *replicarpc.GetReply) error {
	fmt.Println("got into get")
	rs.logMutex.RLock()
	commitval, ok := rs.commitLog[len(rs.commitLog)-1]
	if !ok {
		return errors.New("cannot get latest value")
	}
	rs.logMutex.RUnlock()
	reply.V = commitval.V
	return nil
}

func (rs *replicaServer) GetList(args *replicarpc.GetArgs, reply *replicarpc.GetListReply) error {
	return errors.New("not implemented")
}

func (rs *replicaServer) Put(args *replicarpc.PutArgs, reply *replicarpc.PutReply) error {
	rs.inChan <- args.V
	return nil
}

func (rs *replicaServer) AppendToList(args *replicarpc.PutArgs, reply *replicarpc.PutReply) error {
	return errors.New("not implemented")
}

func (rs *replicaServer) RemoveFromList(args *replicarpc.PutArgs, reply *replicarpc.PutReply) error {
	return errors.New("not implemented")
}

func (rs *replicaServer) setup() {
	for i, hostport := range rs.httpInit {
		replicaRPC, _ := rpc.DialHTTP("tcp", hostport)
		rs.httpHandlerMap[i] = replicaRPC
	}
}

func (rs *replicaServer) AddServers(args *replicarpc.AddServersArgs, reply *replicarpc.AddServersReply) error {
	rs.httpInit = args.ServerMap
	rs.setup()
	rs.numNodes = len(args.ServerMap)
	fmt.Println("getting out of addServers")
	return nil
}

func (rs *replicaServer) Prepare(args *replicarpc.PrepareArgs, reply *replicarpc.Reply) error {
	N := args.N
	CS := args.CS
	fmt.Println("got into prepare with N=", N, "CS=", CS)

	rs.logMutex.RLock()
	commitval, ok := rs.commitLog[CS]
	rs.logMutex.RUnlock()

	if ok { //commit slot taken
		fmt.Println("prepare going to return committed")
		reply.OK = true
		reply.Committed = true
		reply.Value = commitval.V
		reply.N = commitval.N
		reply.CS = CS
	} else {
		if rs.Nh > N { //higher N seen
			fmt.Println("prepare going to return false")
			reply.OK = false
			return nil
		} else { //higher N not seen
			fmt.Println("prepare going to return true")
			reply.OK = true
			reply.Committed = false
			reply.CS = CS
			reply.Value = rs.Va
			reply.N = rs.Na
			rs.Nh = N
		}
	}
	return nil
}

func (rs *replicaServer) Accept(args *replicarpc.AcceptArgs, reply *replicarpc.Reply) error {
	N := args.N
	V := args.V
	fmt.Println("got into Accept with value:", V)
	if rs.Nh > N { //higher N seen
		reply.OK = false
	} else { //higher N not seen
		reply.OK = true
		rs.Va = V
		rs.Na = N
		rs.Nh = N
	}

	return nil
}

func (rs *replicaServer) Commit(args *replicarpc.CommitVal, reply *replicarpc.Reply) error {
	fmt.Println("got into commit with value:", args.V)

	if rs.testNodeDies_BC && (rs.FailureNode == rs.ID) {
		fmt.Println("value will not be committed")
		rs.testNodeDies_BC = false
	} else {
		rs.logMutex.RLock()
		logLen := len(rs.commitLog)
		rs.logMutex.RUnlock()
		for logLen < args.CS {

			fmt.Println("missing value in log")
			done := make(chan *replicarpc.Reply)
			for _, storageServ := range rs.httpHandlerMap {
				args := &replicarpc.PrepareArgs{
					N:        args.N + 0.01,
					CS:       logLen,
					HostPort: rs.hostport,
				}
				var reply replicarpc.Reply
				call := storageServ.Go("ReplicaServer.Prepare", args, &reply, nil)
				go func() {
					seconds := 1
					duration := time.Duration(seconds) * time.Second
					select {
					case <-time.After(duration):
						done <- nil

					case <-call.Done:
						done <- (&reply)
					}
				}()
			}

			for i := 0; i < len(rs.httpHandlerMap); i++ {
				reply := <-done
				if (reply != nil) && (reply.OK) {
					if reply.Committed { // case of temporarily dead server uncommitted msg
						commitVal := &replicarpc.CommitVal{
							N:         reply.N,
							V:         reply.Value,
							CS:        reply.CS,
							Committed: true,
						}
						rs.logMutex.Lock()
						rs.commitLog[reply.CS] = commitVal
						rs.logMutex.Unlock()
						break
					}
				}
			}
			rs.logMutex.RLock()
			logLen = len(rs.commitLog)
			rs.logMutex.RUnlock()
		}
		rs.logMutex.Lock()
		rs.commitLog[args.CS] = args
		fmt.Println(rs.commitLog)
		rs.logMutex.Unlock()
	}

	rs.Va = nil
	rs.Na = 0.0

	return nil
}

func (rs *replicaServer) paxosRoutine(value []byte) {
	if !(rs.myN > rs.Nh) {
		rs.myN = rs.Nh + 0.01
	}
	rs.V = value
	rs.N = 0.0

	//prepare
	cs := len(rs.commitLog)
	done := make(chan *replicarpc.Reply)
	replies := 0
	var acceptedInit = true
	var PrepareOK = false
	var AcceptOK = false

	for i, storageServ := range rs.httpHandlerMap {
		fmt.Println("sending prepare to node:", i, "value:", rs.V)
		args := &replicarpc.PrepareArgs{
			N:        rs.myN,
			CS:       cs,
			HostPort: rs.hostport,
		}
		var reply replicarpc.Reply
		call := storageServ.Go("ReplicaServer.Prepare", args, &reply, nil)
		go func() {
			seconds := 1
			duration := time.Duration(seconds) * time.Second
			select {
			case <-time.After(duration):
				done <- nil

			case <-call.Done:
				fmt.Println("reply.OK =", (&reply).OK)
				done <- &reply
			}
		}()
	}

	for i := 0; i < len(rs.httpHandlerMap); i++ {
		reply := <-done
		if (reply != nil) && (reply.OK) {
			if reply.Committed { // case of temporarily dead server uncommitted msg
				commitVal := &replicarpc.CommitVal{
					N:         reply.N,
					V:         reply.Value,
					CS:        reply.CS,
					Committed: true,
				}
				rs.logMutex.Lock()
				rs.commitLog[reply.CS] = commitVal
				rs.logMutex.Unlock()
				rs.paxosRoutine(value) //restart paxos with value
				return
			}

			replies += 1

			if reply.Value != nil {
				if reply.N > rs.N {
					acceptedInit = false
					rs.V = reply.Value
					rs.N = reply.N
				}
			}
		}
	}
	fmt.Println("got prepare ok from", replies, "nodes")

	if (replies) >= (rs.numNodes+1)/2 {
		PrepareOK = true
	} else {
		acceptedInit = false
	}

	// Accept

	replies = 0

	if PrepareOK {
		for i, storageServ := range rs.httpHandlerMap {
			fmt.Println("sending accept to node:", i)
			args := &replicarpc.AcceptArgs{
				N:        rs.myN,
				V:        rs.V,
				CS:       cs,
				HostPort: rs.hostport,
			}
			var reply replicarpc.Reply
			call := storageServ.Go("ReplicaServer.Accept", args, &reply, nil)
			go func() {
				seconds := 1
				duration := time.Duration(seconds) * time.Second
				select {
				case <-time.After(duration):
					done <- nil

				case <-call.Done:
					done <- &reply
				}
			}()
		}

		for i := 0; i < len(rs.httpHandlerMap); i++ {
			reply := <-done

			if (reply != nil) && (reply.OK) {
				replies += 1
			}
		}

	}

	//commit
	fmt.Println("got accept ok from", replies, "nodes")

	if (replies) >= (rs.numNodes+1)/2 {
		AcceptOK = true
	} else {
		acceptedInit = false
	}

	/*
		if rs.testNodeDies_BC && (rs.FailureNode == rs.ID){
			ch := make(chan int)
			<- ch
		}
	*/

	if AcceptOK {
		for i, storageServ := range rs.httpHandlerMap {
			fmt.Println("sending commit to node:", i)
			args := &replicarpc.CommitVal{
				N:  rs.myN,
				V:  rs.V,
				CS: cs,
			}
			var reply replicarpc.Reply
			call := storageServ.Go("ReplicaServer.Commit", args, &reply, nil)
			go func() {
				seconds := 2
				duration := time.Duration(seconds) * time.Second
				select {
				case <-time.After(duration):
					done <- nil

				case <-call.Done:
					done <- &reply
				}
			}()
		}
		for i := 0; i < len(rs.httpHandlerMap); i++ {
			_ = <-done
		}

	}

	rs.Va = nil
	rs.Na = 0.0

	/*
		if rs.testNodeDies_AC && (rs.FailureNode == rs.ID){
			ch := make(chan int)
			<- ch
		}
	*/

	if !acceptedInit {
		time.Sleep(time.Duration(1) * time.Second)
		rs.paxosRoutine(value)
		return
	}

	return
}

func handleBuffer(in <-chan []byte, out chan<- []byte) {
	defer close(out)

	// This list will store all values received from 'in'.
	// All values should eventually be sent back through 'out',
	// even if the 'in' channel is suddenly closed.
	buffer := list.New()

	for {
		// Make sure that the list always has values before
		// we select over the two channels.
		if buffer.Len() == 0 {
			v, ok := <-in
			if !ok {
				// 'in' has been closed. Flush all values
				// in the buffer and return.
				flush(buffer, out)
				return
			}
			buffer.PushBack(v)
		}

		select {
		case v, ok := <-in:
			if !ok {
				// 'in' has been closed. Flush all values
				// in the buffer and return.
				flush(buffer, out)
				return
			}
			buffer.PushBack(v)
		case out <- (buffer.Front().Value).([]byte):
			buffer.Remove(buffer.Front())
		}
	}
}

// Blocks until all values in the buffer have been sent through
// the 'out' channel.
func flush(buffer *list.List, out chan<- []byte) {
	for e := buffer.Front(); e != nil; e = e.Next() {
		out <- (e.Value).([]byte)
	}
}
