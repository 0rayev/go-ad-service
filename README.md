# GoAdServer

It is a Go backend REST API designed for managing advertisements, built with the Gin framework. This project leverages Docker for containerization, Redis for caching, MySQL for database management, Prometheus for metrics collection, and OpenTelemetry for distributed tracing. 

## Table of Contents
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Usage](#usage)
- [API Endpoints](#api-endpoints)
  - [Get All Ads](#Get-All-Ads)
  - [Get Ad by ID](#Get-Ad-by-ID)
  - [Create Ad](#Create-Ad)
  - [Delete Ad](#Delete-Ad)
  - [Update Ad](#Update-Ad)
- [Database Migration](#database-migration)
- [Caching](#caching)
- [OpenTelemetry Tracing Setup](#opentelemetry-tracing-setup)
- [Prometheus Metrics](#prometheus-metrics)
- [Configuration](#configuration)


## Features

- Get All Ads: Retrieve a list of all advertisements from the database with pagination support, displaying 10 ads per page.The advertisements can be sorted in ascending or descending order based on specific fields.
- Get Ad: Retrieve advertisement data from the database based on the provided ID.
- Create Ad: Add a new advertisement to the database.
- Delete Ad: Remove an advertisement from the database based on the provided ID.
- Update Ad: Modify existing advertisement data in the database based on the provided ID.

## Prerequisites

Before running this application, ensure that you have the following prerequisites installed:

- Go: [Install Go](https://go.dev/doc/install/)
- Docker: [Install Docker](https://docs.docker.com/get-docker/)
- Docker Compose: [Install Docker Compose](https://docs.docker.com/compose/install/)

## Installation

1. Clone the repository:
  ```bash
    git clone https://github.com/0rayev/GoAdServer.git
  ```

2. Navigate to the project directory:
  ```
    cd ad-service
  ```
3. Build the Docker image:
  ```
    docker-compose up --build
  ```

## Usage
Start the Docker containers:
  ```
    docker-compose up
  ```
This will:
  - Build the Go application.
  - Start MySQL, Redis, and Prometheus services.
  - Expose the application on port 8080.
## API Endpoints

### Get All Ads

- Method: GET
- Endpoint: /ads
- Request Parameters: 
  - page: (Optional) The page number for pagination (default is 1).Must be a positive integer.
  - limit: (Optional) The number of ads to fetch per page (default is 10).Must be a positive integer.
  - sort_by: (Optional) Attribute to sort by (default is created_at).Must be one of id, title, price, created_at, is_active.
  - order: (Optional) Sorting order (asc or desc, default is asc).Must be either asc or desc.

Retrieve all ads from the database with optional pagination and sorting.

- Response:
  - 200 OK: Returns a list of ads.
    - Example response body:
      ```json
      [
        {
          "id": 1,
          "title": "Ad 1",
          "description": "This is the first ad.",
          "created_at": "2024-10-14T12:34:56Z"
        },
        {
          "id": 2,
          "title": "Ad 2",
          "description": "This is the second ad.",
          "created_at": "2024-10-14T12:36:33Z"
        }
      ]
      ```

  - 400 Bad Request: If any of the query parameters are invalid, the following error responses will be returned:
    - Example response body:
      ```json
      {
        "error": "Invalid page value. Must be a positive integer."
      }
      
      {
        "error": "Invalid sort_by value. Must be one of 'id', 'title', 'price', 'created_at', 'is_active'."
      }

      {
        "error": "Invalid order value. Must be either 'asc' or 'desc'."
      }
      ```

  - 500 Internal Server Error: If an error occurs while fetching the ads, the following response is returned:
    - Example response body:
      ```json
      {
        "error": "Failed to fetch ads"
      }
      ```

### Get Ad by ID

- Method: GET
- Endpoint: /ads/:id
- Request Parameters: id (integer), the ID of the advertisement to be retrieved.Must be a positive integer.

Retrieve a specific ad from the database by its ID.

- Response:
  - 200 OK: Returns the ad data.
    - Example response body:
      ```json
      {
        "id": 1,
        "title": "Ad Title",
        "description": "Ad description",
        "price": 99.99,
        "created_at": "2024-10-14T12:34:56Z",
        "is_active": true
      }
      ```
  - 400 Bad Request: If the provided id is not valid. The following error response will be returned:
    - Example response body:
      ```json
      {
        "error": "Invalid ID"
      }
      ```    
  - 404 Not Found: If the ad with the given id does not exist in the database.
    - Example response body:
      ```json
      {
        "error": "Ad not found"
      }
      ```
  - 500 Internal Server Error: If there is an internal error fetching the ad from the database.
    - Example response body:
      ```json
      {
        "error": "Failed to fetch ad by ID"
      }
      ```


### Create Ad

- Method: POST
- Endpoint: /ads
- Request Body: JSON payload containing the new ad's data
  - title (string, required): The title of the advertisement.Cannot be empty.
  - description (string, required): A detailed description of the advertisement.Cannot be empty.
  - price (float, required): The price of the item being advertised.Must be a positive value.
  - is_active (boolean, optional): The status of the ad (default is false).

Add new data to the database.

- Response:
  - 201 Created: Returns the created ad object, including the automatically assigned id and created_at timestamp.
    - Example Request Body:
      ```json
      {
        "title": "Brand New Laptop",
        "description": "High-performance laptop for sale. Barely used.",
        "price": 1200.00,
        "is_active": true
      }
      ```
    - Example Response body:
      ```json
      {
        "id": 1,
        "title": "Brand New Laptop",
        "description": "High-performance laptop for sale. Barely used.",
        "price": 1200.00,
        "created_at": "2024-10-14T12:34:56Z",
        "is_active": true
      }
      ```
  - 400 Bad Request: If the request body is invalid or missing required fields, the following error responses will be returned:
    - Example response body:
      ```json
      {
        "error": "Title and description are required"
      }

      {
        "error": "Price cannot be zero or negative"
      }

      {
        "error": "Invalid request body"
      }
      ```
  - 500 Internal Server Error: If there is an error creating the ad in the database.
    - Example response body:
      ```json
      {
        "error": "Failed to add ad"
      }
      ```

### Delete Ad:

- Method: DELETE
- Endpoint: /ads/:id
- Request Parameters: id (int, required), the unique ID of the advertisement to be deleted.

Remove data from the database based on the provided ID.

- Example:
  - DELETE /ads/1
- Responses:
  - 200 OK: If the ad is successfully deleted.
    - Example response body:
      ```json
      {
        "message": "Ad deleted successfully"
      }
      ```
  - 400 Bad Request: If the provided ID is invalid.
    - Example response body:
      ```json
      {
        "error": "Invalid ID"
      }
      ```
  - 404 Not Found: If the ad does not exist.
    - Example response body:
      ```json
      {
        "error": "Ad not found"
      }
      ```
  - 500 Internal Server Error: If there is an internal error while attempting to delete the ad from the database.
    - Example response body:
      ```json
      {
        "error": "Failed to delete ad"
      }
      ```
### Update Ad:

- Method: PUT
- Endpoint: /ads/:id
- Request Body: JSON payload with the updated ad details (title, description, price, is_active).
  - The fields title, description, and price are required, while is_active is optional. If a field is not provided, its current value in the database will remain unchanged.

Update an existing ad by its ID.

- Response:
  - 200 OK: If the request is successful.
    - Example Request body:
      ```json
      {
        "title": "New Ad Title",
        "description": "Updated description of the ad",
        "price": 100.00,
        "is_active": true
      }
      ```
    - Example Response:
      ```json
      {
        "message": "Ad updated"
      }
      ```
  - 400 Bad Request: If the request body or ID is invalid, if required fields are missing.
    - Example Response: 
      ```json
      {
        "error": "Invalid request body"
      }
      ```
  - 404 Not Found: If the provided ID does not exist in the database.
    - Example Response:
      ```json
      {
          "error": "Ad not found"
      }
      ```
  - 500 Internal Server Error: If there is an internal server error.
    - Example response body:
      ```json
      {
        "error": "Failed to update ad"
      }
      ```

## Database Migration

The init.sql migration file under database/migrations/ will be automatically applied to set up the necessary database schema for the MySQL database.

## Caching

Caching is implemented using Redis to improve the performance and scalability of the ad service.

- Where Caching is Used
  - GetAdByID Method:
      - When retrieving a specific ad by its ID (GetAdByID), the service first checks the cache (Redis) for the requested ad.
      - If the ad is found in the cache (cache hit), it is returned immediately, avoiding a database query.
      - If the ad is not found (cache miss), the service queries the database, retrieves the ad, and stores it in the cache for future requests.
      - The cache is set with a time-to-live (TTL) of 5 minutes, after which the cached data expires and must be fetched again from the database.

  - UpdateAd Method:
      - When an ad is updated (UpdateAd), the cache entry for the specific ad is invalidated (deleted). This ensures that outdated data is not served from the cache after an update.

  - DeleteAd Method:
        Similarly, when an ad is deleted (DeleteAd), the corresponding cache entry is removed to keep the cache consistent with the database.

## OpenTelemetry Tracing Setup

The tracing setup is defined in the tracing.go file under pkg/tracing/. In this project, the tracing system uses the stdout exporter, which prints trace information to the console in a human-readable format.

## Prometheus Metrics

Prometheus metrics are defined in the prometheus.go file under metrics/. You can access Prometheus scraping at http://localhost:9090.

## Configuration

Configuration is handled using Viper. You can update the configuration by modifying the config.yaml file.
