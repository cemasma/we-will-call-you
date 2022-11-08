package room

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"os"
	"sync/atomic"
	"time"
)

// RequestId is used for tagging each incoming http request for logging
// purposes.  The actual implementation is just the ObjectId implementation
// found in launchpad.net/mgo/bson.  This will most likely change and evolve
// into its own format.
type RequestId string

func (id RequestId) String() string {
	return fmt.Sprintf("%x", string(id))
}


// byteSlice returns byte slice of id from start to end.
// Calling this function with an invalid id will cause a runtime panic.
func (id RequestId) byteSlice(start, end int) []byte {
	if len(id) != 12 {
		panic(fmt.Sprintf("Invalid RequestId: %q", string(id)))
	}
	return []byte(string(id)[start:end])
}

// requestIdCounter is atomically incremented when generating a new ObjectId
// using NewObjectId() function. It's used as a counter part of an id.
var requestIdCounter uint32 = 0

// machineId stores machine id generated once and used in subsequent calls
// to NewObjectId function.
var machineId []byte

// initMachineId generates machine id and puts it into the machineId global
// variable. If this function fails to get the hostname, it will cause
// a runtime error.
func initMachineId() {
	var sum [3]byte
	hostname, err := os.Hostname()
	if err != nil {
		panic("Failed to get hostname: " + err.Error())
	}
	hw := md5.New()
	hw.Write([]byte(hostname))
	copy(sum[:3], hw.Sum(nil))
	machineId = sum[:]
}

// NewRequestId returns a new unique RequestId.
// This function causes a runtime error if it fails to get the hostname
// of the current machine.
func NewRequestId() RequestId {
	b := make([]byte, 12)
	// Timestamp, 4 bytes, big endian
	binary.BigEndian.PutUint32(b, uint32(time.Now().Unix()))
	// Machine, first 3 bytes of md5(hostname)
	if machineId == nil {
		initMachineId()
	}
	b[4] = machineId[0]
	b[5] = machineId[1]
	b[6] = machineId[2]
	// Pid, 2 bytes, specs don't specify endianness, but we use big endian.
	pid := os.Getpid()
	b[7] = byte(pid >> 8)
	b[8] = byte(pid)
	// Increment, 3 bytes, big endian
	i := atomic.AddUint32(&requestIdCounter, 1)
	b[9] = byte(i >> 16)
	b[10] = byte(i >> 8)
	b[11] = byte(i)
	return RequestId(b)
}