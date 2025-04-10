package handler

import (
	"context" // State handling across API boundaries; part of native GoLang API
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	// Generic firebase support
	// Firestore-specific support
)

var ctx context.Context

const CountreCollection = "countries"

var client *firestore.Client

/*
Returns Firebase context and initializes if not already done.
*/
func GetFirebaseContext() context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return ctx
}

/*
Initializes Firebase client and returns reference.
Returns error if problems during initialization occur.
*/
func GetFirebaseClient() (*firestore.Client, error) {
	// Firebase initialisation
	ctx = GetFirebaseContext()

	// We use a service account, load credentials file that you downloaded from your project's settings menu.
	// It should reside in your project directory.
	// Make sure this file is git-ignored, since it is the access token to the database.
	sa := option.WithCredentialsFile("./handler/cloudassignment2-test-firebase-adminsdk-fbsvc-3a8f40042b.json")
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Instantiate client
	client, err := app.Firestore(ctx)

	// Alternative setup, directly through Firestore (without initial reference to Firebase); but requires Project ID; useful if multiple projects are used
	// client, err := firestore.NewClient(ctx, projectID)

	// Check whether there is an error when connecting to Firestore
	if err != nil {
		log.Println(err)
		return client, err
	}

	return client, nil
}

/*
Reads a string from the body in plain-text and sends it to Firestore to be registered as a document.
*/
func addDocument(w http.ResponseWriter, r *http.Request) {

	log.Println("Received " + r.Method + " request.")

	// very generic way of reading body; should be customized to specific use case
	content, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Reading payload from body failed.")
		http.Error(w, "Reading payload failed.", http.StatusInternalServerError)
		return
	}
	log.Println("Received request to add document for content ", string(content))
	if len(string(content)) == 0 {
		log.Println("Content appears to be empty.")
		http.Error(w, "Your payload (to be stored as document) appears to be empty. Ensure to terminate URI with /.", http.StatusBadRequest)
		return
	} else {
		// Add element in embedded structure.
		// Note: this structure is defined by the client, not the server!; it exemplifies the use of a complex structure
		// and illustrates how you can use Firestore features such as Firestore timestamps.
		s := Webhook{}
		err := json.Unmarshal(content, &s)
		if err != nil {
			log.Println("Error unmarshalling payload.")
			http.Error(w, "Error unmarshalling payload.", http.StatusInternalServerError)
			return
		}
		/*
			// Update timestamp
			s.Created = time.Now()
			// Counter for this particular entry
			s.Ct = ct*/

		// Get Firebase client
		client, err := GetFirebaseClient()
		if err != nil {
			log.Println("Error getting Firebase client: " + err.Error())
			http.Error(w, "Error establishing database connection.", http.StatusInternalServerError)
			return
		}
		// Ensure Firebase client is properly closed
		defer func() {
			errClose := client.Close()
			if errClose != nil {
				log.Fatal("Closing of the Firebase client failed. Error:", errClose)
			}
		}()

		// Add entry to database
		id, _, err2 := client.Collection(CountreCollection).Add(GetFirebaseContext(), s)

		if err2 != nil {
			// Error handling
			log.Println("Error when adding document " + string(content) + ", Error: " + err2.Error())
			http.Error(w, "Error when adding document "+string(content)+", Error: "+err2.Error(), http.StatusBadRequest)
			return
		} else {
			// Returns document ID in body
			log.Println("Document added to collection. Identifier of returned document: " + id.ID)
			http.Error(w, id.ID, http.StatusCreated)
			return
		}
	}
}

/*
Deletes Countries from Firestore DB. Variably supports deletion of all Countries or specific one, with unique
identifier provided as part of path.
*/
func deleteDocument(w http.ResponseWriter, r *http.Request) {
	log.Println("Received " + r.Method + " request.")

	messageId := r.PathValue("id")

	if messageId == "" {
		log.Println("Deleting all Countres.")

		// Get Firebase client
		client, err := GetFirebaseClient()
		if err != nil {
			log.Println("Error getting Firebase client: " + err.Error())
			http.Error(w, "Error establishing database connection.", http.StatusInternalServerError)
			return
		}
		// Ensure Firebase client is properly closed
		defer func() {
			errClose := client.Close()
			if errClose != nil {
				log.Fatal("Closing of the Firebase client failed. Error:", errClose)
			}
		}()

		it := client.Collection(CountreCollection).Documents(GetFirebaseContext())

		for {
			doc, err := it.Next()
			if errors.Is(err, iterator.Done) {
				break
			}

			log.Println("Deleting document ", doc.Data())

			_, err2 := doc.Ref.Delete(GetFirebaseContext())
			if err2 != nil {
				log.Println("Error deleting document from database. Error: " + err.Error())
				http.Error(w, "Error deleting document from database.", http.StatusInternalServerError)
				return
			}

		}

	} else {
		log.Println("Deleting specific country: " + messageId)

		// Get Firebase client
		client, err := GetFirebaseClient()
		if err != nil {
			log.Println("Error getting Firebase client: " + err.Error())
			http.Error(w, "Error establishing database connection.", http.StatusInternalServerError)
			return
		}
		// Ensure Firebase client is properly closed
		defer func() {
			errClose := client.Close()
			if errClose != nil {
				log.Fatal("Closing of the Firebase client failed. Error:", errClose)
			}
		}()
		_, err2 := client.Collection(CountreCollection).Doc(messageId).Delete(GetFirebaseContext())
		if err2 != nil {
			log.Println("Error deleting document from database. Error: " + err2.Error())
			http.Error(w, "Error deleting document from database.", http.StatusInternalServerError)
			return
		}

		log.Println("Deleted country " + messageId)
		http.Error(w, "Deleted country.", http.StatusNoContent)
		return
	}
}
