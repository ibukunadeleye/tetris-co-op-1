package replicaserver

import (
	"errors"
	"net/http"
	"net/rpc"
	"rpc/replicarpc"
	"time"
	"strconv"
	"net"
	"sync"
)


type replicaServer struct {
	hostport  string
	ID        uint32
	numNodes  int
	myN		  float64
	Na		  float64
	Va		  []byte
	V		  []byte	//current V bieng sent
	N         float64	//current N being sent
	Nh		  float64
	commitLog map [int] *replicarpc.CommitVal
	replyChan chan replicarpc.Reply
	replicaList []string
	httpHandlerMap map[int]*rpc.Client
	logMutex *sync.RWMutex
}

// NewReplicaServer creates and starts a new ReplicaServer.
// numNodes is the number of members in the Paxos quorum.
// port is the number of the port that the replicaserver should listen on.
// memberID is the ID the replicaserver should have
// For now, the address of all replicaservers are "localhost".
// This function should return only once all members have joined the quorum
// and should return a non-nil error if the replicaserver could not be started.
func NewReplicaServer(numNodes, port int, memberID uint32) (ReplicaServer, string, error) {
	rs := &replicaServer{
		hostport : ("localhost:" + strconv.Itoa(port)),
		ID : memberID,
		myN :0.0,
		Na : 0.0,
		Va : nil,
		V  : nil,
		N  : 0.0, 
		Nh : 0.0,
		logMutex : new(sync.RWMutex),
		commitLog : make(map [int] *replicarpc.CommitVal),
	}
	
	replicaListener, err := net.Listen("tcp", rs.hostport)
	if err != nil {
		return nil,"",err
	}
	rpc.RegisterName("ReplicaServer", rs)
	rpc.HandleHTTP()
	go http.Serve(replicaListener, nil)
	
	return rs, rs.hostport, nil
}

func (rs *replicaServer) Get(args *replicarpc.GetArgs, reply *replicarpc.GetReply) error {
	rs.logMutex.RLock()
	commitval,ok := rs.commitLog[len(rs.commitLog)-1]
	if ! ok{
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
	rs.PaxosRoutine(args.V)
	return nil
}

func (rs *replicaServer) AppendToList(args *replicarpc.PutArgs, reply *replicarpc.PutReply) error {
	return errors.New("not implemented")
}

func (rs *replicaServer) RemoveFromList(args *replicarpc.PutArgs, reply *replicarpc.PutReply) error {
	return errors.New("not implemented")
}

func (rs *replicaServer) AddServers(args *replicarpc.AddServersArgs, reply *replicarpc.AddServersReply){
	rs.httpHandlerMap = args.ServerMap
	rs.numNodes = len(args.ServerMap)
}


func (rs *replicaServer) Prepare(args *replicarpc.PrepareArgs, reply *replicarpc.Reply) error{
	N := args.N 
	CS := args.CS
	
	rs.logMutex.RLock()
	commitval,ok := rs.commitLog[CS]
	rs.logMutex.RUnlock()
	
	if ok{	//commit slot taken
		reply.OK = true
		reply.Committed = true
		reply.Value = commitval.V
		reply.N = commitval.N
		reply.CS = CS
		return nil
	}else{
		if rs.Nh > N{			//higher N seen
			reply.OK = false
			return nil
		}else{					//higher N not seen
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

func (rs *replicaServer) Accept(args *replicarpc.AcceptArgs, reply *replicarpc.Reply) error{
	N := args.N
	V := args.V

	if rs.Nh > N{			//higher N seen
		reply.OK = false
	}else{					//higher N not seen
		reply.OK = true
		rs.Va = V
		rs.Na = N
		rs.Nh = N
	}
	
	return nil
}

func (rs *replicaServer) Commit(args *replicarpc.CommitVal, reply *replicarpc.Reply){
	rs.logMutex.RLock()
	logLen := len(rs.commitLog)
	rs.logMutex.RUnlock()
	
	for ( logLen < args.CS){
		done := make(chan *replicarpc.Reply)
		for _,storageServ := range rs.httpHandlerMap{
			args := &replicarpc.PrepareArgs{
				N : args.N,
				CS : logLen,
				HostPort : rs.hostport,
			}
			var reply *replicarpc.Reply
			call := storageServ.Go("ReplicaServer.Prepare", args, reply,nil)
			go func(){
				seconds := 1 
				duration := time.Duration(seconds)*time.Second
				select{
					case <-time.After(duration):
						done <- nil
						
					case <-call.Done:
						done <- reply
				}
			}()
		}
		
		for i:=0;i<len(rs.httpHandlerMap);i++{
			reply := <-done
			if (reply != nil)&&(reply.OK){
				if reply.Committed{					// case of temporarily dead server uncommitted msg
					commitVal := &replicarpc.CommitVal{
						N : reply.N,
						V : reply.Value,
						CS: reply.CS,
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
	rs.logMutex.Unlock()
	
	rs.Va = nil
	rs.Na = 0.0
		
	return 
}


func (rs *replicaServer) PaxosRoutine(value []byte){
	if !(rs.myN > rs.Nh){
		rs.myN = rs.Nh + 0.01
	}
	rs.V = value
	rs.N = 0.0
		
	//prepare
	cs := len(rs.commitLog)
	done := make(chan *replicarpc.Reply)
	replies :=0
	var acceptedInit = true
	var PrepareOK = false
	var AcceptOK = false
	
	for _,storageServ := range rs.httpHandlerMap{
		args := &replicarpc.PrepareArgs{
			N : rs.myN,
			CS :cs,
			HostPort : rs.hostport,
		}
		var reply *replicarpc.Reply
		call := storageServ.Go("ReplicaServer.Prepare", args, reply,nil)
		go func(){
			seconds := 1 
			duration := time.Duration(seconds)*time.Second
			select{
				case <-time.After(duration):
					done <- nil
					
				case <-call.Done:
					done <- reply
			}
		}()
	}
	
	for i:=0;i<len(rs.httpHandlerMap);i++{
		reply := <-done
		if (reply != nil)&&(reply.OK){
			if reply.Committed{					// case of temporarily dead server uncommitted msg
				commitVal := &replicarpc.CommitVal{
					N : reply.N,
					V : reply.Value,
					CS: reply.CS,
					Committed: true,
				}
				rs.logMutex.Lock()
				rs.commitLog[reply.CS] = commitVal
				rs.logMutex.Unlock()
				rs.PaxosRoutine(value)				//restart paxos with value
				return
			}
			
			replies += 1
			
			
			if reply.Value !=nil{
				if reply.N > rs.N{
					acceptedInit = false
					rs.V = reply.Value
					rs.N = reply.N
				}
			}
		}
	}
	
	if (replies) >= (rs.numNodes + 1)/2{  
		PrepareOK = true
	}else{
		acceptedInit = false
	}
	
	
	// Accept
	
	
	
	replies = 0
	
	if PrepareOK {
		for _,storageServ := range rs.httpHandlerMap{
			args := &replicarpc.AcceptArgs{
				N : rs.myN,
				V : rs.V,
				CS : cs,	
				HostPort : rs.hostport,
			}
			var reply *replicarpc.Reply
			call := storageServ.Go("ReplicaServer.Accept",args,reply,nil)
			go func(){
				seconds := 1 
				duration := time.Duration(seconds)*time.Second
				select{
					case <-time.After(duration):
						done <- nil
						
					case <-call.Done:
						done <- reply
				}
			}()
		}
		
		for i:=0;i<len(rs.httpHandlerMap);i++{
			reply := <-done
			
			if (reply != nil) && (reply.OK){
				replies += 1
			}
		}
		
	}
	
	//commit
	
	if (replies) >= (rs.numNodes + 1)/2{
		AcceptOK = true
	}else{
		acceptedInit = false
	}
	
	if AcceptOK {
		for _,storageServ := range rs.httpHandlerMap{
			args := &replicarpc.CommitVal{
				N : rs.myN,
				V : rs.V,
				CS : cs,
			}
			var reply *replicarpc.Reply
			storageServ.Go("ReplicaServer.Commit",args,reply,nil)
		}
	}

	rs.Va = nil
	rs.Na = 0.0
	
	if ! acceptedInit{
		time.Sleep(time.Duration(1)*time.Second)
		rs.PaxosRoutine(value)
		return
	}
	
	return
}
