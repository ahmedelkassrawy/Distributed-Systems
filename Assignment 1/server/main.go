// The chat server.
// by3ml inbox ysta2bl feh msgs untill the clients "polls" for them
// 
// go run ./server -addr :9000


package main

import (
	"errors"
	"flag"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"time"

	"chat/shared"
)

//lw m3mlsh poll n3tbro disconnected
const silenceLimit = 6 * time.Second


type User struct {
	lastPoll time.Time 
	inbox    []string 
}


type Chat struct {
	mu    sync.Mutex       //3shan tb2a locked for once bas
	users map[string]*User   
}

//Join bnzwd new user to chat w error lw tl3 mawgood 2bl kda
func (c *Chat) Join(name string, reply *bool) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return errors.New("username is empty")
	}

	c.mu.Lock()
	defer c.mu.Unlock() 

	if _, taken := c.users[name]; taken {
		return errors.New("username already taken")
	}

	c.users[name] = &User{lastPoll: time.Now()}

	c.broadcast(name, "* "+name+" joined the chat")
	log.Println(name, "joined")

	return nil
}

//Send byb2t msg lkol nas except sender.
func (c *Chat) Send(args shared.SendArgs, reply *bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.users[args.Name]; !ok {
		return errors.New("you are not joined")
	}

	c.broadcast(args.Name, args.Name+": "+args.Text)
	return nil
}

//Fetch bta3 polling
func (c *Chat) Fetch(name string, reply *[]string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	user, ok := c.users[name]
	if !ok {
		return errors.New("you are not joined (removed for being silent?)")
	}

	user.lastPoll = time.Now()
	*reply = user.inbox
	user.inbox = nil
	return nil
}

//list asamy elnas aw users
func (c *Chat) List(_ bool, reply *[]string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for name := range c.users {
		*reply = append(*reply, name)
	}

	sort.Strings(*reply)
	return nil
}


func (c *Chat) Leave(name string, reply *bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.removeUser(name)
	return nil
}

//nsheel user w nbroadcast llkol
func (c *Chat) removeUser(name string) {
	if _, ok := c.users[name]; !ok {
		return
	}
	delete(c.users, name)
	c.broadcast(name, "* "+name+" left the chat")
	log.Println(name, "left")
}



func (c *Chat) broadcast(except, line string) {
	for name, user := range c.users {
		if name != except {
			user.inbox = append(user.inbox, line)
		}
	}
}

//hnsheel elsaket b2alo kteer 
func (c *Chat) removeSilentUsers() {
	for {
		time.Sleep(time.Second)

		c.mu.Lock()

		for name, user := range c.users {
			if time.Since(user.lastPoll) > silenceLimit {
				log.Println(name, "stopped polling, removing")
				c.removeUser(name)
			}
		}
		c.mu.Unlock()
	}
}

func main() {
	addr := flag.String("addr", ":9000", "address to listen on")
	flag.Parse()

	chat := &Chat{users: make(map[string]*User)}

	//register server
	rpcServer := rpc.NewServer()
	if err := rpcServer.RegisterName("Chat", chat); err != nil {
		log.Fatal(err)
	}

	listener, err := net.Listen("tcp", *addr)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("server listening on", listener.Addr(), "- press Ctrl+C to stop")

	go chat.removeSilentUsers()

	//CTRL + C 
	go func() {
		stop := make(chan os.Signal, 1)

		signal.Notify(stop, os.Interrupt)
		<-stop

		log.Println("shutting down")
		listener.Close()
	}()

	//Accpet byakhod connections 3latool 
	for {
		conn, err := listener.Accept()

		if err != nil {
			return // listener was closed by Ctrl+C
		}

		go rpcServer.ServeConn(conn)
	}
}
