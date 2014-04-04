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
	// Generate the binary packet.
	packet := packetise(PACKET_TYPE_COMMAND, []byte(command))
	// Send the packet.
	response := c.sendPacket(packet)
	head, payload := dePacketise(response)
	if head.RequestID == REQUEST_ID_BAD_LOGIN {
		// Auth was bad, panic.
		panic("NO AITH")
	}
	return payload
}

// Authenticate authenticates the user with the server.
func (c *Client) Authenticate() {
	// Generate the authentication packet.
	packet := packetise(PACKET_TYPE_AUTH, []byte(c.password))
	// Send the packet off to the server.
	response := c.sendPacket(packet)
	// Decode the return packet.
	head, _ := dePacketise(response)
	if head.RequestID == REQUEST_ID_BAD_LOGIN {
		// Auth was bad, panic.
		panic("BAD AITH")
	}
}

// sendPacket sends the binary packet representation to the server and returns the response.
func (c *Client) sendPacket(packet []byte) []byte {
	// Send the packet over the wire.
	_, err := c.connection.Write(packet)
	if err != nil {
		panic("WRTE FAIL")
	}
	// Get a response.
	var obuf []byte
	_, err = c.connection.Read(obuf)
	//if err != nil {
	//	panic("RDE FAILE")
	//}
	return obuf
}

// packetise encodes the packet type and payload into a binary representation to send over the wire.
func packetise(t int32, p []byte) []byte {
	// Generate a random request ID.
	ID := requestID()
	pad := [2]byte{}
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, ID)
	binary.Write(&buf, binary.LittleEndian, t)
	binary.Write(&buf, binary.LittleEndian, p)
	binary.Write(&buf, binary.LittleEndian, pad)
	payload := buf.Bytes()
	// Get the length of the payload.
	length := int32(len(payload))
	// Assemble the full buffer now.
	buf.Reset()
	binary.Write(&buf, binary.LittleEndian, length)
	binary.Write(&buf, binary.LittleEndian, payload)
	// Notchian server doesn't like big packets :(
	if buf.Len() >= 1460 {
		panic("Packet too big when packetising.")
	}
	// Return the bytes.
	return buf.Bytes()
}

// depacketise decodes the binary packet into a native Go struct.
func dePacketise(raw []byte) (header, string) {
	buf := bytes.NewBuffer(raw[:])
	head := header{}
	_ = binary.Read(buf, binary.LittleEndian, &head)
	//if err != nil {
	//	panic(err)
	//}
	return head, buf.String()
}

// requestID returns a random positive integer to use as the request ID for an RCON packet.
func requestID() int32 {
	// Return a non-negative integer to use as the packet ID.
	return rand.Int31()
}
