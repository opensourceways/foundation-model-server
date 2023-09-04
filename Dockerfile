FROM openeuler/openeuler:23.03 as BUILDER

MAINTAINER TommyLike<tommylikehu@gmail.com>
RUN dnf update -y && dnf in -y golang
# build binary
COPY . /go/src/github.com/opensourceways/foundation-model-server
RUN cd /go/src/github.com/opensourceways/foundation-model-server && GOPROXY=https://goproxy.cn,direct GO111MODULE=on CGO_ENABLED=0 go build -o server

# copy binary config and utils
FROM openeuler/openeuler:22.03
RUN dnf update -y && dnf in -y shadow
RUN groupadd -g 1000 fms
RUN useradd -u 1000 -g fms -s /bin/bash -m fms
USER fms
WORKDIR /home/fms
COPY --chown=fms --from=BUILDER /go/src/github.com/opensourceways/foundation-model-server/server/foundation-model-server /home/fms/server

ENTRYPOINT ["/home/fms/server"]
