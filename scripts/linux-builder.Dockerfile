# Image used by scripts/build.sh to build the Linux GUI on a non-Linux host.
# The frontend is built on the host; this image only compiles Go + GTK/WebKit.
FROM golang:1.27.3-bookworm
RUN apt-get update \
 && apt-get install -y --no-install-recommends build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev \
 && rm -rf /var/lib/apt/lists/*
RUN go install github.com/wailsapp/wails/v2/cmd/wails@v2.17.1
