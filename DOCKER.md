# Using Docker

1. Get the PUID and PGID of the user the data directory is owned by. This can be done by running the following command:
```bash
echo -e "PUID $(id -u) PGID $(id -g)"
```
 
2. Create a `docker-compose.yml` file

```yaml
version: '3.3'
services:
  seanime:
    image: 5rahim/seanime:latest
    container_name: seanime
    volumes:
      - <path/to/library>:/collection
      - <path/to/data/dir>:/config
    ports:
      - 43211:43211
    environment:
      - SEANIME_DATA_DIR=/config
      - SEANIME_SERVER_HOST=0.0.0.0
      - SEANIME_SERVER_PORT=43211
      - PUID=<puid>
      - PGID=<pgid>
    restart: unless-stopped
    network_mode: "host"
```

Replace `<path/to/library>`, `<path/to/data/dir>`, `<puid>`, and `<pgid>` with the actual values.

3. Run the following command to start the container

```bash
docker-compose up -d
```

4. Run the following commands to update the container

```bash
docker-compose pull
docker-compose up -d
```
