package utils

import (
	"errors"
	"log"
	"net"
	"time"
)

var ErrWaitTimeout = errors.New("timeout while waiting for resource")

// WaitFor helper function that waits until a specified network address become
// available.
// The retry count parameter specifies the maximum number of retries, until an error is
// returned.
// The retry interval specifies the duration between each attempt.
func WaitFor(address string, retryInternal time.Duration, maxRetryCount int) error {
	retryCount := 0
	for {

		retryCount++

		conn, err := net.Dial("tcp", address)

		if err != nil {
			log.Printf("Failed to connect to %s: %s", address, err.Error())
			time.Sleep(retryInternal)
		} else {
			defer conn.Close()
			break
		}

		if retryCount >= maxRetryCount {
			return ErrWaitTimeout
		}
	}

	return nil
}

// GetFreePort returns a free port number on localhost.
func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
