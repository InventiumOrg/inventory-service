#!/bin/bash

# Configuration
BASE_URL="${BASE_URL:-http://localhost/v1/inventory}"

# Number of requests to make (default: 10)
NUM_REQUESTS=${1:-10}

# Delay between requests in seconds (default: 0.5)
DELAY=${2:-0.5}

echo "Starting inventory service simulation with $NUM_REQUESTS requests..."
echo "Delay between requests: ${DELAY}s"
echo "----------------------------------------"

# Function to create an inventory item
create_inventory() {
    local name=$1
    local unit=$2
    local quantity=$3
    local measure=$4
    local category=$5
    local location=$6
    
    curl --location --request POST "${BASE_URL}/create" \
        --form "Name=${name}" \
        --form "Unit=${unit}" \
        --form "Quantity=${quantity}" \
        --form "Measure=${measure}" \
        --form "Category=${category}" \
        --form "Location=${location}" \
        --silent --show-error
}

# Function to get an inventory item
get_inventory() {
    local inventory_id=$1
    
    curl --location --request GET "${BASE_URL}/${inventory_id}" \
        --silent --show-error
}

# Function to list inventory items
list_inventory() {
    curl --location --request GET "${BASE_URL}/list" \
        --silent --show-error
}

# Function to update an inventory item
update_inventory() {
    local inventory_id=$1
    local name=$2
    local unit=$3
    local quantity=$4
    local measure=$5
    local category=$6
    local location=$7
    
    curl --location --request PUT "${BASE_URL}/${inventory_id}" \
        --form "Name=${name}" \
        --form "Unit=${unit}" \
        --form "Quantity=${quantity}" \
        --form "Measure=${measure}" \
        --form "Category=${category}" \
        --form "Location=${location}" \
        --silent --show-error
}

# Function to delete an inventory item
delete_inventory() {
    local inventory_id=$1

    curl --location --request DELETE "${BASE_URL}/${inventory_id}" \
        --silent --show-error
}

# Simulate various operations
for i in $(seq 1 $NUM_REQUESTS); do
    echo "Request $i/$NUM_REQUESTS"
    
    # Generate variation in the data
    INVENTORY_ID=$i
    NAME="Item-${i}"
    UNIT="pcs"
    QUANTITY=$((10 + ($i % 25)))
    MEASURE="unit"
    CATEGORY="category-$((($i % 5) + 1))"
    LOCATION="location-$((($i % 3) + 1))"
    
    # Perform different operations based on request number
    case $((i % 5)) in
        0)
            echo "Creating inventory: Name=${NAME}, Unit=${UNIT}, Quantity=${QUANTITY}, Measure=${MEASURE}, Category=${CATEGORY}, Location=${LOCATION}"
            create_inventory "${NAME}" "${UNIT}" "${QUANTITY}" "${MEASURE}" "${CATEGORY}" "${LOCATION}"
            ;;
        1)
            echo "Getting inventory with ID: ${INVENTORY_ID}"
            get_inventory "${INVENTORY_ID}"
            ;;
        2)
            echo "Listing inventory"
            list_inventory
            ;;
        3)
            echo "Updating inventory ID: ${INVENTORY_ID} with new Quantity: ${QUANTITY}"
            update_inventory "${INVENTORY_ID}" "${NAME}" "${UNIT}" "${QUANTITY}" "${MEASURE}" "${CATEGORY}" "${LOCATION}"
            ;;
        4)
            echo "Deleting inventory with ID: ${INVENTORY_ID}"
            delete_inventory "${INVENTORY_ID}"
            ;;
    esac
    
    echo ""
    echo "Response for request $i completed"
    
    # Add delay between requests (except for the last one)
    if [ $i -lt $NUM_REQUESTS ]; then
        echo "Waiting ${DELAY}s before next request..."
        sleep $DELAY
    fi
    
    echo "----------------------------------------"
done

echo ""
echo "Simulation completed!"
echo "Total requests made: $NUM_REQUESTS"