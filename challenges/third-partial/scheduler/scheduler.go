package scheduler

import (
	"context"
	"log"
	"math/rand"
	"strconv"
	"time"

	pb "github.com/CodersSquad/dc-labs/challenges/third-partial/proto"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/grpc"
)

var workerName = make(chan string)

type Job struct {
	Address string
	RPCName string
}

func setJobWorker(job string, worker string) {
	db, er := bolt.Open("my.db", 0600, nil)
	if er != nil {
		log.Fatal(er)
	}
	defer db.Close()
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Job"))
		err := b.Put([]byte(job), []byte(worker))
		return err
	})
}

func getJobWorker(name string, key string) string {
	value := ""
	db, er := bolt.Open("my.db", 0600, nil)
	if er != nil {
		log.Fatal(er)
	}
	defer db.Close()
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(name))
		if b != nil {
			v := b.Get([]byte(key))
			value = string(v)
		}
		return nil
	})
	return value
}

func getWorkers() []string {
	db, er := bolt.Open("my.db", 0600, nil)
	if er != nil {
		log.Fatal(er)
	}
	defer db.Close()
	var workers []string
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		c := tx.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			if string(k) != "Job" {
				workers = append(workers, string(k))
			}
		}
		return nil
	})
	return workers
}

func getWorkerValue(name string, key string) string {
	value := ""
	db, er := bolt.Open("my.db", 0600, nil)
	if er != nil {
		log.Fatal(er)
	}
	defer db.Close()
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(name))
		if b != nil {
			v := b.Get([]byte(key))
			value = string(v)
		}
		return nil
	})
	return value
}

func setWorkerValue(name string, key string, value string) {
	db, er := bolt.Open("my.db", 0600, nil)
	if er != nil {
		log.Fatal(er)
	}
	defer db.Close()
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(name))
		err := b.Put([]byte(key), []byte(value))
		return err
	})
}

func schedule(task string) {

	var job Job
	worker := ""
	if len(getWorkers()) != 0 {
		worker = getWorkers()[0]
	}
	for _, k := range getWorkers() {
		if getWorkerValue(worker, "usage") > getWorkerValue(k, "usage") {
			worker = k
		}
	}
	setJobWorker(task, worker)
	job.Address = "localhost:" + getWorkerValue(worker, "port")
	job.RPCName = worker
	//Set up a connection to the server.
	conn, err := grpc.Dial(job.Address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: job.RPCName})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Scheduler: RPC respose from %s : %s", job.Address, r.GetMessage())
	s := rand.NewSource(time.Now().UnixNano())
	ra := rand.New(s)
	usage := ra.Intn(100)
	setWorkerValue(worker, "usage", strconv.Itoa(usage))
}

func Start(jobs chan string) error {
	for {
		job := <-jobs
		schedule(job)
	}
	return nil
}

