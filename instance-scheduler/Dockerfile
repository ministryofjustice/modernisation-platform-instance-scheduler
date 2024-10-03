# Stage 1: Build the Go binary
FROM golang:1.23 as builder

# Set the Go working directory
WORKDIR /app

# Copy and download dependency modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /instance-scheduler .

# Stage 2: Create the final image
FROM public.ecr.aws/lambda/provided:al2

# Copy the binary and the bootstrap script from the builder stage
COPY --from=builder /instance-scheduler ${LAMBDA_TASK_ROOT}/instance-scheduler
COPY bootstrap ${LAMBDA_TASK_ROOT}/bootstrap

# Make the bootstrap script executable
RUN chmod +x ${LAMBDA_TASK_ROOT}/bootstrap

# Set the command for the Lambda runtime
CMD [ "bootstrap" ]
