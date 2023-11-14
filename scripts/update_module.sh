#!/bin/bash

# Step 1: Remove .git directory
rm -rf .git

# Step 2: Search and replace "github.com/Lukman-e/try-gost" with "github.com/YourUsername/YourRepoName"
find . -type f -exec sed -i 's/github\.com\/Lukmanern\/gost/github\.com\/Lukman-e\/try-gost/g' {} +

echo "Finish! .git directory removed and search/replace completed."
