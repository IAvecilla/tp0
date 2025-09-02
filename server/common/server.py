import socket
import logging
import signal
from common.utils import Bet, store_bets, load_bets, has_won
from common.protocol import receive_bet_message, receive_new_message, send_results_not_ready, send_winners, encode_bet

class Server:
    def __init__(self, port, listen_backlog, clients):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self.shutdown = False
        self.finished_clients = 0
        self.clients = int(clients)

    def handle_sigterm(self, signum, frame):
        self.shutdown = True
        logging.info(f"action: receive_shutdown_signal | result: in_progress")
        if self._server_socket:
            self._server_socket.close()

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        signal.signal(signal.SIGTERM, self.handle_sigterm)

        # the server
        while not self.shutdown:
            try:
                client_sock = self.__accept_new_connection()
                self.__handle_client_connection(client_sock)
            except Exception as e:
                logging.info(f"Error trying to establish a connection with client: {e}")
        else:
            logging.info(f"action: receive_shutdown_signal | result: success")

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            total_received_bets = 0
            msg = receive_new_message(client_sock)
            while True:
                if msg == "NEW_BET":
                    logging.info(f"action: total_apuestas_recibidas | result: in_progress")
                    received_bets, keep_reading = receive_bet_message(client_sock)
                    if received_bets and keep_reading:
                        store_bets(received_bets)
                        logging.info(f"action: apuesta_recibida | result: success | cantidad: {len(received_bets)}")
                        client_sock.send(f"{len(received_bets)},{total_received_bets}\n".encode('utf-8'))
                        total_received_bets += len(received_bets) 
                    elif not keep_reading:
                        logging.info(f"action: total_apuestas_recibidas | result: success | cantidad: {total_received_bets}")
                        self.finished_clients += 1
                        break
                if msg == "BET_RESULT":
                    logging.info("action: sorteo | result: in_progress")
                    print(self.clients)
                    print(self.finished_clients)
                    if self.clients == self.finished_clients:
                        print("A")
                        agency_id = receive_new_message(client_sock)
                        print("B")
                        final_bets = load_bets()
                        winners = [bet for bet in final_bets if has_won(bet)]
                        print("C")
                        agency_winners = [encode_bet(bet) for bet in winners if bet.agency == int(agency_id)]
                        send_winners(client_sock, agency_winners)
                        print("D")
                        logging.info("action: sorteo | result: success")
                        break
                    else:
                        send_results_not_ready(client_sock)
                        logging.info("action: sorteo | result: fail")
                        break
        except OSError as e:
            logging.error("action: receive_message | result: fail | error: {e}")
        finally:
            client_sock.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c
