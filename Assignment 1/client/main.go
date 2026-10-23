//go run ./client -name kassra

//chat client
//2 things done : 1.goroutine bys2al server 3la message frequently w prints them
//				  2.main func reading what we typed w sends tthem to server 
package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/rpc"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"chat/shared"
)
 
const pollInterval = 500 * time.Millisecond

func main() {
	serverAddr := flag.String("server", "127.0.0.1:9000", "address of the chat server")
	name := flag.String("name", "", "your username")
	flag.Parse()

	input := bufio.NewScanner(os.Stdin)


	if *name == "" {
		fmt.Print("Username: ")
		input.Scan()

		*name = input.Text()
	}
	*name = strings.TrimSpace(*name)

	//connect
	client, err := rpc.Dial("tcp", *serverAddr)

	if err != nil {
		fmt.Println("cannot connect to the server:", err)
		os.Exit(1)
	}

	
	var noReply bool

	//bn7awl n3ml join bs lw mafeesh esm we get an error and exit
	if err := client.Call("Chat.Join", *name, &noReply); err != nil {
		fmt.Println("cannot join:", err)
		os.Exit(1)
	}
	fmt.Println("Joined as -", *name, "- type a message or /users to list users or /quit to leave")

	//3shan my7slh race condition lw 2 goroutines 7awlo y3ml exit
	var exitOnce sync.Once

	//goodbye func
	leave := func() {
		exitOnce.Do(func() {
			client.Call("Chat.Leave", *name, &noReply)
			fmt.Println("Goodbye")

			os.Exit(0)
		})
	}

	// Ctrl+C:
	go func() {
		interrupt := make(chan os.Signal, 1)

		signal.Notify(interrupt, os.Interrupt)
		<-interrupt

		leave()
	}()

	// Polling -> this goroutine runs ma3 typing loop waiting for input
	go func() {
		for {
			var lines []string
			if err := client.Call("Chat.Fetch", *name, &lines); err != nil {
				// server fasl or closed 
				exitOnce.Do(func() {
					fmt.Println("we lost contact with the server:", err)
					os.Exit(1)
				})
			}
			for _, line := range lines {
				fmt.Println(line)
			}
			time.Sleep(pollInterval)
		}
	}()

	//nfdl nloop le7ad mhaga ttktb
	for input.Scan() {
		text := strings.TrimSpace(input.Text())

		switch {
		case text == "":
		case text == "/quit":
			leave()
		case text == "/users":
			var users []string

			if err := client.Call("Chat.List", false, &users); err != nil {
				fmt.Println("error:", err)
			} else {
				fmt.Println("online:", strings.Join(users, ", "))
			}
		default:
			args := shared.SendArgs{Name: *name, Text: text}

			if err := client.Call("Chat.Send", args, &noReply); err != nil {
				fmt.Println("error:", err)
			}
		}
	}

	leave()
}
