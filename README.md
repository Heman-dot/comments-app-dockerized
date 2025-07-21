# 🐳 Comments App (Dockerized)

This project provides Docker resources to deploy a basic microservices-based comments application without modifying the existing application code.

## 📦 Components

- **CommentsInteractor**: A simple frontend interface for posting and viewing comments.
- **CommentsEngine**: A backend API handling comment storage and retrieval.
- **Redis**: In-memory data store used by the backend for storing comments.

## 🚀 Quick Start

### Requirements

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)

### Steps

1. Clone the repository:

   ```bash
   git clone https://github.com/Heman-dot/comments-app-dockerized
   cd comments-app-dockerized

2. Build and run the containers:
```
docker-compose up --build
```

3. Open your browser and visit:
```
http://localhost:3000
(Port may vary depending on how CommentsInteractor is configured.)
```

4. 🧱 Project Structure
```
comments-app-dockerized/
├── docker-compose.yml
├── CommentsInteractor/
│   └── Dockerfile
├── CommentsEngine/
│   └── Dockerfile
└── README.md
```

❗ Disclaimer
This is a deployment-only project for a proof-of-concept application. The application logic is not production-ready, and no modifications were made to the original code.