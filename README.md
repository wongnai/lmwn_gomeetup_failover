
## 📌 **Introduction**
This project is designed to be a **Go Meetup Example Project**, demonstrating how to build a **scalable microservices architecture** using:
- **gRPC & HTTP servers** for communication
- **MongoDB** for persistent storage
- **RabbitMQ** for message queue handling
- **Docker** to containerize dependencies
- **Graceful shutdown handling** for fail-safe operations
- **Health Check Endpoints** to monitor the status of each service
- **Retry & Circuit Breaker Mechanism** to prevent cascading failures
- **Worker Pool & Semaphore Concept** to prevent goroutine overload

This project serves as a real-world example of implementing **high availability, failover handling, and dependency cleanup** in a Go microservices environment.

## 📌 **Setup Guide for Running the Project**

### **Prerequisites**
Ensure you have the following tools installed on your system:
- **[Docker](https://docs.docker.com/get-docker/)** - Required to run MongoDB and RabbitMQ
- **[Go](https://go.dev/dl/)** - Required for compiling and running the application
- **[Protocol Buffers (protoc)](https://grpc.io/docs/protoc-installation/)** - Required for gRPC services
- **`protoc-gen-go` & `protoc-gen-go-grpc`** - Required for generating gRPC code

### **1️⃣ Clone the Repository**
```sh
git clone https://github.com/your-repo/go-meetup-example.git
cd go-meetup-example
```

### **2️⃣ Start MongoDB & RabbitMQ Using Docker**
Ensure **Docker** is running, then start the required services:
```sh
docker-compose up -d
```
This will start:
- **MongoDB** on `localhost:27017`
- **RabbitMQ** on `localhost:5672` (management UI on `localhost:15672`)

To verify containers are running:
```sh
docker ps
```

### **3️⃣ Install Required Go Modules**
Ensure all dependencies are installed:
```sh
go mod tidy
```

### **4️⃣ Generate Protocol Buffers for gRPC**
Install protobuf dependencies:
```sh
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```
Then generate the gRPC code:
```sh
protoc --go_out=. --go-grpc_out=. proto/*.proto
```

### **5️⃣ Run the Services**
Start the required services:

#### **Run the HTTP Server**
```sh
go run cmd/http/main.go
```

#### **Run the gRPC Server**
```sh
go run cmd/grpc/main.go
```

#### **Run the Cron Job**
```sh
go run cmd/cron/main.go
```

#### **Run the RabbitMQ Consumer**
```sh
go run cmd/consumer/main.go
```

### **6️⃣ Verify Everything is Running**
- **Check HTTP Server:** Open `http://localhost:8080` and test API calls.
- **Check RabbitMQ UI:** Visit `http://localhost:15672` (user: `guest`, pass: `guest`).
- **Check MongoDB Connection:** Run `docker exec -it mongodb mongosh`.

### **7️⃣ Stop Services & Cleanup**
To gracefully stop all services:
```sh
docker-compose down
```

---

## 📌 **Why This Setup Matters?**
✅ **Ensures MongoDB & RabbitMQ run correctly before services start**  
✅ **Provides complete setup for gRPC and HTTP-based microservices**  
✅ **Includes `protoc` setup for gRPC message serialization**  
✅ **Ensures smooth onboarding for new contributors**  

Now, anyone can **clone, set up, and run the project seamlessly!** 🚀
