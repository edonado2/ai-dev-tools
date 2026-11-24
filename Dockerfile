# Dockerfile for GitHub Action
FROM gcr.io/distroless/static-debian12
COPY ai-dev /ai-dev
ENTRYPOINT ["/ai-dev"]
CMD ["review"]