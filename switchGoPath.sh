#!/bin/sh

# example:
# ./switchGoPath.sh github.com/elastic/beats github.com/raboof/beats httpOutput
echo "Changing $1 to point to $2 branch $3"

cd $GOPATH/src/$1
git remote add build_remote https://$2
git fetch --all
git checkout remotes/build_remote/$3
