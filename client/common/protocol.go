package common

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"encoding/binary"
	"errors"
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
	BetsProcessedInBatch string
	TotalBetsProcessed string
}

// Writes every byte of data in the socket
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

func encodeBet(bet Bet) (string) {
	betMessage := fmt.Sprintf(
			"%v,%s,%s,%s,%s,%v\n",
				bet.AgencyId,
				bet.Name,
				bet.LastName,
				bet.Document,
				bet.Birthdate,
				bet.Number,
			)
	
	return betMessage
}

func sendBets(bets []Bet, conn net.Conn) (*ServerResponse, error) {
	var encodedBets []string
	for _, bet := range bets {
		encodedBets = append(encodedBets, encodeBet(bet))
	}

	finalBetMessage := strings.Join(encodedBets, "|")
	messageBytes := []byte(finalBetMessage)
	messageLength := len(finalBetMessage)

	if messageLength > 8192 {
		log.Error("action: send_message | result: fail | error: message larger than 8kb")
		return nil, errors.New("message larger than 8kb, consider changing the max amount of bets per batch")
	}

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

	if msg == "ERR_INVALID_BET" {
		return nil, fmt.Errorf("Error processing bet on the server")
	}

	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}
	
	msg = strings.TrimSpace(msg)
	responseFields := strings.Split(msg, ",")
	
	if len(responseFields) < 2 {
		return nil, fmt.Errorf("invalid response format: %s", msg)
	}
	
	res := &ServerResponse{
		BetsProcessedInBatch: responseFields[0],
		TotalBetsProcessed: strings.TrimRight(responseFields[1], "\n"),
	}
	return res, err
}

func sendAllBetsSentMessage(conn net.Conn) error {
	allBatchSendMessage := "ALL_SENT"
	messageBytes := []byte(allBatchSendMessage)
	messageLength := len(allBatchSendMessage)
	messageSize := uint16(messageLength)
	sizeBuffer := make([]byte, 2)
	binary.BigEndian.PutUint16(sizeBuffer, messageSize)

	if err := writeAll(conn, sizeBuffer); err != nil {
		return fmt.Errorf("error writing size: %w", err)
	}
	
	if err := writeAll(conn, messageBytes); err != nil {
		return fmt.Errorf("error writing message: %w", err)
	}

	return nil
}
