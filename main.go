package main

import (
	"fmt"
	"github.com/ungerik/go-dry"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"time"
)

var actions = []string{"logged in", "logged out", "created record", "deleted record", "updated account"}

var startTime time.Time

type logItem struct {
	action    string
	timestamp time.Time
}

type User struct {
	id    int
	email string
	logs  []logItem
}

func (u User) getActivityInfo() string {
	output := fmt.Sprintf("UID: %d; Email: %s;\nActivity Log:\n", u.id, u.email)
	for index, item := range u.logs {
		output += fmt.Sprintf("%d. [%s] at %s\n", index, item.action, item.timestamp.Format(time.RFC3339))
	}

	return output
}

func init() {
	prepareDB()
	rand.Seed(time.Now().Unix())
	startTime = time.Now()
}

func main() {
	const resultsCount, workersCount = 100, 100

	wg := &sync.WaitGroup{}

	jobs := make(chan int, resultsCount)
	users := make(chan User, resultsCount)

	// generate users with the help of goroutines
	generateUsers(workersCount, jobs, users)

	// generate jobs for users
	generateJobs(resultsCount, jobs, wg)

	// wait for all users to be generated and save their info
	saveUsersInfo(workersCount, users, wg)

	wg.Wait()

	fmt.Printf("DONE! Time Elapsed: %.2f seconds\n", time.Since(startTime).Seconds())
}

func generateJobs(count int, jobs chan<- int, wg *sync.WaitGroup) {
	for i := 0; i < count; i++ {
		wg.Add(1)
		jobs <- i
	}
}

func generateUsers(workersCount int, jobs <-chan int, users chan<- User) {
	for i := 0; i < workersCount; i++ {
		go func() {
			for j := range jobs {
				users <- User{
					id:    j,
					email: fmt.Sprintf("user%d@company.com", j),
					logs:  generateLogs(rand.Intn(1000)),
				}
				fmt.Printf("generated user %d\n", j)
				time.Sleep(time.Millisecond * 100)
			}
			close(users)
		}()
	}
}

func saveUsersInfo(workersCount int, users chan User, wg *sync.WaitGroup) {
	for i := 0; i < workersCount; i++ {
		go func() {
			for user := range users {
				fmt.Printf("WRITING FILE FOR UID %d\n", user.id)

				// create file
				filename := fmt.Sprintf("users/uid%d.txt", user.id)
				file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
				dry.PanicIfErr(err)

				_, err = file.WriteString(user.getActivityInfo())
				dry.PanicIfErr(err)

				time.Sleep(time.Second)

				// wait for all users to be saved
				wg.Done()
			}
		}()
	}
}

func generateLogs(count int) []logItem {
	logs := make([]logItem, count)

	for i := 0; i < count; i++ {
		logs[i] = logItem{
			action:    actions[rand.Intn(len(actions)-1)],
			timestamp: time.Now(),
		}
	}

	return logs
}

func prepareDB() {
	// check if folder exists
	if _, err := os.Stat("users"); os.IsNotExist(err) {
		err = os.Mkdir("users", 0755)
		dry.PanicIfErr(err)
	}

	// delete existing files in users folder
	files, err := ioutil.ReadDir("users")
	dry.PanicIfErr(err)

	for _, file := range files {
		err = os.Remove("users/" + file.Name())
		dry.PanicIfErr(err)
	}
}
