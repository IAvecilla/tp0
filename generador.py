import sys
import random


def write_server_service(output_file):
    """Write server service"""
    server_content = """  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    volumes:
      - ./server/config.ini:/config.ini
    environment:
      - PYTHONUNBUFFERED=1
    networks:
      - testing_net

"""
    output_file.write(server_content)


def write_clients_service(output_file, num_clients):
    """Write all client services"""
    nombres = ["Juan", "Lucas", "Sofia", "Martina", "Matias"]
    apellidos = ["Perez", "Diaz", "Gomez", "Lopez", "Martinez"]

    clients_content = ""
    for client_id in range(1, num_clients + 1):
        year = random.randint(2000, 2030)
        month = random.randint(1, 12)
        day = random.randint(1, 28)
        random_date = f"{year:04d}-{month:02d}-{day:02d}"
        random_dni = random.randint(1000000, 50000000)
        random_number = random.randint(1, 10000)

        # Use modulo to cycle through names if clients greater than names
        nombre = nombres[(client_id - 1) % len(nombres)]
        apellido = apellidos[(client_id - 1) % len(apellidos)]

        clients_content += f"""  client{client_id}:
    container_name: client{client_id}
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID={client_id}
      - CLI_NUMERO={random_number}
      - CLI_NOMBRE={nombre}
      - CLI_APELLIDO={apellido}
      - CLI_DOCUMENTO={random_dni}
      - CLI_NACIMIENTO={random_date}
    volumes:
      - ./client/config.yaml:/config.yaml
      - ./.data/agency-{client_id}.csv:/agency-data.csv\n
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
