FROM golang:1.16 AS build

COPY . /workspace/

ENV CGO_ENABLED=0
RUN cd /workspace && go build .


FROM alpine:3.13
RUN apk add --no-cache ca-certificates bash
COPY --from=build /workspace/hcloud-cloud-controller-manager /bin/hcloud-cloud-controller-manager
ENTRYPOINT ["/bin/hcloud-cloud-controller-manager"]
