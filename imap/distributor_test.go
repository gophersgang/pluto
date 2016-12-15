package imap_test

import (
	"log"
	"testing"
	"time"

	"github.com/numbleroot/pluto/imap"
	"github.com/numbleroot/pluto/utils"
)

// Functions

// TestInitDistributor executes a white-box unit test
// on the implemented InitDistributor() function.
func TestInitDistributor(t *testing.T) {

	var err error

	// Create needed test environment.
	config, _, err := utils.CreateTestEnv()
	if err != nil {
		log.Fatal(err)
	}

	go func() {

		// Correct storage initialization.
		storage, err := imap.InitStorage(config)
		if err != nil {
			t.Fatalf("[imap_test.TestInitDistributor] Expected correct storage initialization but failed with: '%s'\n", err.Error())
		}

		// Close the socket after 500ms.
		time.AfterFunc((1000 * time.Millisecond), func() {
			log.Println("[imap_test.TestInitDistributor] Timeout reached, closing storage socket. BEWARE.")
			storage.MailSocket.Close()
			storage.SyncSocket.Close()
		})

		// Run the storage node.
		_ = storage.Run()
	}()

	time.Sleep(400 * time.Millisecond)

	go func() {

		// Correct worker initialization.
		worker, err := imap.InitWorker(config, "worker-1")
		if err != nil {
			t.Fatalf("[imap_test.TestInitDistributor] Expected correct worker-1 initialization but failed with: '%s'\n", err.Error())
		}

		// Close the socket after 500ms.
		time.AfterFunc((600 * time.Millisecond), func() {
			log.Println("[imap_test.TestInitDistributor] Timeout reached, closing worker-1 socket. BEWARE.")
			worker.MailSocket.Close()
			worker.SyncSocket.Close()
		})

		// Run the worker.
		_ = worker.Run()
	}()

	time.Sleep(400 * time.Millisecond)

	// Correct distributor initialization.
	distr, err := imap.InitDistributor(config)
	if err != nil {
		t.Fatalf("[imap_test.TestInitDistributor] Expected correct distributor initialization but failed with: '%s'\n", err.Error())
	}

	distr.Socket.Close()

	time.Sleep(800 * time.Millisecond)
}
