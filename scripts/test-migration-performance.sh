#!/bin/bash

# Migration Performance Test Script
# This script tests migration performance with large datasets

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PERF_DB_NAME="projectflow_perf_test"
PERF_DB_USER="projectflow_perf"
PERF_DB_PASSWORD="perf_password"
DB_HOST="localhost"
DB_PORT="5432"

# Test sizes
SMALL_DATASET=1000
MEDIUM_DATASET=10000
LARGE_DATASET=100000

echo -e "${BLUE}ProjectFlow Migration Performance Test${NC}"
echo "======================================"

# Function to create performance test database
create_perf_database() {
    echo -e "${YELLOW}Setting up performance test database...${NC}"

    # Drop existing database if it exists
    psql -h $DB_HOST -p $DB_PORT -U postgres -c "DROP DATABASE IF EXISTS $PERF_DB_NAME;" 2>/dev/null || true
    psql -h $DB_HOST -p $DB_PORT -U postgres -c "DROP USER IF EXISTS $PERF_DB_USER;" 2>/dev/null || true

    # Create user and database
    psql -h $DB_HOST -p $DB_PORT -U postgres -c "CREATE USER $PERF_DB_USER WITH PASSWORD '$PERF_DB_PASSWORD';"
    psql -h $DB_HOST -p $DB_PORT -U postgres -c "CREATE DATABASE $PERF_DB_NAME OWNER $PERF_DB_USER;"
    psql -h $DB_HOST -p $DB_PORT -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE $PERF_DB_NAME TO $PERF_DB_USER;"

    echo -e "${GREEN}✅ Performance test database created${NC}"
}

# Function to create test data
create_test_data() {
    local dataset_size=$1
    local dataset_name=$2

    echo -e "${YELLOW}Creating $dataset_name dataset ($dataset_size records)...${NC}"

    # Set database environment variables
    export DB_HOST=$DB_HOST
    export DB_PORT=$DB_PORT
    export DB_NAME=$PERF_DB_NAME
    export DB_USER=$PERF_DB_USER
    export DB_PASSWORD=$PERF_DB_PASSWORD
    export DB_SSL_MODE="disable"

    # Initialize migration table
    ./build/migrate init

    # Create initial schema without tenant support
    psql -h $DB_HOST -p $DB_PORT -U $PERF_DB_USER -d $PERF_DB_NAME << EOF
CREATE TABLE IF NOT EXISTS projects (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    display_prefix VARCHAR(10) NOT NULL,
    task_counter INTEGER NOT NULL DEFAULT 0,
    settings JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tasks (
    id VARCHAR(36) PRIMARY KEY,
    display_id VARCHAR(50) UNIQUE,
    project_id VARCHAR(36),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'todo',
    priority VARCHAR(20) NOT NULL DEFAULT 'medium',
    type VARCHAR(20) NOT NULL DEFAULT 'task',
    parent_id VARCHAR(36),
    children JSONB DEFAULT '[]'::jsonb,
    started_at TIMESTAMPTZ,
    due_date TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
EOF

    # Generate test data using a Go script
    go run -<<'EOGO'
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func main() {
	datasetSize, _ := strconv.Atoi(os.Args[1])
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"))

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create projects (10% of dataset size)
	projectCount := max(1, datasetSize/10)
	fmt.Printf("Creating %d projects...\n", projectCount)

	tx, _ := db.Begin()
	for i := 0; i < projectCount; i++ {
		_, err := tx.Exec(`
			INSERT INTO projects (id, name, display_prefix, description, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, NOW(), NOW())
		`, uuid.New().String(), fmt.Sprintf("Perf Project %d", i), fmt.Sprintf("PP%d", i), fmt.Sprintf("Performance test project %d", i))
		if err != nil {
			log.Printf("Error inserting project %d: %v", i, err)
		}
	}
	tx.Commit()

	// Get project IDs
	rows, _ := db.Query("SELECT id FROM projects")
	var projectIDs []string
	for rows.Next() {
		var id string
		rows.Scan(&id)
		projectIDs = append(projectIDs, id)
	}
	rows.Close()

	// Create tasks
	fmt.Printf("Creating %d tasks...\n", datasetSize)
	tx, _ = db.Begin()
	for i := 0; i < datasetSize; i++ {
		projectID := projectIDs[i%len(projectIDs)]
		_, err := tx.Exec(`
			INSERT INTO tasks (id, display_id, project_id, title, description, status, priority, type, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		`, uuid.New().String(), fmt.Sprintf("PERF-%d", i), projectID, 
		   fmt.Sprintf("Performance Task %d", i), fmt.Sprintf("Performance test task %d", i),
		   []string{"todo", "in_progress", "done", "blocked"}[i%4],
		   []string{"low", "medium", "high", "critical"}[i%4],
		   []string{"task", "story", "epic", "subtask"}[i%4])
		if err != nil {
			log.Printf("Error inserting task %d: %v", i, err)
		}

		if i%1000 == 0 {
			tx.Commit()
			tx, _ = db.Begin()
		}
	}
	tx.Commit()

	fmt.Printf("Test data creation completed: %d projects, %d tasks\n", projectCount, datasetSize)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
EOGO
    $dataset_size

    echo -e "${GREEN}✅ $dataset_name dataset created${NC}"
}

# Function to run migration performance test
test_migration_performance() {
    local dataset_name=$1

    echo -e "${YELLOW}Testing migration performance on $dataset_name dataset...${NC}"

    # Record start time
    start_time=$(date +%s.%N)

    # Run migrations
    ./build/migrate up

    # Record end time
    end_time=$(date +%s.%N)

    # Calculate duration
    duration=$(echo "$end_time - $start_time" | bc)

    echo -e "${GREEN}✅ Migration completed in ${duration}s${NC}"

    # Count records after migration
    local project_count=$(psql -h $DB_HOST -p $DB_PORT -U $PERF_DB_USER -d $PERF_DB_NAME -t -c "SELECT COUNT(*) FROM projects WHERE tenant_id IS NOT NULL;")
    local task_count=$(psql -h $DB_HOST -p $DB_PORT -U $PERF_DB_USER -d $PERF_DB_NAME -t -c "SELECT COUNT(*) FROM tasks WHERE tenant_id IS NOT NULL;")

    echo -e "${BLUE}📊 Results:${NC}"
    echo "  - Migration time: ${duration}s"
    echo "  - Projects migrated: $project_count"
    echo "  - Tasks migrated: $task_count"
    echo "  - Records per second: $(echo "scale=2; ($project_count + $task_count) / $duration" | bc)"

    # Store results for summary
    echo "$dataset_name,$duration,$project_count,$task_count" >> /tmp/perf_results.csv
}

# Function to cleanup
cleanup() {
    psql -h $DB_HOST -p $DB_PORT -U postgres -c "DROP DATABASE IF EXISTS $PERF_DB_NAME;" 2>/dev/null || true
    psql -h $DB_HOST -p $DB_PORT -U postgres -c "DROP USER IF EXISTS $PERF_DB_USER;" 2>/dev/null || true
}

# Function to show summary
show_summary() {
    echo -e "${BLUE}📈 Performance Test Summary${NC}"
    echo "==========================="
    echo ""
    printf "%-15s %-12s %-10s %-8s %-12s\n" "Dataset" "Time (s)" "Projects" "Tasks" "Records/s"
    echo "------------------------------------------------------------"
    
    while IFS=',' read -r dataset time projects tasks; do
        records_per_sec=$(echo "scale=2; ($projects + $tasks) / $time" | bc)
        printf "%-15s %-12s %-10s %-8s %-12s\n" "$dataset" "$time" "$projects" "$tasks" "$records_per_sec"
    done < /tmp/perf_results.csv
}

# Main execution
main() {
    # Initialize results file
    echo "Dataset,Time,Projects,Tasks" > /tmp/perf_results.csv

    # Check if bc is available for calculations
    if ! command -v bc &> /dev/null; then
        echo -e "${RED}❌ bc calculator not found. Please install bc for performance calculations.${NC}"
        exit 1
    fi

    # Build the project
    go build -o build/migrate ./cmd/migrate

    # Test with different dataset sizes
    for dataset_info in "Small,$SMALL_DATASET" "Medium,$MEDIUM_DATASET" "Large,$LARGE_DATASET"; do
        IFS=',' read -r dataset_name dataset_size <<< "$dataset_info"
        
        echo -e "\n${BLUE}Testing $dataset_name Dataset ($dataset_size records)${NC}"
        echo "================================================"
        
        create_perf_database
        create_test_data $dataset_size $dataset_name
        test_migration_performance $dataset_name
        cleanup
        
        sleep 2  # Brief pause between tests
    done

    show_summary

    # Clean up results file
    rm -f /tmp/perf_results.csv

    echo -e "\n${GREEN}🎉 Performance testing completed!${NC}"
}

# Set trap for cleanup
trap cleanup EXIT

# Run main function
main
