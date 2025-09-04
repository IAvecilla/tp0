package common

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"encoding/binary"
)

type Bet struct {
	AgencyId 	  string
	Name          string
	LastName      string
	Document      string
	Birthdate     string
	Number        string
}

type ServerResponse struct {
	Document string
	Number string
}

// Writes every inside data to the socket
func writeAll(conn net.Conn, data []byte) error {
	totalWritten := 0
	for totalWritten < len(data) {
		n, err := conn.Write(data[totalWritten:])
		if err != nil {
			return err
		}
		totalWritten += n
	}
	return nil
}

// Forms the message of the bet following the protocol:
//
// AgencyId,Name,LastName,Document,Birthdate,Number
//
// And waits for the response from the server
func sendBet(bet Bet, conn net.Conn) (*ServerResponse, error) {
	// Creates the bet message
	betMessage := fmt.Sprintf(
		"%v,%s,%s,%s,%s,%v\n",
		bet.AgencyId,
		bet.Name,
		bet.LastName,
		bet.Document,
		bet.Birthdate,
		bet.Number,
	)
	
	messageBytes := []byte(betMessage)
	messageLength := len(messageBytes)
	messageSize := uint16(messageLength)
	
	// Create buffer of message size
	sizeBuffer := make([]byte, 2)
	binary.BigEndian.PutUint16(sizeBuffer, messageSize)
	
	// Write the size of the message to the server
	if err := writeAll(conn, sizeBuffer); err != nil {
		return nil, fmt.Errorf("error writing size: %w", err)
	}
	
	// Send the actual bet message
	if err := writeAll(conn, messageBytes); err != nil {
		return nil, fmt.Errorf("error writing message: %w", err)
	}
	
	// Read and parse the server response
	reader := bufio.NewReader(conn)
	msg, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}
	
	msg = strings.TrimSpace(msg)
	responseFields := strings.Split(msg, ",")
	
	if len(responseFields) < 2 {
		return nil, fmt.Errorf("invalid response format: %s", msg)
	}
	
	res := &ServerResponse{
		Document: responseFields[0],
		Number:   strings.TrimRight(responseFields[1], "\n"),
	}
	
	return res, nil
}
