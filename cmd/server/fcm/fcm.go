package fcm

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/plankton4/chat-app-server/cmd/server/misc"
	"google.golang.org/api/option"
)

var FirebaseApp *firebase.App

func SetupFCM() {
	log.Println("INIT FIREBASE")
	FirebaseApp = initializeAppWithServiceAccount()

	badge := 1
	sound := "default"

	iptoken := []string{"fVEaWXndgUjJt2mJKcdaPd:APA91bGwbhvIlSbpRQkdm9CSDm02rSodbRIdSKnEwq_SqIyshYPkjzbthX9rz7Zig15LbwQt2bvM6GOl9CZ0mjc3VaCsDK77AnjeJr8e__k8GAZWD9CwYWfX7pawczytc_3jywusHnCG"}

	SendMulticast(
		FirebaseApp,
		iptoken, //[]string{"coWx0HCFrkaNv9GoPp1lr9:APA91bGv-smoymxlS6MGkcmQDWEuzHPo1a7uMW0L2bAwQDwh1lZI9wqBHRkqDuY6tNagR3b4OwreaR4T1AVBNa5TCGtysYJNW2ocp2f-KyW0mwu3asWAFqYWo8MXfu6J9vhj5zO1rJls"},
		messaging.Notification{
			Title: "Duck",
			Body:  "Hi!",
		},
		messaging.Aps{
			Badge: &badge,
			Sound: sound,
		},
		map[string]string{},
	)

	//go sendToToken(FirebaseApp, "")
}

func initializeAppWithServiceAccount() *firebase.App {
	dir, _ := os.Getwd()

	if dir == "" {
		return nil
	}

	path := ""
	if misc.IsConnectedToRemoteServer() {
		path = filepath.Join(dir, "/somepath.json") // WORK
	} else {
		// use your own private key file in JSON format.
		// See: "Add the Firebase Admin SDK to your server"
		// https://firebase.google.com/docs/admin/setup
		path, _ = filepath.Abs("/Users/plankton4/chatapp-fc4fa-6146605b7a05.json")
	}

	opt := option.WithCredentialsFile(path)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	//log.Println("OPTION ", opt)

	return app
}

func sendToToken(app *firebase.App, token string, notificationData messaging.Notification) {
	// [START send_to_token_golang]
	// Obtain a messaging.Client from the App.
	ctx := context.Background()
	client, err := app.Messaging(ctx)
	if err != nil {
		log.Fatalf("error getting Messaging client: %v\n", err)
	}

	// See documentation on defining a message payload.
	badge := 1
	message := &messaging.Message{
		Notification: &notificationData,
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Badge: &badge,
					Sound: "default",
				},
			},
		},
		Token: token,
	}

	// Send a message to the device corresponding to the provided
	// registration token.
	response, err := client.Send(ctx, message)
	if err != nil {
		log.Fatalln(err)
	}
	// Response is a message ID string.
	fmt.Println("Successfully sent message:", response)
	// [END send_to_token_golang]
}

/// tokens is a list containing up to 500 registration tokens.
func SendMulticast(
	app *firebase.App,
	tokens []string, notificationData messaging.Notification,
	aps messaging.Aps,
	customData map[string]string) {
	// [START send_multicast]

	// Obtain a messaging.Client from the App.
	ctx := context.Background()
	client, err := app.Messaging(ctx)
	if err != nil {
		log.Fatalf("error getting Messaging client: %v\n", err)
	}

	message := &messaging.MulticastMessage{
		Notification: &notificationData,
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &aps,
			},
		},
		Data:   customData,
		Tokens: tokens,
	}

	br, err := client.SendMulticast(context.Background(), message)
	if err != nil {
		log.Fatalln(err)
	}

	// See the BatchResponse reference documentation
	// for the contents of response.
	fmt.Printf("%d messages were sent successfully\n", br.SuccessCount)
	// [END send_multicast]
}

func sendMulticastAndHandleErrors(app *firebase.App, tokens []string) {
	// [START send_multicast_error]

	// tokens is a list containing up to 500 registration tokens.

	// Obtain a messaging.Client from the App.
	ctx := context.Background()
	client, err := app.Messaging(ctx)
	if err != nil {
		log.Fatalf("error getting Messaging client: %v\n", err)
	}

	message := &messaging.MulticastMessage{
		Data: map[string]string{
			"score": "850",
			"time":  "2:45",
		},
		Tokens: tokens,
	}

	br, err := client.SendMulticast(context.Background(), message)
	if err != nil {
		log.Fatalln(err)
	}

	if br.FailureCount > 0 {
		var failedTokens []string
		for idx, resp := range br.Responses {
			if !resp.Success {
				// The order of responses corresponds to the order of the registration tokens.
				failedTokens = append(failedTokens, tokens[idx])
			}
		}

		fmt.Printf("List of tokens that caused failures: %v\n", failedTokens)
	}
	// [END send_multicast_error]
}
