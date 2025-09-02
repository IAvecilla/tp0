import socket
import logging
import signal
from common.utils import Bet, has_won

def encode_bet(bet):
    return f"{bet.agency},{bet.first_name},{bet.last_name},{bet.document},{bet.birthdate},{bet.number}"

def receive_bet_message(client_sock):
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
            client_sock.send("{}\n".format("ERR_INVALID_BET").encode('utf-8'))
            raise ValueError("Invalid Bet")
        
        processed_bets.append(Bet(bet_fields[0],bet_fields[1],bet_fields[2], bet_fields[3], bet_fields[4], bet_fields[5]))

    return processed_bets, True

def receive_new_message(client_sock):
    size = int.from_bytes(client_sock.recv(2), byteorder='big')
    data = b""
    while len(data) < size:
        packet = client_sock.recv(size - len(data))
        if not packet:
            raise ConnectionError("Connection closed unexpectedly")
        data += packet

    msg = data.decode('utf-8').strip()
    return msg

def send_winners(client_sock, agency_winners):
    if len(agency_winners) != 0:
        winners_message = "|".join(agency_winners)
        client_sock.send(f"{winners_message}".encode('utf-8'))
    else:
        client_sock.send("NO_WINNERS".encode('utf-8'))

def send_results_not_ready(client_sock):
    client_sock.send(f"NOT_READY".encode('utf-8'))

    