#!/bin/bash

# Fix context parameters in storage test files
echo "Fixing context parameters in storage test files..."

# Fix file_storage_test.go
sed -i '' 's/storage\.CreateTask(\([^)]*\))/storage.CreateTask(context.Background(), \1)/g' internal/storage/file_storage_test.go
sed -i '' 's/storage\.GetTask(\([^)]*\))/storage.GetTask(context.Background(), \1)/g' internal/storage/file_storage_test.go
sed -i '' 's/storage\.UpdateTask(\([^)]*\))/storage.UpdateTask(context.Background(), \1)/g' internal/storage/file_storage_test.go
sed -i '' 's/storage\.DeleteTask(\([^)]*\))/storage.DeleteTask(context.Background(), \1)/g' internal/storage/file_storage_test.go
sed -i '' 's/storage\.ListTasks(\([^)]*\))/storage.ListTasks(context.Background(), \1)/g' internal/storage/file_storage_test.go
sed -i '' 's/storage\.GetTaskByDisplayID(\([^)]*\))/storage.GetTaskByDisplayID(context.Background(), \1)/g' internal/storage/file_storage_test.go
sed -i '' 's/storage\.GetProject(\([^)]*\))/storage.GetProject(context.Background(), \1)/g' internal/storage/file_storage_test.go
sed -i '' 's/storage\.CreateProject(\([^)]*\))/storage.CreateProject(context.Background(), \1)/g' internal/storage/file_storage_test.go
sed -i '' 's/storage\.UpdateProject(\([^)]*\))/storage.UpdateProject(context.Background(), \1)/g' internal/storage/file_storage_test.go
sed -i '' 's/storage\.DeleteProject(\([^)]*\))/storage.DeleteProject(context.Background(), \1)/g' internal/storage/file_storage_test.go
sed -i '' 's/storage\.ListProjects(\([^)]*\))/storage.ListProjects(context.Background(), \1)/g' internal/storage/file_storage_test.go
sed -i '' 's/storage\.GetTaskChildren(\([^)]*\))/storage.GetTaskChildren(context.Background(), \1)/g' internal/storage/file_storage_test.go
sed -i '' 's/storage\.GetTaskHierarchy(\([^)]*\))/storage.GetTaskHierarchy(context.Background(), \1)/g' internal/storage/file_storage_test.go
sed -i '' 's/storage\.GetNextDisplayID(\([^)]*\))/storage.GetNextDisplayID(context.Background(), \1)/g' internal/storage/file_storage_test.go
sed -i '' 's/storage\.GetProjectByDisplayPrefix(\([^)]*\))/storage.GetProjectByDisplayPrefix(context.Background(), \1)/g' internal/storage/file_storage_test.go
sed -i '' 's/storage\.GetProjectByName(\([^)]*\))/storage.GetProjectByName(context.Background(), \1)/g' internal/storage/file_storage_test.go
sed -i '' 's/storage\.TaskExists(\([^)]*\))/storage.TaskExists(context.Background(), \1)/g' internal/storage/file_storage_test.go
sed -i '' 's/storage\.ProjectExists(\([^)]*\))/storage.ProjectExists(context.Background(), \1)/g' internal/storage/file_storage_test.go

# Fix postgres_storage_test.go
sed -i '' 's/storage\.CreateTask(\([^)]*\))/storage.CreateTask(context.Background(), \1)/g' internal/storage/postgres_storage_test.go
sed -i '' 's/storage\.GetTask(\([^)]*\))/storage.GetTask(context.Background(), \1)/g' internal/storage/postgres_storage_test.go
sed -i '' 's/storage\.UpdateTask(\([^)]*\))/storage.UpdateTask(context.Background(), \1)/g' internal/storage/postgres_storage_test.go
sed -i '' 's/storage\.DeleteTask(\([^)]*\))/storage.DeleteTask(context.Background(), \1)/g' internal/storage/postgres_storage_test.go
sed -i '' 's/storage\.ListTasks(\([^)]*\))/storage.ListTasks(context.Background(), \1)/g' internal/storage/postgres_storage_test.go
sed -i '' 's/storage\.GetTaskByDisplayID(\([^)]*\))/storage.GetTaskByDisplayID(context.Background(), \1)/g' internal/storage/postgres_storage_test.go
sed -i '' 's/storage\.GetProject(\([^)]*\))/storage.GetProject(context.Background(), \1)/g' internal/storage/postgres_storage_test.go
sed -i '' 's/storage\.CreateProject(\([^)]*\))/storage.CreateProject(context.Background(), \1)/g' internal/storage/postgres_storage_test.go
sed -i '' 's/storage\.UpdateProject(\([^)]*\))/storage.UpdateProject(context.Background(), \1)/g' internal/storage/postgres_storage_test.go
sed -i '' 's/storage\.DeleteProject(\([^)]*\))/storage.DeleteProject(context.Background(), \1)/g' internal/storage/postgres_storage_test.go
sed -i '' 's/storage\.ListProjects(\([^)]*\))/storage.ListProjects(context.Background(), \1)/g' internal/storage/postgres_storage_test.go
sed -i '' 's/storage\.GetTaskChildren(\([^)]*\))/storage.GetTaskChildren(context.Background(), \1)/g' internal/storage/postgres_storage_test.go
sed -i '' 's/storage\.GetTaskHierarchy(\([^)]*\))/storage.GetTaskHierarchy(context.Background(), \1)/g' internal/storage/postgres_storage_test.go
sed -i '' 's/storage\.GetNextDisplayID(\([^)]*\))/storage.GetNextDisplayID(context.Background(), \1)/g' internal/storage/postgres_storage_test.go
sed -i '' 's/storage\.GetProjectByDisplayPrefix(\([^)]*\))/storage.GetProjectByDisplayPrefix(context.Background(), \1)/g' internal/storage/postgres_storage_test.go
sed -i '' 's/storage\.GetProjectByName(\([^)]*\))/storage.GetProjectByName(context.Background(), \1)/g' internal/storage/postgres_storage_test.go
sed -i '' 's/storage\.TaskExists(\([^)]*\))/storage.TaskExists(context.Background(), \1)/g' internal/storage/postgres_storage_test.go
sed -i '' 's/storage\.ProjectExists(\([^)]*\))/storage.ProjectExists(context.Background(), \1)/g' internal/storage/postgres_storage_test.go

echo "Done!"
