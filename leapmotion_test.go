package leapmotion

import (
	"fmt"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	wait := make(chan struct{})

	// Exit the test as soon as we get a frame of data
	f := func(frame *Frame) {
		fmt.Println(frame)
		close(wait)
	}
	// Create a new client
	c, err := Connect(f)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close() // stop the client connection

	select {
	case <-wait:
	case <-time.After(time.Second * 5):
		t.Fatal("TestConnect timed out. Make sure you have a leap sensor connected and use it within 5 seconds")
	}
}

func TestNormalizePoint(t *testing.T) {
	tests := []struct {
		position []float64
		clamp    bool
		expected []float64
	}{
		{[]float64{0.355, 0.532342, 0.134234}, true, []float64{0, 0.03234199999999998, 0}},
		{[]float64{1, 1, 1}, false, []float64{0.500000, 0.500000, 0.500000}},
	}

	interactionBox := InteractionBox{
		Center: []int{1, 1, 1},
		Size:   []float64{1, 1, 1},
	}

	for _, test := range tests {
		normalizedPosition, err := interactionBox.NormalizePoint(test.position, test.clamp)
		if err != nil {
			t.Fatal(err)
		}

		for i, p := range normalizedPosition {
			if p != test.expected[i] {
				t.Fatalf("Received %f. Expected %f", normalizedPosition, test.expected)
			}
		}
	}
}
