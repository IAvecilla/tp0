import socket
import logging
import signal
import multiprocessing
from common.utils import store_bets, load_bets, has_won
from common.protocol import receive_bet_batch_message, send_bet_response, send_results_not_ready, send_winners, encode_bet, receive_new_message


class Server:
    def __init__(self, port, listen_backlog, clients):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.settimeout(5.0)
        manager = multiprocessing.Manager()
        self._server_socket.listen(listen_backlog)
        self.shutdown = False
        self.finished_clients_lock = manager.Lock()
        self.finished_clients = manager.Value('i', 0)
        self.clients = int(clients)
        self._storage_lock = manager.Lock()
        self.active_processes = []
        self.final_winners_lock = manager.Lock()
        self.final_winners = manager.Value('i', 0)
        signal.signal(signal.SIGTERM, self.handle_sigterm)

    def handle_sigterm(self, _signum, _frame):
        """Handler for the SIGTERM signal"""
        logging.info(f"action: receive_shutdown_signal | result: in_progress")
        self.shutdown = True

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        while not self.shutdown:
            try:
                client_sock = self.__accept_new_connection()
                p = multiprocessing.Process(
                    target=self.__handle_client_connection,
                    args=(client_sock,)
                )
                p.daemon = True
                p.start()
                self.active_processes.append(p)
                
                client_sock.close()
                self._cleanup_finished_processes()
            except socket.timeout:
                # Timeout is expected, just check shutdown flag
                continue
            except Exception as e:
                logging.error(
                    f"Error processing client connection: {e}")
        else:
            if self._server_socket:
                self._server_socket.close()
            logging.info(f"action: receive_shutdown_signal | result: success")

    def _cleanup_finished_processes(self):
        """Remove finished processes from active list"""
        self.active_processes = [p for p in self.active_processes if p.is_alive()]

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            total_received_bets = 0
            msg = receive_new_message(client_sock)
            while not self.shutdown:
                if msg == "NEW_BET":
                    logging.info(f"action: total_apuestas_recibidas | result: in_progress")
                    received_bets, keep_reading = receive_bet_batch_message(client_sock)
                    if received_bets and keep_reading:
                        with self._storage_lock:
                            store_bets(received_bets)
                        logging.info(f"action: apuesta_recibida | result: success | cantidad: {len(received_bets)}")
                        send_bet_response(client_sock, received_bets, total_received_bets)
                        total_received_bets += len(received_bets)       
                    else:
                        with self.finished_clients_lock:
                            self.finished_clients.value += 1
                        logging.info(f"action: total_apuestas_recibidas | result: success | cantidad: {total_received_bets}")
                        break
                if msg.startswith("BET_RESULT"):
                    logging.info("action: sorteo | result: in_progress")
                    with self.finished_clients_lock:
                        current_finished_clients = self.finished_clients.value

                    if self.clients == current_finished_clients:
                        agency_id = msg.split(",")[1]
                        if len(self.final_winners) == 0:
                            with self._storage_lock:
                                final_bets = load_bets()
                            with self.final_winners_lock:
                                self.final_winners = [bet for bet in final_bets if has_won(bet)]
                        agency_winners = [encode_bet(bet) for bet in self.final_winners if bet.agency == int(agency_id)]
                        send_winners(client_sock, agency_winners)
                        logging.info("action: sorteo | result: success")
                        break
                    else:
                        send_results_not_ready(client_sock)
                        break
        except OSError as e:
            logging.error(
                "action: receive_message | result: fail | error: {e}")
        except Exception as e:
            logging.error(
                f"Error processing client requests: {e}")
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
        logging.info(
            f'action: accept_connections | result: success | ip: {addr[0]}')
        return c
