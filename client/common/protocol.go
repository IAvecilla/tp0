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

// Response expected from the server
type ServerResponse struct {
	BetsProcessedInBatch string
	TotalBetsProcessed string
}

// Sends two messages through the socket, one with the length and other with actual message
func sendMessage(conn net.Conn, message string) error {
	messageBytes := []byte(message)
	messageLength := len(message)

	if messageLength > 8192 {
		log.Error("action: send_message | result: fail | error: message larger than 8kb")
		return fmt.Errorf("message larger than 8kb, consider changing the max amount of bets per batch")
	}
	messageSize := uint16(messageLength)
	
	// Create buffer of message size
	sizeBuffer := make([]byte, 2)
	binary.BigEndian.PutUint16(sizeBuffer, messageSize)
	
	// Write the size of the message to the server
	if err := writeAll(conn, sizeBuffer); err != nil {
		return fmt.Errorf("error writing message size: %w", err)
	}
	
	// Send the actual bet message
	if err := writeAll(conn, messageBytes); err != nil {
		return fmt.Errorf("error writing message: %w", err)
	}

	return nil
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

// Encodes the bet as a string to send as message
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

// Sends a batch of bets
func sendBets(bets []Bet, conn net.Conn) (*ServerResponse, error) {
	var encodedBets []string
	for _, bet := range bets {
		encodedBets = append(encodedBets, encodeBet(bet))
	}
	finalBetMessage := strings.Join(encodedBets, "|")
	
	// Send the actual batch of bets message
	if err := sendMessage(conn, finalBetMessage); err != nil {
		return nil, err
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

// Send the last message of the batch sending process indicating that all of them were sent
func sendAllBetsSentMessage(conn net.Conn) error {
	allBatchSentMessage := "ALL_SENT"
	if err := sendMessage(conn, allBatchSentMessage); err != nil {
		return err
	}

	return nil
}

// Send the message to ask for the results of the lottery
func sendAskForResults(conn net.Conn, agencyId string) ([]string, bool, error) {
	resultMessage := fmt.Sprintf("%s,%s", "BET_RESULT", agencyId) 
	if err := sendMessage(conn, resultMessage); err != nil {
		return nil, false, err
	}

	msg, err := bufio.NewReader(conn).ReadString('\n')
	if msg == "NOT_READY" {
		return []string{}, true, err
	} else if msg == "NO_WINNERS" {
		return []string{}, false, err
	}

	winners := strings.Split(msg, "|")
	return winners, false, nil
}

// Send the initial message that indicates that batches of bets are going to be send from now on
func sentNewBetMessage(conn net.Conn) error {
	newBetMessage := "NEW_BET"
	if err := sendMessage(conn, newBetMessage); err != nil {
		return err
	}

	return nil
}
