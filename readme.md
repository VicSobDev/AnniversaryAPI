
# AnniversaryAPI

## Overview

AnniversaryAPI is the backend component of the Anniversary project, a personal photo album developed as a Valentine's gift. It is designed to manage photo uploads, secure authentication, and ensure that photos can only be viewed on special days like the couple's dating anniversary or Valentine's Day. The backend is built with Go (Golang) and integrates with Grafana and Prometheus for monitoring, and gin for handling HTTP requests efficiently.

## Features

- **RESTful API:** Facilitates photo uploads and retrievals with a focus on security and efficiency.
- **Secure Authentication:** Utilizes JWT for secure authentication, ensuring that only authorized users can access the application.
- **Monitoring with Grafana and Prometheus:** Offers insights into the application's performance and usage metrics.
- **Efficient HTTP Requests:** Leverages the gin framework for fast and lightweight HTTP request handling.
- **SQLite Database:** Utilizes SQLite for lightweight and reliable data storage, making it easy to manage user data and photo metadata.

## Technologies

- **Go (Golang):** For robust backend logic and performance.
- **Grafana & Prometheus:** For application monitoring and performance metrics.
- **Gin:** A HTTP web framework written in Go for building efficient and scalable RESTful APIs.
- **SQLite:** A lightweight, disk-based database that does not require a separate server process.

## Setup and Installation

### Prerequisites

Ensure you have the following installed on your system before you begin:

- Git
- Docker & Docker Compose
- Make (optional)
- Insomnia (for API documentation viewing)

### Installation Steps

1. **Clone the Backend Repository:**
   ```bash
   git clone https://github.com/VicSobDev/AnniversaryAPI
   ```

2. **Navigate to the Project Directory:**
   ```bash
   cd AnniversaryAPI
   ```

3. **Environment Setup:**
   Before building or running the application, set up the `.env` file with the required environment variables:
   - `JWT_KEY`
   - `GF_SECURITY_ADMIN_PASSWORD`
   - `PROMETHEUS_KEY`
   - `API_KEY`

4. **Create a Key File for Prometheus:**
   Within the `prometheus` folder, create a file named `key` containing the `API_KEY` for accessing Prometheus metrics.

5. **Run the Backend:**
   You can start the backend server using the Makefile or by building the Docker Compose file:
   - Using the Makefile:
     ```bash
     make run
     ```
   - Using Docker Compose:
     ```bash
     docker-compose up --build
     ```

### Viewing API Documentation

- To view the API documentation, install Insomnia and import the `insomnia docs.json` file provided in the project directory.

## Connecting Backend with Frontend

Make sure to configure the frontend application to communicate with this backend API by setting the appropriate API service configurations in the React application.