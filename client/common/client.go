package common

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"encoding/csv"
	"io"
	"strconv"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	MaxBatchAmount int

}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
	keepRunning bool
	bets []Bet
}

// NewClient Initializes a new client receiving the configuration
// as a parameter along with the bet to send to the server
func NewClient(config ClientConfig, bet Bet) *Client {
	bets, err := loadTotalBets(config)
	if err != nil {
		return nil
	}
	client := &Client{
		config: config,
		keepRunning: true,
		bets: bets,
	}

	// If the SIGTERM signal is received it will be sent to the sigc channel triggering
	// the shutdown in the goroutine
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGTERM)
	go func() {
		<-sigc
		client.shutdown()
	}()
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

// Handles graceful shutdown when SIGTERM is received
func (c *Client) shutdown() {
	log.Infof("action: receive_shutdown_signal | result: in_progress")
	if c.keepRunning {
		c.keepRunning = false
	}
}

// RunClient Send a set of messages to the server containing a batch of bets 
func (c *Client) RunClient() {
		if !c.keepRunning {
			log.Infof("action: receive_shutdown_signal | result: success")
			return
		}
		c.createClientSocket()
		totalBetAmount := len(c.bets)
		betsSent := 0
		for i := c.config.MaxBatchAmount; i < totalBetAmount; i = i + c.config.MaxBatchAmount {
			betsInBatch := c.bets[betsSent:i]
			response, err := sendBets(betsInBatch, c.conn)
			log.Infof("action: batch_sending | result: in_progress | bets_already_sent: %v | total_bets: %v", betsSent, totalBetAmount)
			if err != nil {
				log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
					c.config.ID,
					err,
				)
				c.conn.Close()
				return
			}
			parsedBetAmount, _ := strconv.Atoi(response.BetsProcessedInBatch)
			betsSent += parsedBetAmount
			log.Infof("action: batch_sending | result: success | bets_already_sent: %v | total_bets: %v", betsSent, totalBetAmount)
		}

		if betsSent < totalBetAmount && betsSent + c.config.MaxBatchAmount >= totalBetAmount {
			betsInBatch := c.bets[betsSent:totalBetAmount]
			response, err := sendBets(betsInBatch, c.conn)
			log.Infof("action: batch_sending | result: in_progress | bets_already_sent: %v | total_bets: %v", betsSent, totalBetAmount)
			if err != nil {
				log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
					c.config.ID,
					err,
				)
				c.conn.Close()
				return
			}
			parsedBetAmount, _ := strconv.Atoi(response.BetsProcessedInBatch)
			betsSent += parsedBetAmount
			log.Infof("action: batch_sending | result: success | bets_already_sent: %v | total_bets: %v", betsSent, totalBetAmount)
		}

		sendAllBetsSentMessage(c.conn)
		c.conn.Close()
}

func loadTotalBets(config ClientConfig) ([]Bet, error) {
	file, err := os.Open("agency-data.csv")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
    var bets []Bet

    for {
        record, err := reader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Infof("Error reading record: %v", err)
            continue
        }
        
        bet := Bet{
            AgencyId:  config.ID,
            Name:      record[0],
            LastName:  record[1],
            Document:  record[2],
            Birthdate: record[3],
            Number:    record[4],
        }
        
        bets = append(bets, bet)
    }
    
    return bets, nil
}
