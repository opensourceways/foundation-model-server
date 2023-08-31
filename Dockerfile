FROM golang:1.18 as BUILDER

MAINTAINER TommyLike<tommylikehu@gmail.com>

# build binary
COPY . /go/src/github.com/opensourceways/foundation-model-server
RUN cd /go/src/github.com/opensourceways/foundation-model-server && GO111MODULE=on CGO_ENABLED=0 go build -o server

# copy binary config and utils
FROM golang:1.18
RUN groupadd -g 1000 fms
RUN useradd -u 1000 -g fms -s /bin/bash -m fms
USER fms
WORKDIR /home/fms
COPY --chown=fms --from=BUILDER /go/src/github.com/opensourceways/foundation-model-server/server /home/fms

ENTRYPOINT ["/home/fms/server"]
