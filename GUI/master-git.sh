#!/bin/bash

# Ensure we are in a Git repository
if ! git rev-parse --is-inside-work-tree > /dev/null 2>&1; then
  echo "Error: This is not a Git repository."
  exit 1
fi

# Checkout the master branch
git checkout master

# Create a new orphan branch
git checkout --orphan temp-branch

# Add all files to the new branch
git add -A

# Commit the changes
git commit -m "Initial commit"

# Delete the old master branch
git branch -D master

# Rename the new branch to master
git branch -m master

# Force push the changes to the remote repository
echo "WARNING: This will overwrite the remote master branch. Proceed? (y/n)"
read -r response
if [[ "$response" =~ ^[Yy]$ ]]; then
  git push -f origin master
  echo "Master branch has been reset and pushed to the remote repository."
else
  echo "Operation canceled. The local master branch has been reset, but the remote was not updated."
fi