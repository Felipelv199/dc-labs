package api

import (
	"fmt"
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

func userStatus(name string) string {
	status := ""
	db, er := bolt.Open("my.db", 0600, nil)
	if er != nil {
		log.Fatal(er)
	}
	defer db.Close()
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(name))
		if b != nil {
			v := b.Get([]byte("status"))
			status = string(v)
		}
		return nil
	})
	return status
}

func userTags(name string) string {
	status := ""
	db, er := bolt.Open("my.db", 0600, nil)
	if er != nil {
		log.Fatal(er)
	}
	defer db.Close()
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(name))
		if b != nil {
			v := b.Get([]byte("tags"))
			status = string(v)
		}
		return nil
	})
	return status
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

func getWorkerUsage(name string) int {
	db, er := bolt.Open("my.db", 0600, nil)
	if er != nil {
		log.Fatal(er)
	}
	defer db.Close()
	usage := 0
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(name))
		if b != nil {
			v := b.Get([]byte("usage"))
			i, _ := strconv.Atoi(string(v))
			usage = i
		}
		return nil
	})
	return usage
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
					"Status": userStatus(worker),
					"Tags":   userTags(worker),
					"Usage":  strconv.Itoa(getWorkerUsage(worker)) + "%",
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
			if userExist(worker) == true {
				c.JSON(http.StatusOK, gin.H{
					"Worker": worker,
					"Status": userStatus(worker),
					"Tags":   userTags(worker),
					"Usage":  strconv.Itoa(getWorkerUsage(worker)) + "%",
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
			workers := getWorkers()
			if len(workers) > 0 {
				worker := workers[0]
				for _, v := range workers {
					if worker != v {
						if getWorkerUsage(worker) > getWorkerUsage(v) {
							worker = v
						}
					}
				}
				c.JSON(http.StatusOK, gin.H{
					"Workload": "test",
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

func Start() {
	fmt.Println("Entre")
	r := gin.Default()
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"user1": "pass1",
		"user2": "pass2",
		"user3": "pass3",
	}))

	authorized.GET("/login", Login)
	r.GET("/logout", Logout)
	r.GET("/status", Status)
	r.GET("/status/:worker", StatusParam)
	r.GET("/workloads/test", WorkloadsTest)
	r.POST("/upload", Upload)

	r.Run() // listen and serve on 0.0.0.0:8080

}

