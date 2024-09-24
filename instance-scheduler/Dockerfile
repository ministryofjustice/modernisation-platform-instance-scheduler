# Stage 1: Build the Go binary
FROM public.ecr.aws/lambda/go:1 as builder

# Install the specific version of Go
ARG GO_VERSION=1.23.0
RUN curl -sSL https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz | tar -C /usr/local -xz && \
    ln -s /usr/local/go/bin/go /usr/local/bin/go && \
    ln -s /usr/local/go/bin/gofmt /usr/local/bin/gofmt

# Verify Go installation
RUN echo "Go Version: $(go version)" && go version

# Copy the rest of the source code
COPY . .

# Build the Go binary
RUN CGO_ENABLED=0 go build -o /instance-scheduler .

# Stage 2: Create the final image
FROM public.ecr.aws/lambda/go:1

# Copy the binary from the builder stage
COPY --from=builder /instance-scheduler ${LAMBDA_TASK_ROOT}

# Set the CMD to the binary
CMD [ "instance-scheduler" ]
