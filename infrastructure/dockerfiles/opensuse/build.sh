#! /bin/bash

verbose='false'
appenv=''
push=false
deploy=false
gover=''

while getopts 'e:pdvg:' flag; do
  case "${flag}" in
    e) appenv="${OPTARG}" ;;
    p) push=true ;;
    d) deploy=true ;;
    v) verbose='true' ;;
	g) gover="${OPTARG}" ;;
    *) error "Unexpected option ${flag}" ;;
  esac
done

[[ -z "$appenv" ]] && { echo "You must give supply the 'appenv' '-e' argument" ; exit 1; }
[[ -z "$gover" ]] && { echo "You must give supply the 'gover' '-g' argument" ; exit 1; }

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $DIR

goloc="https://storage.googleapis.com/golang/go${gover}.linux-amd64.tar.gz"
echo "Downloading Go version ${gover}..."
# now grab the Go version
cd $(mktemp -d)
curl -O $goloc
echo "Installing Go version ${gover}..."
sudo tar -C $DIR -xzf go${gover}.linux-amd64.tar.gz

# build the container image
echo "Building container..."
if [ "$appenv" = "go" ] ; then
	docker build --rm -t registry.golang.scot/go:${gover} $DIR
else
	docker build --rm -t registry.golang.scot/go:${appenv} -f $DIR/${appenv}.Dockerfile $DIR
fi

echo "Cleaning up..."
# cleanup
