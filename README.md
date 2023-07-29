# Crypto Vote API

This is a simple RESTful API implementation for managing cryptocurrencies using MySQL as the database, by which allows users to perform various operations, such as fetching all the most well known cryptocurrencies available, getting a cryptocurrency by ID, creating a new cryptocurrency, upvoting or downvoting a cryptocurrency, and deleting a cryptocurrency.

## Table of Contents
- [Components](#components)
- [Database Schema](#database-schema)
- [Endpoints Specification](#endpoints-specification)

## Components

Here's a brief summary of the main components in the code:

- **go.sum and go.mod**: These files specify the dependencies of the project along with their versions.

- **main.go**: This is the main entry point of the application. It sets up the server, initializes the database, and registers the API endpoints for handling different requests.

- **database.go**: This file contains the logic to initialize the database connection. It uses the [Go MySQL Driver](https://github.com/go-sql-driver/mysql) package to connect to a MySQL database.

- **crypto_currency_model.go**: This file defines the CryptoCurrency struct, which represents the structure of a cryptocurrency entry.

- **crypto_currency_service.go**: This file contains the main business logic for handling various API requests related to cryptocurrencies. It implements the CRUD (Create, Read, Update, Delete) operations for cryptocurrencies and interacts with the database to perform these operations.

- **crypto_currency_service_test.go**: This file contains unit tests for the CryptoCurrencyService methods. It uses the [Go SQLmock](https://github.com/DATA-DOG/go-sqlmock) package to mock the database and test the service's functionality.

## Database Schema

The MySQL database schema for the Crypto Vote API is as follows:

```
Field        | Type         | Null | Key | Default          | Extra
-------------------------------------------------------------------------
id           | int          | NO   | PRI | NULL             | auto_increment
name         | varchar(255) | NO   |     | NULL             |
up_vote      | int          | YES  |     | 0                |
down_vote    | int          | YES  |     | 0                |
total_votes  | int          | YES  |     | 0                |
```

## Endpoints specification

Below are the available endpoints and their functionalities:

### Get All Crypto Currencies

- Endpoint: `GET /v1/cryptovote`

- Description: This endpoint returns a list of all registered cryptocurrencies along with their voting statistics.

- Response: The response will be a JSON array containing objects representing each cryptocurrency and its properties (ID, name, up votes, down votes, and total votes).

### Get Crypto Currency by ID

- Endpoint: `GET /v1/cryptovote/{id}`

- Description: This endpoint retrieves a specific cryptocurrency by its unique ID.

- Response: The response will be a JSON object representing the cryptocurrency with the given ID, along with its properties (ID, name, up votes, down votes, and total votes).

### Create Crypto Currency

- Endpoint: `POST /v1/cryptovote`

- Description: This endpoint allows you to create a new cryptocurrency entry in the database.

- Request Body: The request should contain a JSON object representing the cryptocurrency to be created. The only required field is the name.

- Response: The response will be a JSON object representing the newly created cryptocurrency, including its automatically assigned ID.

### Up Vote Crypto Currency

- Endpoint: `PUT /v1/cryptovote/{id}/upvote`

- Description: This endpoint lets you cast an upvote for a specific cryptocurrency.

- Response: The response will be a JSON object representing the cryptocurrency with the updated voting statistics after the upvote.

### Down Vote Crypto Currency

- Endpoint: `PUT /v1/cryptovote/{id}/downvote`

- Description: This endpoint lets you cast a downvote for a specific cryptocurrency.

- Response: The response will be a JSON object representing the cryptocurrency with the updated voting statistics after the downvote.

### Delete Crypto Currency

- Endpoint: `DELETE /v1/cryptovote/{id}`

- Description: This endpoint allows you to delete a cryptocurrency from the database based on its ID.

- Response: If the cryptocurrency is successfully deleted, the response will have a status code of 204 (No Content) with an empty body.

### API Usage

To use these endpoints, you can make HTTP requests to the server hosting the crypto-vote application. 

You can interact with it using a tool for testing APIs, such as [Postman](https://www.postman.com/) or [cURL](https://curl.se/).

Here are some examples of using `curl` to interact with the endpoints:

- **Get All Crypto Currencies:**

```bash
curl -X GET http://localhost:8080/v1/cryptovote
```

- **Get Crypto Currency by ID** 

Replace {id} with the desired cryptocurrency ID:

```bash
curl -X GET http://localhost:8080/v1/cryptovote/{id}
```

- **Create Crypto Currency**

Replace {"name": "Bitcoin"} with the desired cryptocurrency data:

```bash
curl -X POST -H "Content-Type: application/json" -d '{"name": "Bitcoin"}' http://localhost:8080/v1/cryptovote
```

- **Up Vote Crypto Currency**

Replace {id} with the desired cryptocurrency ID:

```bash
curl -X PUT http://localhost:8080/v1/cryptovote/{id}/upvote
```

- **Down Vote Crypto Currency**

Replace {id} with the desired cryptocurrency ID:

```bash
curl -X PUT http://localhost:8080/v1/cryptovote/{id}/downvote
```

- **Delete Crypto Currency**

Replace {id} with the desired cryptocurrency ID:

```bash
curl -X DELETE http://localhost:8080/v1/cryptovote/{id}
```

Remember to replace `localhost:8080` with the actual address and port where your server is running. Additionally, for endpoints that require a request body (e.g., creating a cryptocurrency), make sure to provide valid JSON data in the `-d` parameter.