#! /bin/bash

verbose=false
release=''
push=false

while getopts 'r:pv' flag; do
  case "${flag}" in
    r) release="${OPTARG}" ;;
    p) push=true ;;
    v) verbose=true ;;
    *) error "Unexpected option ${flag}" ;;
  esac
done

[[ -z "$release" ]] && { echo "You must give supply the 'release' '-r' argument" ; exit 1; }
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $DIR
go build .
docker build --rm -t registry.golang.scot/liberty:${release} -f $DIR/Dockerfile $DIR

if [ "$push" = true ] ; then
	docker push registry.golang.scot/liberty:${release}
fi
