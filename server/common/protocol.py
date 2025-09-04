import socket
import logging
import signal
from common.utils import Bet, has_won

def encode_bet(bet):
    """Encodes bet into string"""
    return f"{bet.agency},{bet.first_name},{bet.last_name},{bet.document},{bet.birthdate},{bet.number}"

def receive_bet_batch_message(client_sock):
    """Receives and handles a batch of bets"""
    processed_bets = []
    msg = receive_new_message(client_sock)
    
    if msg == "ALL_SENT":
        return None, False

    processed_bets = []
    addr = client_sock.getpeername()
    logging.info(f'action: receive_message | result: success | ip: {addr[0]} | msg: {msg}')

    total_bets = msg.split("|")
    for bet in total_bets:
        bet_fields = bet.split(",")
        if len(bet_fields) != 6:
            logging.error(f"action: apuesta_recibida | result: fail | cantidad: {len(total_bets)}")
            send_all(client_sock, "ERR_INVALID_BET\n".encode('utf-8'))
            raise ValueError("Invalid Bet")
        
        agency_id, name, last_name, document, birthdate, number = bet_fields[0], bet_fields[1], bet_fields[2], bet_fields[3], bet_fields[4], bet_fields[5]
        processed_bets.append(Bet(agency_id, name, last_name, document, birthdate, number))

    return processed_bets, True

def receive_new_message(client_sock):
    """Process the message received from the socket and encode it in utf-8"""
    size_bytes = recv_exact(client_sock, 2)
    size = int.from_bytes(size_bytes, byteorder='big')
    msg = recv_exact(client_sock, size)
    msg = msg.decode('utf-8').strip()
    return msg

def send_winners(client_sock, agency_winners):
    """Sends a message with the loterry winners for the specific agency"""
    if len(agency_winners) != 0:
        winners_message = "|".join(agency_winners)
        send_all(client_sock, f"{winners_message}\n".encode('utf-8'))
    else:
       send_all(client_sock, "NO_WINNERS\n".encode('utf-8'))


def send_results_not_ready(client_sock):
    """Sends a message indicating that the results of the lottery are not ready yet"""
    send_all(client_sock, "NOT_READY\n".encode('utf-8'))
    
def recv_exact(sock, n):
    """Recieves exactly n bytes"""
    data = b""
    while len(data) < n:
        packet = sock.recv(n - len(data))
        if not packet:
            raise ConnectionError("Connection closed unexpectedly")
        data += packet
    return data

def send_all(sock, data):
    """Writes all the data in the socket"""
    total_sent = 0
    while total_sent < len(data):
        sent = sock.send(data[total_sent:])
        if sent == 0:
            raise ConnectionError("Socket connection broken")
        total_sent += sent

def send_bet_response(client_sock, received_bets, total_received_bets):
    """Sends the correct response to new batches of bets from a client if they were correctly processed"""
    send_all(client_sock, f"{len(received_bets)},{total_received_bets}\n".encode('utf-8'))
