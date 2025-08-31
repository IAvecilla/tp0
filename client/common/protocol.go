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

func sendBet(bet Bet, conn net.Conn) (*ServerResponse, error) {
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
	messageLength := len(betMessage)
	messageSize := uint16(messageLength)
	sizeBuffer := make([]byte, 2)
	binary.BigEndian.PutUint16(sizeBuffer, messageSize)

	_, err := conn.Write(sizeBuffer)
	_, err = conn.Write(messageBytes)
	
	msg, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return nil, err
	}
	msg = strings.TrimSpace(msg)
	responseFields := strings.Split(msg, ",")

	res := &ServerResponse{
		Document: responseFields[0],
		Number: strings.TrimRight(responseFields[1], "\n"),
	}
	return res, err
}