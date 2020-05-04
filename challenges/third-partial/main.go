package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/CodersSquad/dc-labs/challenges/third-partial/api"
	"github.com/CodersSquad/dc-labs/challenges/third-partial/controller"
	"github.com/CodersSquad/dc-labs/challenges/third-partial/scheduler"
	bolt "go.etcd.io/bbolt"
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
			workers = append(workers, string(k))
		}
		return nil
	})
	return workers
}
func main() {
	log.Println("Welcome to the Distributed and Parallel Image Processing System")
	// Start Controller
	go controller.Start()
	// Start Scheduler
	jobs := make(chan scheduler.Job)
	go scheduler.Start(jobs)
	// Send sample jobs
	sampleJob := scheduler.Job{Address: "localhost:50051", RPCName: "hello"}

	sampleJob.RPCName = fmt.Sprintf("hello-%v", rand.Intn(10000))
	jobs <- sampleJob
	time.Sleep(time.Second * 5)
	//Start Api
	api.Start()
}

