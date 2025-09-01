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
	messageSize := uint16(messageLength)
	sizeBuffer := make([]byte, 2)
	binary.BigEndian.PutUint16(sizeBuffer, messageSize)

	_, err := conn.Write(sizeBuffer)
	_, err = conn.Write(messageBytes)
	
	msg, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return nil, err
	}
	if msg == "ERR_INVALID_BET" {
		return nil, errors.New("Error processing bet")
	}
	msg = strings.TrimSpace(msg)
	responseFields := strings.Split(msg, ",")

	res := &ServerResponse{
		BetsProcessedInBatch: responseFields[0],
		TotalBetsProcessed: strings.TrimRight(responseFields[1], "\n"),
	}
	return res, err
}

func sendFinalMessage(conn net.Conn) {
	allBatchSendMessage := "ALL_SENT"
	messageBytes := []byte(allBatchSendMessage)
	messageLength := len(allBatchSendMessage)
	messageSize := uint16(messageLength)
	sizeBuffer := make([]byte, 2)
	binary.BigEndian.PutUint16(sizeBuffer, messageSize)

	_, _ = conn.Write(sizeBuffer)
	_, _ = conn.Write(messageBytes)
}