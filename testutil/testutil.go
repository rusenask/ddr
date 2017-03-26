package testutil

import (
	"context"
	"math/rand"
	"net"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/rusenask/ddr/messaging"
)

func Expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

type DummyDistributor struct {
	Serving        bool
	ShutdownCalled bool
	CleanupCalled  bool

	SharedImage         string
	StoppedSharingImage string

	PulledImage   string
	PulledTorrent []byte
}

func NewDummyDistributor() *DummyDistributor {
	return &DummyDistributor{}
}

func (d *DummyDistributor) Serve(ctx context.Context) error {
	d.Serving = true
	return nil
}

func (d *DummyDistributor) Shutdown() error {
	d.ShutdownCalled = true
	return nil
}
func (d *DummyDistributor) Cleanup() error {
	d.CleanupCalled = true
	return nil
}
func (d *DummyDistributor) ShareImage(name string) (torrent []byte, err error) {
	d.SharedImage = name
	return
}

func (d *DummyDistributor) StopSharing(name string) error {
	d.StoppedSharingImage = name
	return nil
}
func (d *DummyDistributor) PullImage(name string, torrent []byte) error {
	d.PulledImage = name
	d.PulledTorrent = torrent
	return nil
}

var NatsStarted bool

func StartTestingNats() error {

	//nats config:
	config := messaging.DefaultConfig()

PORT:
	// initialRoute will only be set if INITIAL environment variable defined

	//generat renadom number
	rand.Seed(time.Now().Unix())
	port := rand.Intn(99999-1024) + 1024
	//check port if open
	host := ":" + strconv.Itoa(port)
	server, err := net.Listen("tcp", host)
	// if it fails then the port is likely taken
	if err != nil {
		goto PORT
	}
	// close the server
	server.Close()

	//overwrtie port
	config.Port = port

	msg := messaging.NewNats(config)

	if !NatsStarted {
		NatsStarted = true
		return msg.StartServer()
	}

	return nil
}
