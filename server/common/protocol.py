import socket
import logging
import signal
from common.utils import Bet

def receive_bet_message(client_sock):
    size = int.from_bytes(client_sock.recv(2), byteorder='big')
    data = b""
    while len(data) < size:
        packet = client_sock.recv(size - len(data))
        if not packet:
            raise ConnectionError("Connection closed unexpectedly")
        data += packet

    msg = data.decode('utf-8').strip()
    addr = client_sock.getpeername()
    logging.info(f'action: receive_message | result: success | ip: {addr[0]} | msg: {msg}')
    bet_fields = msg.split(",")
    return Bet(bet_fields[0],bet_fields[1],bet_fields[2], bet_fields[3], bet_fields[4], bet_fields[5])