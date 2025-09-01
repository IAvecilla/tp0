import socket
import logging
import signal
from common.utils import Bet

def receive_bet_message(client_sock):
    processed_bets = []
    size = int.from_bytes(client_sock.recv(2), byteorder='big')
    data = b""
    while len(data) < size:
        packet = client_sock.recv(size - len(data))
        if not packet:
            raise ConnectionError("Connection closed unexpectedly")
        data += packet

    msg = data.decode('utf-8').strip()
    
    if msg == "ALL_SENT":
        return None, False

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