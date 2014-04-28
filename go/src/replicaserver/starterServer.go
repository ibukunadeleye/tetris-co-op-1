package replicaserver

import (
	"fmt"
	"math/rand"
	"net/rpc"
	"rpc/replicarpc"
	"time"
	//"encoding/json"
)

type starterServer struct{
	replicas map[int] *rpc.Client
	currentAcceptor int
	numNodes int
}

func NewStarterServer (numNodes int) error {
	
	ss :=& starterServer{
		numNodes : numNodes,
		replicas : make(map [int] *rpc.Client),
	}
	
	//providing random ports to Nodes
	for i:=0; i<numNodes; i++{
		replicaPort := (rand.Int() %1000) +1000
		_,replicaHostPort,err := NewReplicaServer(numNodes,replicaPort,uint32(i))
		replicaRPC,err := rpc.DialHTTP("tcp",replicaHostPort)
		if err !=nil{
			return err
		}
		ss.replicas [i] = replicaRPC
	}
	ss.currentAcceptor = 0
	
	for i:=0; i<numNodes; i++{
		replicaRPC,_ := ss.replicas[i]
		args := & replicarpc.AddServersArgs{
			ServerMap : ss.replicas,
		}
		reply := new(replicarpc.AddServersReply)
		replicaRPC.Call("ReplicaServer.AddServers",args,reply)
	}
	return nil
}


func (ss *starterServer) Get(args *replicarpc.GetArgs, reply *replicarpc.GetReply) error {
	
	call := ss.replicas[ss.currentAcceptor].Go("ReplicaServer.Get", args, reply,nil)
	go func(){
		seconds := 1
		duration := time.Duration(seconds)*time.Second
		select{
			case <-time.After(duration):
				ss.currentAcceptor =(ss.currentAcceptor+1) % (len(ss.replicas))
				fmt.Println("node failure switching nodes")
				ss.Get(args,reply)
				
			case <-call.Done:
		}
	}()
	return nil
}

func (ss *starterServer) Put(args *replicarpc.PutArgs, reply *replicarpc.PutReply) error {
	
	call := ss.replicas[ss.currentAcceptor].Go("ReplicaServer.Put", args, reply,nil)
	go func(){
		seconds := 3
		duration := time.Duration(seconds)*time.Second
		select{
			case <-time.After(duration):
				ss.currentAcceptor =(ss.currentAcceptor+1) % (len(ss.replicas))
				fmt.Println("node failure switching nodes")
				ss.Put(args,reply)
				
			case <-call.Done:
			
		}
	}()
	return nil
}