package dispatcher

import (
	"fmt"
	"time"
	"github.com/teslamotors/vehicle-command/internal/authentication"
	universal "github.com/teslamotors/vehicle-command/pkg/protocol/protobuf/universalmessage"
)

var receiverBufferSize = 10

const (
	uuidLength      = 16
	challengeLength = 16
	addressLength   = 16
)

// receiverKey represents the key identifying a specific receiver.
type receiverKey struct {
	address [addressLength]byte
	uuid    [uuidLength]byte
	domain  universal.Domain
}

// String formats the receiver key for easier display.
func (r *receiverKey) String() string {
	return fmt.Sprintf("<%02x-%02x: %s>", r.address, r.uuid, r.domain)
}

// receiver represents a vehicle's pending response to a command.
type receiver struct {
	key           *receiverKey
	ch            chan *universal.RoutableMessage
	dispatcher    *Dispatcher
	requestSentAt time.Time
	lastActive    time.Time
	antireplay    authentication.SlidingWindow
	requestID     []byte
}

// Recv returns a channel that receives responses to the command that created the receiver.
func (r *receiver) Recv() <-chan *universal.RoutableMessage {
	// Log opening the receiver channel for debugging purposes
	fmt.Println("Receiver channel opened:", r.key)
	return r.ch
}

// Close tells the dispatcher to stop listening for responses to this command, freeing the corresponding resources.
func (r *receiver) Close() {
	// Log closing the receiver channel for debugging purposes
	fmt.Println("Closing receiver:", r.key)
	if r.dispatcher != nil {
		r.dispatcher.closeHandler(r)
	} else {
		fmt.Println("Error: Dispatcher is nil, cannot close receiver.")
	}
}

// expired returns true if the request was sent long enough ago that any included session info
// should be discarded as stale.
func (r *receiver) expired(lifetime time.Duration) bool {
	// Check for both request sent time and last active time expiration
	return time.Now().After(r.requestSentAt.Add(lifetime)) || time.Now().After(r.lastActive.Add(lifetime))
}
