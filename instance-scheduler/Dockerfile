# Stage 1: Build the Go binary
FROM golang:1.23 as builder

# Copy the rest of the source code
COPY . .

# Build the Go binary
RUN CGO_ENABLED=0 go build -o /instance-scheduler .

# Stage 2: Create the final image
FROM public.ecr.aws/lambda/provided:al2

# Copy the binary from the builder stage
COPY --from=builder /instance-scheduler ${LAMBDA_TASK_ROOT}

# Set the CMD to the binary
CMD [ "instance-scheduler" ]
