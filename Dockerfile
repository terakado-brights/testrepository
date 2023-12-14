# Start with a base image with Go installed
FROM golang:1.23-alpine

# Install Python 3, git, and other dependencies
RUN apk update && apk add --no-cache python3 py3-pip py3-virtualenv git

# Set the working directory
WORKDIR /app

# Create a virtual environment
RUN python3 -m venv /app/venv

# Activate the virtual environment and install recon-ng from GitHub
RUN . /app/venv/bin/activate && git clone https://github.com/lanmaster53/recon-ng.git /app/recon-ng && \
    cd /app/recon-ng && pip install -r REQUIREMENTS

# Set the virtual environment path as an environment variable
ENV PATH="/app/venv/bin:$PATH"
ENV PATH="/app/venv/bin:/app/recon-ng:$PATH"


# Copy the application code
COPY . .

# Build the Go application
RUN go mod tidy
RUN go build -o main .

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./main"]
