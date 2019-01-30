package main

import (
	"log"
	"net/http" //web support for chat API
	"net"
	

	"github.com/gorilla/websocket" //library to make open constantly open connections
)

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan Message)           // broadcast channel
var clientlist = make([]string,1) //clientList username
var clientlistx = make([]*websocket.Conn,1)  //client List ws


// Configure the upgrader
//takes normal HTTP connections and upgrades them to Websockets
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Define our message object
type Message struct { //backticks are metadata to make sense of JSONs
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

func main() {
	// Create a simple file server
	fs := http.FileServer(http.Dir("../public")) //reads index.html
	http.Handle("/", fs)

	// Configure websocket route
	http.HandleFunc("/ws", handleConnections)

	// Start listening for incoming chat messages
	go handleMessages()

	// Start the server on localhost port 8000 and log any errors
	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil { //works continuosly if there's no error
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request and error (if occured) to a websocket
	ws, err := upgrader.Upgrade(w, r, nil) 
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection only when the function returns
	defer ws.Close()
	//get client ip
	ip,_,_ := net.SplitHostPort(r.RemoteAddr)
	log.Print(ip, " Connected")

	// Register our new client into map
	clients[ws] = true
	//to store username before exiting
	var user string
	for {
		firsttime:=true
		var msg Message //struct of type Message
		// Read a new message as JSON and maps it to a Message object called err
		err:=ws.ReadJSON(&msg)
		//add user to register
		if err ==nil{
			for i:=0;i<len(clientlist);i++{
				if clientlist[i]==msg.Username{
					firsttime=false
				}
			}
			if firsttime{
				clientlist=append(clientlist, msg.Username)
				clientlistx=append(clientlistx, ws)
			}
		}
		if err!=nil{ // only if error exists
			log.Print(user, " Exited")
			delete(clients, ws)
			//find username in array
			for i:=0;i<len(clientlist);i++{
				if clientlist[i]==user{
					clientlist=deleteelement(clientlist, i)
					clientlistx=deleteelementx(clientlistx, i)
				}
			}
			break
		}
		if msg.Message=="show"{
			log.Print(clientlist)
			continue //so that message is not broadcasted
		}
		// Send the newly received message to the broadcast channel
		broadcast <- msg
		if string(msg.Message[0])!="@"{
				log.Print("User ", msg.Username, " messaged : ",msg.Message)
		        user=(msg.Username)
		}		
	}
}

func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast //new variable
		// Send it out to every client that is currently connected
		for client := range clients { //client is a new variable
			if string(msg.Message[0])=="@"{
				log.Print("Private Message sent to ", getUsername(msg.Message))

				//check if present in clientlist
					for k:=range clientlist{
						if clientlist[k]==getUsername(msg.Message) ||clientlist[k]==msg.Username{
							err:=clientlistx[k].WriteJSON(msg)//send only to that client
							if err != nil {
								log.Printf("error: %v", err)
								client.Close()
								delete(clients, client)
							}
						}
					}
				break	
			}else{
				err := client.WriteJSON(msg) //write message into JSON for all chat
				if err != nil {
					log.Printf("error: %v", err)
					client.Close()
					delete(clients, client)
				}
			}
		}
	}
}

func getUsername( uin string) string{
	var uout string 
	for i:=0;i<len(uin)-1;i++{
		if string(uin[i])==" "{break}
		uout=uout+string(uin[i+1])
	}

	return uout[0:len(uout)-1]
}

//to delete element in array
func deleteelement(a []string, i int) []string{
	a[i] = a[len(a)-1] 
	a[len(a)-1] = ""   
	a = a[:len(a)-1] 
	return a
}
//to delete element in array x
func deleteelementx(a []*websocket.Conn, i int) []*websocket.Conn{
	a[i] = a[len(a)-1] 
	a[len(a)-1] = a[0]   
	a = a[:len(a)-1] 
	return a
}
