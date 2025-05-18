# LiTe wedDAV server

Simple webdav server with htpassword authentification and ssl.

# Build
```
git clone https://github.com/mrlinqu/ltdav
cd ./ltdav
make Build
```

# Run options
```
-d Directory to serve from. Default is CWD
-l address to listen. Default 0.0.0.0:9800
-c Path to TLS cert file
-k Path to TLS key file
-a Path to auth file
-r Auth realm text
```

# Auth
For authorisation use apache htpasswd file. For make this file may use standart apache util
https://httpd.apache.org/docs/current/programs/htpasswd.html


# Docker
`docker pull ghcr.io/mrlinqu/ltdav:latest`

## Simple
```
docker run -d \
    --name ltdav \
    -p 9800 \
    -v /var/ltdav:/data \
    ghcr.io/mrlinqu/ltdav:latest
```

## With auth
```
docker run -d \
    --name ltdav \
    -p 9800 \
    -v /var/ltdav:/ltdav \
    -e LTDAV_WORK_DIR='/ltdav/data' \
    -e LTDAV_AUTH_FILE='/ltdav/htpasswd' \
    ghcr.io/mrlinqu/ltdav:latest
```

## With TLS
```
docker run -d \
    --name ltdav \
    -p 9800 \
    -v /var/ltdav:/ltdav \
    -e LTDAV_WORK_DIR='/ltdav/data' \
    -e LTDAV_CERT_FILE='/ltdav/cert' \
    -e LTDAV_KEY_FILE='/ltdav/key' \
    ghcr.io/mrlinqu/ltdav:latest
```