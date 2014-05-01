package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"regexp"
	"time"

	"rpc/replicarpc"
)

type storageTester struct {
	srv        *rpc.Client
	myhostport string
	recvRevoke map[string]bool // whether we have received a RevokeLease for key x
	compRevoke map[string]bool // whether we have replied the RevokeLease for key x
	delay      float32         // how long to delay the reply of RevokeLease
}

type testFunc struct {
	name string
	f    func()
}


var (
	portnum   = flag.Int("port", 9019, "port # to listen on")
	masterPort= flag.Int("mport",9009, "port # to talk to")
	numServer = flag.Int("N", 1, "(jtest only) total # of storage servers")
	testRegex = flag.String("t", "", "test to run")
	passCount int
	failCount int
	st        *storageTester
)

var LOGE = log.New(os.Stderr, "", log.Lshortfile|log.Lmicroseconds)

func initStorageTester(server, myhostport string) (*storageTester, error) {
	tester := new(storageTester)
	tester.myhostport = myhostport
	tester.recvRevoke = make(map[string]bool)
	tester.compRevoke = make(map[string]bool)

	// Create RPC connection to storage server.
	srv, err := rpc.DialHTTP("tcp", server)
	if err != nil {
		return nil, fmt.Errorf("could not connect to server %s", server)
	}

	rpc.Register(tester)
	rpc.HandleHTTP()

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", *portnum))
	if err != nil {
		LOGE.Fatalln("Failed to listen:", err)
	}
	go http.Serve(l, nil)
	tester.srv = srv
	return tester, nil
}

func (st *storageTester) Get() (*replicarpc.GetReply, error) {
	args := &replicarpc.GetArgs{}
	var reply replicarpc.GetReply
	err := st.srv.Call("StarterServer.Get", args, &reply)
	return &reply, err
}

func (st *storageTester) Put(value []byte) (*replicarpc.PutReply, error) {
	args := &replicarpc.PutArgs{V: value}
	var reply replicarpc.PutReply
	err := st.srv.Call("StarterServer.Put", args, &reply)
	return &reply, err
}

func testPut3Get() {

	_, err := st.Put([]byte{1,2,3})
	if err != nil{
		fmt.Println(err)
	}
	
	//time.Sleep(time.Duration(10)*time.Second)
	
	_, err = st.Put([]byte{4,5,6})
	if err != nil{
		fmt.Println(err)
	}
	_, err = st.Put([]byte{7,8,9})
	if err != nil{
		fmt.Println(err)
	}
/*	//time.Sleep(time.Duration(2)*time.Second)
	replyG, err := st.Get()
	
	fmt.Println("expect value: [7,8,9]")
	fmt.Println("got back:", replyG.V,"from get method")
*/
	fmt.Println("PASS")
	passCount++
}

func testPut10Get3() {
	_, err := st.Put([]byte{1,2,3})					//PUT
	if err != nil{
		fmt.Println(err)
	}
	
	_, err = st.Put([]byte{4,5,6})					//PUT
	if err != nil{
		fmt.Println(err)
	}
	
	_, err = st.Put([]byte{7,8,9})					//PUT
	if err != nil{
		fmt.Println(err)
	}
	
	//time.Sleep(time.Duration(10)*time.Second)		//SLEEP
	
	_, err = st.Put([]byte{10,11,12,13})			//PUT
	if err != nil{
		fmt.Println(err)
	}
	

/*	//time.Sleep(time.Duration(2)*time.Second)
	replyG, err := st.Get()							//GET
	
	fmt.Println("expect value: [10,11,12,13]")
	fmt.Println("got back:", replyG.V,"from get method")
*/	
	_, err = st.Put([]byte{14,15,16})				//PUT
	if err != nil{
		fmt.Println(err)
	}
	
	_, err = st.Put([]byte{17})						//PUT
	if err != nil{
		fmt.Println(err)
	}
	
	_, err = st.Put([]byte{18,19,20,22,23,24,25})	//PUT
	if err != nil{
		fmt.Println(err)
	}
	
/*	//time.Sleep(time.Duration(2)*time.Second)		//GET
	replyG, err = st.Get()	
	
	fmt.Println("expect value: [18,19,20,22,23,24,25]")
	fmt.Println("got back:", replyG.V,"from get method")
*/	
	_, err = st.Put([]byte{26,27})				//PUT
	if err != nil{
		fmt.Println(err)
	}
	
	_, err = st.Put([]byte{28,29,30,31})				//PUT
	if err != nil{
		fmt.Println(err)
	}
	
	_, err = st.Put([]byte{32,33,34,35})				//PUT
	if err != nil{
		fmt.Println(err)
	}
	
	for i:=0; i<5; i++{
		time.Sleep(time.Duration(100)*time.Millisecond)
		replyG, _ := st.Get()
		
		fmt.Println("got back:", replyG.V,"from get method")
	}
	
//	fmt.Println("expect value: [32,33,34,35]")
//	fmt.Println("got back:", replyG.V,"from get method")

	fmt.Println("PASS")
	passCount++	
}

func main() {
	jtests := []testFunc{{"testPutGet", testPut3Get},
						 {"testPut10Get3", testPut10Get3},}
	
	flag.Parse()
	
	// Run the tests with a single tester
	storageTester, err := initStorageTester(fmt.Sprintf("localhost:%d", *masterPort), fmt.Sprintf("localhost:%d", *portnum))
	if err != nil {
		LOGE.Fatalln("Failed to initialize test:", err)
	}
	st = storageTester
	
	for _, t := range jtests {
		if b, err := regexp.MatchString(*testRegex, t.name); b && err == nil {
			fmt.Printf("Running %s:\n", t.name)
			t.f()
		}
	}
	
	fmt.Printf("Passed (%d/%d) tests\n", passCount, passCount+failCount)
}


