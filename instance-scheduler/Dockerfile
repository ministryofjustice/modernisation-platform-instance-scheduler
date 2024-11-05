# Stage 1: Build the Go binary
FROM golang:1.23 as builder

# Set the Go working directory
WORKDIR /app

# Copy the rest of the source code
COPY . .

# Build the Go binary
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o /app/instance-scheduler .

# Stage 2: Create the final image with Lambda runtime
FROM public.ecr.aws/lambda/provided:al2023

# Copy the built binary from the builder stage to the Lambda task root
COPY --from=builder /app/instance-scheduler ${LAMBDA_TASK_ROOT}

# Set the entry point for the Lambda runtime
ENTRYPOINT [ "./instance-scheduler" ]
