# build static binary
FROM golang:1.16.5-alpine3.12 as builder 


WORKDIR /go/src/github.com/bots-house/webshot

# download dependencies 
COPY go.mod go.sum ./
RUN go mod download 

COPY . .

# git tag 
ARG BUILD_VERSION

# git commit sha
ARG BUILD_REF

# build time 
ARG BUILD_TIME

# compile 
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
      -ldflags="-w -s -extldflags \"-static\" -X \"main.buildVersion=${BUILD_VERSION}\" -X \"main.buildRef=${BUILD_REF}\" -X \"main.buildTime=${BUILD_TIME}\"" \
      -a \
      -tags timetzdata \
      -o /bin/webshot .


# run 
FROM chromedp/headless-shell:91.0.4472.114

COPY --from=builder /bin/webshot /bin/webshot
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/


# Reference: https://github.com/opencontainers/image-spec/blob/master/annotations.md
LABEL org.opencontainers.image.source="https://github.com/bots-house/webshot"

HEALTHCHECK --interval=10s --timeout=5s --retries=3 CMD [ "/bin/webshot", "--healthcheck" ]

EXPOSE 8000/tcp

ENTRYPOINT [ "/bin/webshot" ]