package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"gopkg.in/gomail.v2"
)

func main() {
	t1 := time.Now()
	host := "smtp.example.com"
	username := "example@example.com"
	password := "password"

	from := username
	subject := "This is the email subject"
	body := "Hello <b>Bob</b> and <i>Cora</i>!"
	list := []string{"receive1@example.com", "receive2@example.com"}

	lens := len(list)
	// Set maximum goroutine number as 5
	if lens > 5 {
		lens = 5
	}

	ch := make(chan *gomail.Message, lens)
	clientch := make(chan gomail.SendCloser, lens)

	d := gomail.NewDialer(host, 994, username, password)
	//Use ssl if support
	d.SSL = true

	var wg sync.WaitGroup
	wg.Add(lens)
	for i := 0; i < lens; i++ {
		i := i
		go func() {
			defer wg.Done()
			var s gomail.SendCloser
			var err error
			open := false
			for {
				select {
				case msg, ok := <-ch:
					if !ok {
						clientch <- nil
						return
					}
					if !open {
						if s, err = d.Dial(); err != nil {
							fmt.Println(err)
							clientch <- s
							return
						}
						fmt.Println("goroutine:" + strconv.Itoa(i) + " client connection successfully.")
						clientch <- nil
						open = true
					}
					if err := gomail.Send(s, msg); err != nil {
						log.Print(err)
					}
					to := msg.GetHeader("To")[0]
					fmt.Println("Send to " + to + " successfully! run at goroutine:" + strconv.Itoa(i))
				}
			}
		}()
	}

	// Use the channel in your program to send emails.
	for _, to := range list {
		m := gomail.NewMessage()
		m.SetHeader("From", from)
		m.SetHeader("To", to)
		m.SetHeader("Subject", subject)
		m.SetBody("text/html", body)
		ch <- m
	}

	var senders []gomail.SendCloser
	for i := 0; i < lens; i++ {
		s := <-clientch
		senders = append(senders, s)
	}

	// Close the channel to stop the mail daemon.
	close(ch)
	// Waiting for all goroutine done.
	wg.Wait()
	// Close all SendCloser
	for i, s := range senders {
		if s != nil {
			s.Close()
			fmt.Println("client:" + strconv.Itoa(i) + " closed")
		}
	}
	t2 := time.Now()
	fmt.Println(t2.Sub(t1))
}
