import sys

def write_server_service(output_file):
    """Write server service"""
    server_content = """  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
      - LOGGING_LEVEL=DEBUG
    networks:
      - testing_net

"""
    output_file.write(server_content)


def write_clients_service(output_file, num_clients):
    """Write all client services"""
    clients_content = ""
    for client_id in range(1, num_clients + 1):
        clients_content += f"""  client{client_id}:
    container_name: client{client_id}
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID={client_id}
      - CLI_LOG_LEVEL=DEBUG
    networks:
      - testing_net
    depends_on:
      - server

"""
    output_file.write(clients_content)


def write_networks(output_file):
    """Write networks section"""
    networks_content = """networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
"""
    output_file.write(networks_content)


def write_compose_file(output_file, num_clients):
    """Generate the docker compose file"""
    try:
        with open(output_file, "w") as f:
            # Write header
            f.write("name: tp0\nservices:\n")
            
            # Write all sections
            write_server_service(f)
            write_clients_service(f, num_clients)
            write_networks(f)
        
        print(f"Generated {output_file} with {num_clients} clients")
    except Exception as e:
        print(f"Error generating docker file: {e}")


if __name__ == "__main__":
    args = sys.argv
    write_compose_file(args[1], int(args[2]))
