// Package leapmotion is a WebSocket API for Leap Motion v6 JSON
// https://developer.leapmotion.com/documentation/javascript/supplements/Leap_JSON.html
package leapmotion

import (
	"errors"
	"math"

	"golang.org/x/net/websocket"
)

const (
	defaultLeapWebSocketAddress = "ws://localhost:6437/v6.json"
)

// DeviceEvent is sent from the server to the client when the Leap Motion when the service/daemon
// is paused or resumed and when the controller hardware is plugged in or unplugged:
// TODO: this is not implemented
type DeviceEvent struct {
	ID        string `json:"id"`
	Attached  bool   `json:"attached"`
	Streaming bool   `json:"streaming"`
	Type      string `json:"type"`
}

// Frame represents the tracking data format
// https://developer.leapmotion.com/documentation/javascript/supplements/Leap_JSON.html#json-tracking-data-format
type Frame struct {
	CurrentFrameRate float64        `json:"currentFrameRate"`
	ID               float64        `json:"id"`
	R                [][]float64    `json:"r"`
	S                float64        `json:"s"`
	T                []float64      `json:"t"`
	Timestamp        int            `json:"timestamp"`
	Gestures         []Gesture      `json:"gestures"`
	Hands            []Hand         `json:"hands"`
	InteractionBox   InteractionBox `json:"interactionBox"`
	Pointables       []Pointable    `json:"pointables"`
}

// Gesture represents a Gesture object in a Frame
type Gesture struct {
	Center        []float64 `json:"center"`
	Direction     []float64 `json:"direction"`
	Duration      int       `json:"duration"`
	HandsIDs      []int     `json:"handIds"`
	ID            int       `json:"id"`
	Normal        []float64 `json:"normal"`
	PointableIDs  []int     `json:"pointableIds"`
	Position      []float64 `json:"position"`
	Progress      float64   `json:"progress"`
	Radius        float64   `json:"radius"`
	Speed         float64   `json:"speed"`
	StartPosition []float64 `json:"startPosition"`
	State         string    `json:"state"`
	Type          string    `json:"type"`
}

// Hand represents a Hand object in a Frame
type Hand struct {
	ArmBasis               []float64   `json:"armBasis"`
	ArmWidth               float64     `json:"armWidth"`
	Confidence             float64     `json:"confidence"`
	Direction              []float64   `json:"direction"`
	Elbow                  []float64   `json:"elbow"`
	GrabStrength           float64     `json:"grabStrength"`
	ID                     int         `json:"id"`
	PalmNormal             []float64   `json:"palmNormal"`
	PalmPosition           []float64   `json:"PalmPosition"`
	PalmVelocity           []float64   `json:"PalmVelocity"`
	PinchStrength          float64     `json:"pinchStrength"`
	R                      [][]float64 `json:"r"`
	S                      float64     `json:"s"`
	SphereCenter           []float64   `json:"sphereCenter"`
	SphereRadius           float64     `json:"sphereRadius"`
	StabilizedPalmPosition []float64   `json:"stabilizedPalmPosition"`
	T                      []float64   `json:"t"`
	TimeVisible            float64     `json:"TimeVisible"`
	Type                   string      `json:"type"`
	Wrist                  []float64   `json:"wrist"`
}

// InteractionBox represents an interactionBox in a Frame
type InteractionBox struct {
	Center []int     `json:"center"`
	Size   []float64 `json:"size"`
}

// NormalizePoint Normalizes the coordinates of a point using the interaction box.
// Coordinates from the Leap Motion frame of reference (millimeters) are
// converted to a range of [0..1] such that the minimum value of the
// InteractionBox maps to 0 and the maximum value of the InteractionBox maps to 1.
func (i *InteractionBox) NormalizePoint(position []float64, clamp bool) ([]float64, error) {
	if i.Center == nil || len(i.Center) < 3 {
		return nil, errors.New("Center isn't set or doesn't have enough values")
	}
	if i.Size == nil || len(i.Size) < 3 {
		return nil, errors.New("Size isn't set or doesn't have enough values")
	}
	if position == nil || len(position) < 3 {
		return nil, errors.New("postion isn't set or doesn't have enough values")
	}

	vec := []float64{0, 0, 0}
	vec[0] = ((position[0] - float64(i.Center[0])) / i.Size[0]) + 0.5
	vec[1] = ((position[1] - float64(i.Center[1])) / i.Size[1]) + 0.5
	vec[2] = ((position[2] - float64(i.Center[2])) / i.Size[2]) + 0.5

	vec[0] = ((position[0] - float64(i.Center[0])) / i.Size[0]) + 0.5
	vec[1] = ((position[1] - float64(i.Center[1])) / i.Size[1]) + 0.5
	vec[2] = ((position[2] - float64(i.Center[2])) / i.Size[2]) + 0.5

	if clamp {
		vec[0] = math.Min(math.Max(vec[0], 0), 1)
		vec[1] = math.Min(math.Max(vec[1], 0), 1)
		vec[2] = math.Min(math.Max(vec[2], 0), 1)
	}

	return vec, nil
}

// Pointable represents a Pointable in a Frame
type Pointable struct {
	Bases                 []float64 `json:"bases"`
	BtipPosition          []float64 `json:"btipPosition"`
	CarpPosition          []float64 `json:"carpPosition"`
	DipPosition           []float64 `json:"dipPosition"`
	Direction             []float64 `json:"direction"`
	Extended              bool      `json:"extended"`
	HandID                int       `json:"handId"`
	ID                    int       `json:"id"`
	Length                float64   `json:"length"`
	McpPosition           []float64 `json:"mcpPosition"`
	PipPosition           []float64 `json:"pipPosition"`
	StabilizedTipPosition []float64 `json:"stabilizedTipPosition"`
	TimeVisible           float64   `json:"timeVisible"`
	TipPosition           []float64 `json:"tipPosition"`
	TipVelocity           []float64 `json:"tipVelocity"`
	Tool                  bool      `json:"tool"`
	TouchDistance         float64   `json:"touchDistance"`
	TouchZone             string    `json:"touchZone"`
	Type                  int       `json:"type"`
	Width                 float64   `json:"width"`
}

// Client represents a connection to a Leap Motion WebSocket server
type Client struct {
	ws           *websocket.Conn
	frameHandler func(*Frame)
	done         chan struct{}
}

// Connect to WebSocket and pass a frameHandler that is called whenever the WebSocket
// sends frame data
func Connect(frameHandler func(frame *Frame)) (*Client, error) {
	conn, err := websocket.Dial(defaultLeapWebSocketAddress, "", "http://localhost/")
	if err != nil {
		return nil, err
	}

	c := &Client{
		ws:           conn,
		done:         make(chan struct{}),
		frameHandler: frameHandler,
	}

	// Enable gestures recognition from leap sensor
	if err := websocket.JSON.Send(c.ws, map[string]bool{"enableGestures": true}); err != nil {
		return nil, err
	}

	// Enable our application to run in the background and receive messages
	if err := websocket.JSON.Send(c.ws, map[string]bool{"backgroundMessage": true}); err != nil {
		return nil, err
	}

	go c.processData() // loops until socket is closed

	return c, nil
}

func (c *Client) processData() {
	defer close(c.done)
	data := &Frame{}
	for {
		if err := websocket.JSON.Receive(c.ws, data); err != nil {
			continue
		}

		if c.frameHandler != nil {
			c.frameHandler(data)
		}
	}
}

// Close the websocket and stop processData for loop
func (c *Client) Close() error {
	if c.ws == nil {
		return nil
	}
	return c.ws.Close()
}

// Done returns a read only channel to know when the client is closed
func (c *Client) Done() <-chan struct{} {
	return c.done
}
