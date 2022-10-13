FROM public.ecr.aws/docker/library/alpine:latest

COPY container-test /
ENV PORT 80
EXPOSE $PORT

ENTRYPOINT [ "/container-test" ]
