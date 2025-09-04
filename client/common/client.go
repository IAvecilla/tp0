package common

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"encoding/csv"
	"io"
	"strconv"
	"time"
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
	file   *os.File
	reader *csv.Reader
}

// NewClient Initializes a new client receiving the configuration
// as a parameter.
func NewClient(config ClientConfig) *Client {
	// Open file once during initialization
	file, reader, err := openDataFile()
	if err != nil {
		log.Errorf("action: open_file | result: fail | error: %v", err)
		return nil
	}
	
	client := &Client{
		config: config,
		keepRunning: true,
		file: file,
		reader: reader,
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
	time.Sleep(1 * time.Second)
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

	// Close file when shutting down
	if c.file != nil {
		c.file.Close()
	}
}

// RunClient Send a set of messages to the server containing a batch of bets 
func (c *Client) RunClient() {
	if !c.keepRunning {
		log.Infof("action: receive_shutdown_signal | result: success")
		return
	}
	
	// Ensure file is closed when function ends
	defer func() {
		if c.file != nil {
			c.file.Close()
		}
	}()
	
	c.createClientSocket()
	
	// Send the initial message
	err := sentNewBetMessage(c.conn)
	if err != nil {
		log.Errorf("action: batch_sending | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	totalBetsSent := 0
	
	// Loop to handle all batches
	for {
		if !c.keepRunning {
			log.Infof("action: receive_shutdown_signal | result: success")
			return
		}
		
		// Get next batch from file
		betsToSend, err := c.getNextBatch(c.config.MaxBatchAmount)
		if err != nil && err != io.EOF {
			log.Errorf("action: batch_sending | result: fail | client_id: %v | error: %v",
				c.config.ID, err)
			return
		}
		
		// If no more bets, break the loop
		if len(betsToSend) == 0 {
			break
		}
		
		log.Infof("action: batch_sending | result: in_progress | batch_size: %v", len(betsToSend))
		response, err := sendBets(betsToSend, c.conn)
		
		if err != nil {
			log.Errorf("action: batch_sending | result: fail | client_id: %v | error: %v",
				c.config.ID, err)
			return
		}
		
		parsedBetAmount, _ := strconv.Atoi(response.BetsProcessedInBatch)
		totalBetsSent += parsedBetAmount
		log.Infof("action: batch_sending | result: success | bets_sent_in_batch: %v | total_bets_sent: %v", 
			parsedBetAmount, totalBetsSent)
		
		// If we got EOF, this was the last batch
		if err == io.EOF {
			break
		}
	}

		sendAllBetsSentMessage(c.conn)
		c.conn.Close()

		for {
			log.Infof("action: consulta_ganadores | result: in_progress")
			c.createClientSocket()

			results, keepRequesting, err := sendAskForResults(c.conn, c.config.ID)
			if !keepRequesting {
				log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %v", len(results))
				c.conn.Close()
				time.Sleep(100 * time.Millisecond)
				return
			} else if err != nil {
				log.Errorf("action: consulta_ganadores | result: fail | err: %s", err)
				return
			} else {
				c.conn.Close()
				time.Sleep(1000 * time.Millisecond)
			}
		}
	}

// openDataFile opens the CSV file and returns file handle and reader
func openDataFile() (*os.File, *csv.Reader, error) {
	file, err := os.Open("agency-data.csv")
	if err != nil {
		return nil, nil, err
	}
	
	reader := csv.NewReader(file)
	return file, reader, nil
}

// getNextBatch reads the next batch of bets from the CSV file
func (c *Client) getNextBatch(batchSize int) ([]Bet, error) {
	var bets []Bet
	
	for i := 0; i < batchSize; i++ {
		record, err := c.reader.Read()
		if err == io.EOF {
			// End of file reached
			return bets, err
		}
		if err != nil {
			log.Errorf("Error reading record: %v", err)
			continue
		}
		
		bet := Bet{
			AgencyId:  c.config.ID,
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
