# build static binary
FROM golang:1.16.3-alpine3.12 as builder 


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
FROM chromedp/headless-shell:90.0.4430.212

COPY --from=builder /bin/webshot /bin/webshot

# HEALTHCHECK --interval=30s --timeout=30s --start-period=5s --retries=3 CMD [ "/bin/webshot", "-health" ]

EXPOSE 8000


# Reference: https://github.com/opencontainers/image-spec/blob/master/annotations.md
LABEL org.opencontainers.image.source="https://github.com/bots-house/webshot"

EXPOSE 8000/tcp

ENTRYPOINT [ "/bin/webshot" ]