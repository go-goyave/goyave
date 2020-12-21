---
meta:
  - name: "og:title"
    content: "Deployment - Goyave"
  - name: "twitter:title"
    content: "Deployment - Goyave"
  - name: "title"
    content: "Deployment - Goyave"
---

# Deployment

[[toc]]

## Introduction

There are some important details to think about before deploying your Goyave application to production. We are going to cover them in this section of the guide, and make sure your applications are deployed properly.

## Application configuration

Be sure to deploy your application with a `config.production.json` config file containing the correct values for your production environment and **set the `GOYAVE_ENV` environment variable to `production`.**

1. Ensure that the `app.environment` entry is `production`.
2. The `server.host` entry should be `0.0.0.0` if you want to open access to your service from anywhere. If you're using Apache or Nginx as a proxy on the same machine, keep it at `127.0.0.1` so the server will only be accessible through the proxy.
3. Change the `server.domain` entry to your domain name, if you use one.
4. The `server.port` and `server.httpsPort` will very likely require a change. Most of the time, you need `80` and `443` respectively.
5. If you use `https`, be sure to provide the paths to your `server.tls.cert` and `server.tls.key`. Learn more [here](./configuration.html#setting-up-https).
6. `server.debug` **must** be set do `false`. You don't want anyone to get important information about your internal errors and therefore your code when an error occurs.
7. Change your database connection credentials. `database.autoMigrate` should be set to `false`.

## Build

Of course, don't run your application with `go run` in production. Build your application using `go build` and deploy the executable, alongside the config files and resources directory.

The following Dockerfile is an example of a goyave application called `docker-goyave`:
``` docker
FROM golang:alpine as builder

WORKDIR /app

COPY . .

RUN go build -ldflags "-w -s"

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/docker-goyave ./docker-goyave
COPY resources /app/resources
COPY config.production.json ./config.production.json

RUN useradd -r -U go-exec -M -d /app -s /bin/false
RUN chown -R go-exec /app

USER go-exec

EXPOSE 80
ENV GOYAVE_ENV=production

ENTRYPOINT [ "./docker-goyave" ]
```

## Deamon

Goyave applications are standalone and don't require any web server (such as Apache or Nginx) to function. However, this also means that you will need some additional installation steps on your production server.

When running a web service, it is important that it stays online. Creating a **deamon** is necessary so your server can run in the background, have the correct filesystem permissions, and restart automatically after a crash or a reboot.

You can achieve this with any supervisor program, but in this guide, we will use systemd as it's the most common one.

### Systemd

First, we need to create a user and its group for our application:
```
sudo addgroup goyave
sudo useradd -r -s /bin/false goyave
sudo usermod -a -G goyave goyave
```

You may use the following systemd unit file to deploy your Goyave application:

`/etc/systemd/system/my-goyave-application.service`:
```
[Unit]
Description=Goyave application deamon

# If you are using a database, mysql for example, make sure the application
# service starts AFTER the database service.
# After=mysqld.service
After=network.target
StartLimitIntervalSec=0

[Service]
Environment=GOYAVE_ENV=production
WorkingDirectory=/path/to/goyave-application
ExecStart=/path/to/goyave-application/executable

# Process management
####################

Type=simple

# Can be changed to "on-failure"
Restart=always
RestartSec=1
TimeoutStopSec=600

# Run as goyave:goyave
User=goyave
Group=goyave

# Hardening measures
####################

# Provide a private /tmp and /var/tmp.
PrivateTmp=true

# Mount /usr, /boot/ and /etc read-only for the process.
ProtectSystem=full

# Deny access to /home, /root and /run/user
ProtectHome=true

# Disallow the process and all of its children to gain
# new privileges through execve().
NoNewPrivileges=true

# Use a new /dev namespace only populated with API pseudo devices
# such as /dev/null, /dev/zero and /dev/random.
PrivateDevices=true

# Deny the creation of writable and executable memory mappings.
MemoryDenyWriteExecute=true

[Install]
WantedBy=multi-user.target
```

::: warning
Replace `my-goyave-application` with the name of your application and `/path/to/goyave-application` with the **absolute** path to your application. This is important to have unique application names in case you run multiple goyave applications on the same server. You should also create a separate user and group for each application.
:::

Now let's run the service and enable it on system startup:
```
sudo systemctl start my-goyave-application
sudo systemctl enable my-goyave-application
```
