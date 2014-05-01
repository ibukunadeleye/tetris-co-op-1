package replicaserver

import (
	"fmt"
	"net/rpc"
	"rpc/replicarpc"
	"time"
	"sync"
	"strconv"
	"net"
	"net/http"
	//"encoding/json"
)

type starterServer struct{
	replicas map[int] *rpc.Client
	replicaHostPort map[int] string
	replicaMutex *sync.RWMutex
	currentAcceptor int
	numNodes int
	masterHostPort string
	ready bool
	readyChan chan int
}

func NewStarterServer (masterServerHostPort string, port, numNodes int) (StarterServer,error) {
	fmt.Println("NewStarter started on:",port)
	p := strconv.Itoa(port)
	ss :=& starterServer{
		numNodes : numNodes,
		replicas : make(map [int] *rpc.Client),
		replicaHostPort : make(map[int] string),
		masterHostPort : ("localhost:" + p),
		replicaMutex : new(sync.RWMutex),
		ready : false,
		readyChan : make(chan int),
	}
	masterListener, err := net.Listen("tcp", ss.masterHostPort)
	if err != nil {
		fmt.Println("net.Listen returned error")
		return nil,err
	}
	rpc.RegisterName("StarterServer", ss)
	rpc.HandleHTTP()
	go http.Serve(masterListener, nil)
	go ss.waitReady()



	return ss,nil
}



func (ss *starterServer) waitReady(){
	select{
		case <-ss.readyChan:
			ss.ready = true
			for _,replicaRPC:= range ss.replicas{
				args := & replicarpc.AddServersArgs{
					ServerMap : ss.replicaHostPort,
				}
				reply := new(replicarpc.AddServersReply)
				fmt.Println("rpc addServers called")
				err := replicaRPC.Call("ReplicaServer.AddServers",args,reply)
				if err != nil{
					fmt.Println("got error from call AddServers:", err)
				}
			}
	}
}

func (ss *starterServer) RegisterReplica(args *replicarpc.RegisterArgs,reply *replicarpc.RegisterReply)error{
	hostport := args.Hostport
	fmt.Println("registering Server:",hostport)
	replicaRPC,err := rpc.DialHTTP("tcp",hostport)
	ss.replicaMutex.Lock()
	nodeID := len(ss.replicas)
	if (nodeID < ss.numNodes){
		ss.replicas[nodeID] = replicaRPC
		ss.replicaHostPort[nodeID] = hostport
	}
	ss.replicaMutex.Unlock()
	
	fmt.Println("regisering nodeID:",nodeID)
	reply.NodeID = nodeID
	
	ss.replicaMutex.RLock()
	if (ss.numNodes == len(ss.replicas)){
		ss.readyChan <- 1
	}
	ss.replicaMutex.RUnlock()
	fmt.Println("registered Server:",hostport)
	return err
}


func (ss *starterServer) Get(args *replicarpc.GetArgs, reply *replicarpc.GetReply) error {
	replicaRPC := ss.replicas[ss.currentAcceptor]
	call := replicaRPC.Go("ReplicaServer.Get", args, reply,nil)
	func(){
		seconds := 3
		duration := time.Duration(seconds)*time.Second
		select{
			case <-time.After(duration):
				ss.currentAcceptor =(ss.currentAcceptor+1) % (len(ss.replicas))
				fmt.Println("node failure switching nodes")
				ss.Get(args,reply)
				
			case <-call.Done:
				fmt.Println("(in ss) got back:",reply.V)
		}
	}()
	return nil
}

func (ss *starterServer) Put(args *replicarpc.PutArgs, reply *replicarpc.PutReply) error {
	replicaRPC := ss.replicas[ss.currentAcceptor]
	call := replicaRPC.Go("ReplicaServer.Put", args, reply,nil)
	go func(){
		seconds := 2
		duration := time.Duration(seconds)*time.Second
		select{
			case <-time.After(duration):
				if call.Error !=nil {
					fmt.Println("1got error from call.Go err:",call.Error)
				}
				ss.currentAcceptor =(ss.currentAcceptor+1) % (len(ss.replicas))
				fmt.Println("node failure switching nodes")
				ss.Put(args,reply)
				
			case <-call.Done:
				fmt.Println("put was completed")
				if call.Error !=nil {
					fmt.Println("2got error from call.Go err:",call.Error)
					ss.currentAcceptor =(ss.currentAcceptor+1) % (len(ss.replicas))
					fmt.Println("node failure switching nodes")
					ss.Put(args,reply)
				}
			
		}
	}()
	return nil
}


