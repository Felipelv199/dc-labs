package controller

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nanomsg/mangos/protocol/surveyor"
	"go.nanomsg.org/mangos"

	// register transports

	_ "go.nanomsg.org/mangos/transport/all"

	bolt "go.etcd.io/bbolt"
)

var controllerAddress = "tcp://localhost:40899"

func createWorker(name string) {
	db, er := bolt.Open("my.db", 0600, nil)
	if er != nil {
		log.Fatal(er)
	}
	defer db.Close()
	db.Update(func(tx *bolt.Tx) error {
		if tx.Bucket([]byte(name)) == nil {
			_, err := tx.CreateBucketIfNotExists([]byte(name))
			return err
		}
		return nil
	})
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

func die(format string, v ...interface{}) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func date() string {
	return time.Now().Format(time.ANSIC)
}

func Start() {
	var sock mangos.Socket
	var err error
	var msg []byte
	if sock, err = surveyor.NewSocket(); err != nil {
		die("can't get new surveyor socket: %s", err)
	}
	if err = sock.Listen(controllerAddress); err != nil {
		die("can't listen on surveyor socket: %s", err.Error())
	}
	err = sock.SetOption(mangos.OptionSurveyTime, time.Second/2)
	if err != nil {
		die("SetOption(): %s", err.Error())
	}
	for {
		time.Sleep(time.Second)
		fmt.Println("SERVER: SENDING DATE SURVEY REQUEST")
		if err = sock.Send([]byte("DATE")); err != nil {
			die("Failed sending survey: %s", err.Error())
		}
		for {
			if msg, err = sock.Recv(); err != nil {
				break
			}
			msgSplitted := strings.Split(string(msg), "|")
			if len(msgSplitted) > 2 {
				createWorker(msgSplitted[0])
				setWorkerValue(msgSplitted[0], "status", msgSplitted[1])
				setWorkerValue(msgSplitted[0], "tags", msgSplitted[2])
				setWorkerValue(msgSplitted[0], "port", msgSplitted[3])
				s := rand.NewSource(time.Now().UnixNano())
				r := rand.New(s)
				usage := r.Intn(100)
				setWorkerValue(msgSplitted[0], "usage", strconv.Itoa(usage))
				fmt.Printf("SERVER: RECEIVED CLIENT(%s) MESSAGE: WORKER CREATED SURVEY RESPONSE\n",
					msgSplitted[0])
			} else {
				fmt.Printf("SERVER: RECEIVED CLIENT(%s) MESSAGE: \"%s\" SURVEY RESPONSE\n",
					msgSplitted[0], msgSplitted[1])
			}

		}
		fmt.Println("SERVER: SURVEY OVER")
	}
}

