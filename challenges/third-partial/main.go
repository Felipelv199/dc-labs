package main

import (
	"fmt"
	"log"

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

func createJob() {
	db, er := bolt.Open("my.db", 0600, nil)
	if er != nil {
		log.Fatal(er)
	}
	defer db.Close()
	db.Update(func(tx *bolt.Tx) error {
		if tx.Bucket([]byte("Job")) == nil {
			_, err := tx.CreateBucketIfNotExists([]byte("Job"))
			return err
		}
		return nil
	})
}

func main() {
	createJob()
	jobs := make(chan string)
	jobsName := make(chan string)
	log.Println("Welcome to the Distributed and Parallel Image Processing System")
	// Start Controller
	go controller.Start()
	// Start Scheduler

	go scheduler.Start(jobs)
	//Start Api
	go api.Start(jobsName)
	for {
		name := <-jobsName
		jobs <- name
	}
}

