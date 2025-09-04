package common

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
	keepRunning bool
	bet Bet
}

// NewClient Initializes a new client receiving the configuration
// as a parameter along with the bet to send to the server
func NewClient(config ClientConfig, bet Bet) *Client {
	client := &Client{
		config: config,
		keepRunning: true,
		bet: bet,
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
}

// RunClient Send the bet message to the server
func (c *Client) RunClient() {
		if !c.keepRunning {
			log.Infof("action: receive_shutdown_signal | result: success")
			return
		}
		c.createClientSocket()

		response, err := sendBet(c.bet, c.conn)
		c.conn.Close()

		if response.Document == c.bet.Document && response.Number == c.bet.Number && err == nil {
			log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v", response.Document, response.Number)
		} else {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
		}
}
