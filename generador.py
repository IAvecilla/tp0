import sys

def write_server_service(output_file):
    output_file.write("  server:\n")
    output_file.write("    container_name: server\n")
    output_file.write("    image: server:latest\n")
    output_file.write("    entrypoint: python3 /main.py\n")
    output_file.write("    environment:\n")
    output_file.write("      - PYTHONUNBUFFERED=1\n")
    output_file.write("      - LOGGING_LEVEL=DEBUG\n")
    output_file.write("    networks:\n")
    output_file.write("      - testing_net\n")
    output_file.write("\n")

def write_clients_service(output_file, num_clients):
    for client_id in range(1, num_clients + 1):
        output_file.write(f"  client{client_id}:\n")
        output_file.write(f"    container_name: client{client_id}\n")
        output_file.write("    image: client:latest\n")
        output_file.write("    entrypoint: /client\n")
        output_file.write("    environment:\n")
        output_file.write(f"      - CLI_ID={client_id}\n")
        output_file.write("      - CLI_LOG_LEVEL=DEBUG\n")
        output_file.write("    networks:\n")
        output_file.write("      - testing_net\n")
        output_file.write("    depends_on:\n")
        output_file.write("      - server\n")
        output_file.write("\n")

def write_networks(output_file):
    output_file.write("networks:\n")
    output_file.write("  testing_net:\n")
    output_file.write("    ipam:\n")
    output_file.write("      driver: default\n")
    output_file.write("      config:\n")
    output_file.write("        - subnet: 172.25.125.0/24\n")

def write_compose_file(output_file: str, num_clients: int):
    with open(output_file, "w") as f:
        f.write("name: tp0\n")
        f.write("services:\n")

        write_server_service(f)
        write_clients_service(f, num_clients)
        write_networks(f)

if __name__ == "__main__":
    args = sys.argv
    write_compose_file(args[1], int(args[2]))