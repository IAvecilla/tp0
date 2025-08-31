import sys
import random

def write_server_service(output_file):
    output_file.write("  server:\n")
    output_file.write("    container_name: server\n")
    output_file.write("    image: server:latest\n")
    output_file.write("    entrypoint: python3 /main.py\n")
    output_file.write("    environment:\n")
    output_file.write("      - PYTHONUNBUFFERED=1\n")
    output_file.write("    volumes:\n")
    output_file.write("      - ./server/config.ini:/config.ini\n")
    output_file.write("    networks:\n")
    output_file.write("      - testing_net\n")
    output_file.write("\n")

def write_clients_service(output_file, num_clients):
    nombres = ["Juan", "Lucas", "Sofia", "Martina", "Matias"]
    apellidos = ["Perez", "Diaz", "Gomez", "Lopez", "Martinez"]

    
    for client_id in range(1, num_clients + 1):
        year = random.randint(2000, 2030)
        month = random.randint(1, 12)
        day = random.randint(1, 28)
        random_date = f"{year:04d}-{month:02d}-{day:02d}"
        random_dni = random.randint(1000000, 50000000)
        random_number = random.randint(1, 10000)
        output_file.write(f"  client{client_id}:\n")
        output_file.write(f"    container_name: client{client_id}\n")
        output_file.write("    image: client:latest\n")
        output_file.write("    entrypoint: /client\n")
        output_file.write("    environment:\n")
        output_file.write(f"      - CLI_ID={client_id}\n")
        output_file.write(f"      - CLI_NUMERO={random_number}\n")
        output_file.write(f"      - CLI_NOMBRE={nombres[client_id]}\n")
        output_file.write(f"      - CLI_APELLIDO={apellidos[client_id]}\n")
        output_file.write(f"      - CLI_DOCUMENTO={random_dni}\n")
        output_file.write(f"      - CLI_NACIMIENTO={random_date}\n")
        output_file.write("    volumes:\n")
        output_file.write("      - ./client/config.yaml:/config.yaml\n")
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