# use official Golang image
FROM golang:1.21.3-alpine3.18

# set working directory 
WORKDIR /app

# Copy the source code 
COPY . .

# Download and install the dependencies 
RUN go get -d -v ./...

# Build the go app
RUN go build -o astro-service .

#EXPOSE the port
EXPOSE 8000

# Run th excutable 
CMD ["./astro-service"]