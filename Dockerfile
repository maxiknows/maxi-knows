#Multistage:
#Builder:
#During build = RUN
# Use Go 1.27.1 to build the application:
FROM golang:1.27.1 AS builder

# Set /app as the working directory in the image:
WORKDIR /app

#Copy dependencies files:
COPY go/go.mod go/go.sum ./

#Download the dependencies from go.mod and go.sum:
RUN go mod download

#Copy the GO src code into the build stage:
COPY go/cmd ./cmd
COPY go/internal ./internal

#Compile the application = build and make the executable:
RUN go build -o maxi-knows ./cmd/maxi-knows




#runtime:
#Use a small Debian image to run the compiled app:
FROM debian:bookworm-slim

#Set /app/go as the working dir:
WORKDIR /app/go

#Copy the compiled executable from the builder stage:
COPY --from=builder /app/maxi-knows ./maxi-knows


#Copy templates and static files needed:
COPY go/templates ./templates
COPY go/static ./static



#Port:
EXPOSE 8080


#After build / when the container starts = CMD
#Start the maxi-knows app when the container starts:
CMD ["./maxi-knows"]