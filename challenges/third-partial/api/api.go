package api

import (
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	bolt "go.etcd.io/bbolt"
)

var activeUsers = make(map[string]string)
var jobName = make(chan string)

func getJobWorker(task string) string {
	value := ""
	db, er := bolt.Open("my.db", 0600, nil)
	if er != nil {
		log.Fatal(er)
	}
	defer db.Close()
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Job"))
		if b != nil {
			for {
				v := b.Get([]byte(task))
				value = string(v)
				break
			}
		}
		return nil
	})
	return value
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

func workerExist(name string) bool {
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

func loginVerification(user string) bool {
	for _, value := range activeUsers {
		if user == value {
			return false
		}
	}
	return true
}

func Login(c *gin.Context) {

	user := c.MustGet(gin.AuthUserKey).(string)
	temp := loginVerification(user)

	if !temp {
		c.JSON(http.StatusOK, gin.H{
			"message": "You already in " + user,
		})
	} else {
		token := (xid.New()).String()
		activeUsers[token] = user

		c.JSON(http.StatusOK, gin.H{
			"message": "Hi " + user + " welcome to the DPIP System",
			"token":   token,
		})
	}

}

func Logout(c *gin.Context) {
	token := strings.Trim(strings.TrimLeft(c.GetHeader("authorization"), "Bearer"), " ")

	for key, value := range activeUsers {
		if token == key {
			user := value
			delete(activeUsers, key)
			c.JSON(http.StatusOK, gin.H{
				"message": "Bye " + user + ", your token has been revoked",
			})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Token " + token + " does not exist",
	})
}

func Status(c *gin.Context) {
	token := strings.Trim(strings.TrimLeft(c.GetHeader("authorization"), "Bearer"), " ")
	for key, value := range activeUsers {
		if token == key {
			user := value
			activeWorkers := getWorkers()
			c.JSON(http.StatusOK, gin.H{
				"message": "Hi " + user + ", the DPIP System is Up and Running",
				"time":    time.Now().UTC().String(),
			})

			for _, worker := range activeWorkers {

				c.JSON(http.StatusOK, gin.H{
					"Worker": worker,
					"Status": getWorkerValue(worker, "status"),
					"Tags":   getWorkerValue(worker, "tags"),
					"Usage":  getWorkerValue(worker, "usage") + "%",
				})
			}
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Token " + token + " does not exist",
	})
}

func StatusParam(c *gin.Context) {
	token := strings.Trim(strings.TrimLeft(c.GetHeader("authorization"), "Bearer"), " ")
	for key, _ := range activeUsers {
		if token == key {
			worker := c.Param("worker")
			if workerExist(worker) == true {
				c.JSON(http.StatusOK, gin.H{
					"Worker": worker,
					"Status": getWorkerValue(worker, "status"),
					"Tags":   getWorkerValue(worker, "tags"),
					"Usage":  getWorkerValue(worker, "usage") + "%",
				})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"message": "Worker doesn't exist",
				})
			}
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Token " + token + " does not exist",
	})
}

func Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("data")
	if err != nil {
		log.Fatal(err)
	}
	token := strings.Trim(strings.TrimLeft(c.GetHeader("authorization"), "Bearer"), " ")

	for key := range activeUsers {
		if token == key {
			fileName := header.Filename
			fileSize := header.Size

			out, err := os.Create("C:\\Users\\quint\\AppData\\Local\\Temp\\" + fileName)
			if err != nil {
				log.Fatal(err)
			}
			defer out.Close()
			_, err = io.Copy(out, file)
			if err != nil {
				log.Fatal(err)
			}
			fileSize = fileSize / 1000
			str := strconv.FormatInt(fileSize, 10)
			c.JSON(http.StatusOK, gin.H{
				"message":  "An image has been successfully uploaded",
				"filename": fileName,
				"size":     str + "kb",
			})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Token " + token + " does not exist",
	})
}

func WorkloadsTest(c *gin.Context) {
	token := strings.Trim(strings.TrimLeft(c.GetHeader("authorization"), "Bearer"), " ")
	for key, _ := range activeUsers {
		if token == key {
			name := c.Param("test")
			workers := getWorkers()
			if len(workers) > 0 {
				jobName <- name
				worker := getJobWorker(name)
				c.JSON(http.StatusOK, gin.H{
					"Workload": name,
					"Job ID":   "1",
					"Status":   "Scheduling",
					"Result":   "Done in Worker: " + worker,
				})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"message": "There're any worker active",
				})
			}
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Token " + token + " does not exist",
	})
}

func Start(jobs chan string) {
	r := gin.Default()
	jobName = jobs
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"user1": "pass1",
		"user2": "pass2",
		"user3": "pass3",
	}))

	authorized.GET("/login", Login)
	r.GET("/logout", Logout)
	r.GET("/status", Status)
	r.GET("/status/:worker", StatusParam)
	r.GET("/workloads/:test", WorkloadsTest)
	r.POST("/upload", Upload)

	r.Run() // listen and serve on 0.0.0.0:8080

}

