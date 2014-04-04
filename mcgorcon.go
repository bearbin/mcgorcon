// mc-gorcon is a Minecraft RCON Client written in Go.

package mcgorcon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"time"
)

const PACKET_TYPE_COMMAND int32 = 2
const PACKET_TYPE_AUTH int32 = 3
const REQUEST_ID_BAD_LOGIN int32 = -1

type Client struct {
	password   string
	connection net.Conn
}

type header struct {
	Size       int32
	RequestID  int32
	PacketType int32
}

// Dial up the server and establish a RCON conneciton.
func Dial(host string, port int, pass string) Client {
	// Combine the host and port to form the address.
	address := host + ":" + fmt.Sprint(port)
	// Actually establish the conneciton.
	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		panic(err)
	}
	// Create the client object, since the connection has been established.
	c := Client{password: pass, connection: conn}
	// TODO - server validation to make sure we're talking to a real RCON server.
	// For now, just return the client and assume it's a real server.
	return c
}

// SendCommand sends a command to the server and returns the result (often nothing).
func (c *Client) SendCommand(command string) string {
	// Because I'm lazy, just authenticate with every command.
	c.authenticate()
	// Generate the binary packet.
	packet := packetise(PACKET_TYPE_COMMAND, []byte(command))
	// Send the packet.
	response := c.sendPacket(packet)
	head, payload := depacketise(response)
	if head.RequestID == REQUEST_ID_BAD_LOGIN {
		// Auth was bad, panic.
		panic("NO AITH")
	}
	return payload
}

// authenticate authenticates the user with the server.
func (c *Client) authenticate() {
	// Generate the authentication packet.
	packet := packetise(PACKET_TYPE_AUTH, []byte(c.password))
	// Send the packet off to the server.
	response := c.sendPacket(packet)
	// Decode the return packet.
	fmt.Println(response)
	head, _ := depacketise(response)
	if head.RequestID == REQUEST_ID_BAD_LOGIN {
		// Auth was bad, panic.
		panic("BAD AUTH")
	}
}

// sendPacket sends the binary packet representation to the server and returns the response.
func (c *Client) sendPacket(packet []byte) []byte {
	// Send the packet over the wire.
	_, err := c.connection.Write(packet)
	if err != nil {
		panic("WRITE FAIL")
	}
	// Get a response.
	out := make([]byte, 4096)
	n, err := c.connection.Read(out)
	if err != nil {
		panic(err)
	}
	return out[:n]
}

// packetise encodes the packet type and payload into a binary representation to send over the wire.
func packetise(t int32, p []byte) []byte {
	// Generate a random request ID.
	ID := requestID()
	pad := [2]byte{}
	length := int32(len(p) + 10)
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, length)
	binary.Write(&buf, binary.LittleEndian, ID)
	binary.Write(&buf, binary.LittleEndian, t)
	binary.Write(&buf, binary.LittleEndian, p)
	binary.Write(&buf, binary.LittleEndian, pad)
	// Notchian server doesn't like big packets :(
	if buf.Len() >= 1460 {
		panic("Packet too big when packetising.")
	}
	// Return the bytes.
	fmt.Println(buf.Bytes())
	return buf.Bytes()
}

// depacketise decodes the binary packet into a native Go struct.
func depacketise(raw []byte) (header, string) {
	buf := bytes.NewBuffer(raw[:])
	head := header{}
	err := binary.Read(buf, binary.LittleEndian, &head)
	if err != nil {
		panic(err)
	}
	return head, buf.String()
}

// requestID returns a random positive integer to use as the request ID for an RCON packet.
func requestID() int32 {
	// Return a non-negative integer to use as the packet ID.
	id := rand.Int31()
	fmt.Println("USING ID", id)
	return id
}
