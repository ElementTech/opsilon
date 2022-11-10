#!/usr/bin/env bash

set -euo pipefail

# read -rp "GitHub Username: " user
# read -rp "Projectname: " projectname
user="jatalocks"
projectname="opsilon"
# git clone git@github.com:jatalocks/opsilon.git "$projectname"
# cd "$projectname"
# rm -rf .git
find . -type f -exec sed -i '' -e "s/opsilon/$projectname/g" {} +
find . -type f -exec sed -i '' -e "s/jatalocks/$user/g" {} +
# git init
# git add .
# git commit -m "initial commit"
# git remote add origin "git@github.com:$user/$projectname.git"

echo "template successfully installed."
