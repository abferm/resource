FROM golang:1.18.1-buster AS dev

RUN go install -v golang.org/x/tools/gopls@v0.7.4 && \
    go install -v honnef.co/go/tools/cmd/staticcheck@latest && \
    go install -v github.com/fullstorydev/grpcurl/cmd/grpcurl@latest && \
    go install -v golang.org/x/tools/cmd/goimports@latest && \
    go install -v gotest.tools/gotestsum@latest && \
    go install -v github.com/stamblerre/gocode@latest && \
    go install -v github.com/uudashr/gopkgs/v2/cmd/gopkgs@latest && \
    go install -v github.com/ramya-rao-a/go-outline@latest && \
    go install -v github.com/rogpeppe/godef@latest && \
    go install -v github.com/sqs/goreturns@latest && \
    go install -v golang.org/x/lint/golint@latest && \
    go install -v golang.org/x/tools/cmd/stringer@latest && \
    go install -v github.com/abice/go-enum@latest && \
    go install -v github.com/go-delve/delve/cmd/dlv@latest