package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	pb "github.com/CodersSquad/dc-labs/challenges/third-partial/proto"
	"github.com/nanomsg/mangos/protocol/respondent"
	"go.nanomsg.org/mangos"
	"google.golang.org/grpc"

	// register transports
	_ "go.nanomsg.org/mangos/transport/all"

	bolt "go.etcd.io/bbolt"
)

var url = "tcp://localhost:40899"

var (
	defaultRPCPort = 50051
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

var (
	controllerAddress = ""
	workerName        = ""
	tags              = ""
	nodeName          = ""
)

func deleteUser(name string) {
	db, er := bolt.Open("my.db", 0600, nil)
	if er != nil {
		log.Fatal(er)
	}
	defer db.Close()
	db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket([]byte(name))
		if err != nil {
			return fmt.Errorf("delete bucket: %s", err)
		}
		return nil
	})
}

func createUser(name string) {
	db, er := bolt.Open("my.db", 0600, nil)
	if er != nil {
		log.Fatal(er)
	}
	defer db.Close()
	db.Update(func(tx *bolt.Tx) error {
		if tx.Bucket([]byte(name)) == nil {
			b, _ := tx.CreateBucketIfNotExists([]byte(name))
			err := b.Put([]byte("status"), []byte("running"))
			return err
		}
		return nil
	})
}

func setWorkerTags(name string, tags string) {
	db, er := bolt.Open("my.db", 0600, nil)
	if er != nil {
		log.Fatal(er)
	}
	defer db.Close()
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(name))
		err := b.Put([]byte("tags"), []byte(tags))
		return err
	})
}

func setWorkerUsage(name string) {
	db, er := bolt.Open("my.db", 0600, nil)
	if er != nil {
		log.Fatal(er)
	}
	defer db.Close()
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(name))
		s := rand.NewSource(time.Now().UnixNano())
		r := rand.New(s)
		usage := r.Intn(100)
		err := b.Put([]byte("usage"), []byte(strconv.Itoa(usage)))
		return err
	})
}

func userExist(name string) bool {
	db, er := bolt.Open("my.db", 0600, nil)
	if er != nil {
		log.Fatal(er)
	}
	defer db.Close()
	exist := true
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(name))
		if b == nil {
			exist = false
		}
		return nil
	})
	return exist
}

func SetupCloseHandler(name string) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		deleteUser(name)
		os.Exit(0)
	}()
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("RPC: Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func init() {
	flag.StringVar(&controllerAddress, "controller", "tcp://localhost:40899", "Controller address")
	flag.StringVar(&nodeName, "node-name", "worker0", "Worker Name")
	flag.StringVar(&workerName, "worker-name", "hard-worker", "Worker Name")
	flag.StringVar(&tags, "tags", "gpu,superCPU,largeMemory", "Comma-separated worker tags")
}

func die(format string, v ...interface{}) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func date() string {
	return time.Now().Format(time.ANSIC)
}

// joinCluster is meant to join the controller message-passing server
func joinCluster(name string) {
	var sock mangos.Socket
	var err error
	var msg []byte

	if sock, err = respondent.NewSocket(); err != nil {
		die("can't get new respondent socket: %s", err.Error())
	}
	if err = sock.Dial("tcp://" + controllerAddress); err != nil {
		die("can't dial on respondent socket: %s", err.Error())
	}

	for {
		if msg, err = sock.Recv(); err != nil {
			die("Cannot recv: %s", err.Error())
		}
		fmt.Printf("CLIENT(%s): RECEIVED \"%s\" SURVEY REQUEST\n", name, string(msg))

		message := name + "|" + "|" + date()
		fmt.Printf("CLIENT(%s): SENDING DATE SURVEY RESPONSE\n", name)
		if err = sock.Send([]byte(message)); err != nil {
			die("Cannot send: %s", err.Error())
		}
	}
}

func getAvailablePort() int {
	port := defaultRPCPort
	for {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
		if err != nil {
			port = port + 1
			continue
		}
		ln.Close()
		break
	}
	return port
}

func main() {
	flag.Parse()
	SetupCloseHandler(nodeName)

	if userExist(nodeName) != true {
		createUser(nodeName)
		setWorkerTags(nodeName, tags)
		setWorkerUsage(nodeName)
	} else {
		fmt.Println("User already exit")
	}

	// Subscribe to Controller
	go joinCluster(nodeName)

	// Setup Worker RPC Server
	rpcPort := getAvailablePort()
	log.Printf("Starting RPC Service on localhost:%v", rpcPort)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", rpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

