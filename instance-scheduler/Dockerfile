FROM public.ecr.aws/lambda/go:1 as builder

# Copy function code
COPY . .

RUN yum install -y golang
RUN go env -w GOPROXY=direct

RUN CGO_ENABLED=0 go build \
    -o /instance-scheduler .

FROM public.ecr.aws/lambda/go:1

COPY --from=builder instance-scheduler ${LAMBDA_TASK_ROOT}
# Set the CMD to your handler (could also be done as a parameter override outside of the Dockerfile)
CMD [ "instance-scheduler" ]
