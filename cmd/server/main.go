package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/plankton4/chat-app-server/cmd/server/database"
	"github.com/plankton4/chat-app-server/cmd/server/database/mongodb"
	"github.com/plankton4/chat-app-server/cmd/server/fcm"
	"github.com/plankton4/chat-app-server/cmd/server/misc"
	"github.com/plankton4/chat-app-server/cmd/server/network"
	"github.com/plankton4/chat-app-server/cmd/server/user"
)

var addr = flag.String("addr", ":8048", "http service address")

func main() {
	flag.Parse()

	// handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/", network.ServeHome())
	mux.HandleFunc("/ping", network.Ping())
	mux.HandleFunc("/pingmysqldb", network.PingMysqlDB())
	mux.HandleFunc("/pingmongo", network.PingMongo())
	mux.HandleFunc("/pingservaddr", network.PingServerAddr())
	mux.HandleFunc("/applesigninauth", network.AppleSignInAuthHandler())
	mux.HandleFunc("/endregistration", network.RegistrationEndHandler())
	mux.HandleFunc("/lastvideo", network.GetLastVideoHandler())
	mux.HandleFunc("/uploadimage", network.UploadImageHandler())
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		network.ServeWs(w, r)
	})

	log.Println("CONNECTED TO REMOTE? ", misc.IsConnectedToRemoteServer())

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	// launch hubs
	network.SetupHubs()

	// database
	db := database.OpenDatabase()
	defer db.Close()

	mongodb.RunDB()
	defer mongodb.Disconnect()

	// hardcoded user for testing purposes.
	// Client app knows about this user's SessionKey and UserID.
	user.CreateGuestUser()

	// FCM
	fcm.SetupFCM()

	// for checking things
	//mainTest()

	// listen
	fmt.Println("Listen and serve...")

	error := http.ListenAndServe(*addr, mux)
	if error != nil {
		log.Fatal("ListenAndServe: ", error)
	}
}
