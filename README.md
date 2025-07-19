# Inventory Service
Repository for Inventory Service

Product Journey:
https://varsentinel.atlassian.net/wiki/spaces/Inventiums/overview

## Using Gin Framework:

https://gin-gonic.com/en/docs/quickstart/

## SQLC:

https://docs.sqlc.dev/en/stable/index.html


## Project Structures:
```
    inventory-service/
    ├── api/                   # Gin Router Controller
    │     ├── server.go        #
    ├── config/                # Store Applications Configs
    │     ├── config.go        #
    ├── handlers/              # Handlers Controller for different API Methods
    │     └── inventory.go     #
    ├── middlewares/           # Middlewares to check foe authorized client
    |     └── authenticate.go  #
    |── models/                # Models for working with Postgresl
    |     |── migration        # DB Migration
    |     |── query.           # DB Query
    |     |── sqlc             # DB Connection
    |── routes/                # Stores Route
    |     └── routes.go        #
```
## API Routes

- List Inventory:   /inventory
- Get Inventory:    /inventory/:id
- Create Inventory: /inventory/:id
- Update Inventory: /inventory/:id
- Delete Inventory: /inventory/:id
