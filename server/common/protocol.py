import socket
import logging
import signal
from common.utils import Bet


def recv_exact(sock, n):
    """Recives exactly n bytes"""
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

def receive_bet_message(client_sock):
    """Handles the bet message coming from a client"""
    size_bytes = recv_exact(client_sock, 2)
    size = int.from_bytes(size_bytes, byteorder='big')
    msg = recv_exact(client_sock, size)
    msg = msg.decode('utf-8').strip()
    addr = client_sock.getpeername()

    logging.info(
        f'action: receive_message | result: success | ip: {addr[0]} | msg: {msg}')
    bet_fields = msg.split(",")
    agency_id, name, last_name, document, birthdate, number = bet_fields[
        0], bet_fields[1], bet_fields[2], bet_fields[3], bet_fields[4], bet_fields[5]

    return Bet(agency_id, name, last_name, document, birthdate, number)

def send_bet_response(client_sock, received_bet):
    """Sends the correct response to a new bet from a client if it was correctly processed"""
    send_all(client_sock, f"{received_bet.document},{received_bet.number}\n".encode('utf-8'))
