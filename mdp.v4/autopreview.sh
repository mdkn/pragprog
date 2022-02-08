#!/usr/bin/env bash

if [ "$(uname)" == 'Darwin' ]; then
  FHASH=`md5 $1`
elif [ "$(expr substr $(uname -s) 1 5)" == 'Linux' ]; then
  FHASH=`md5sum $1`
else
  echo "Your platform ($(uname -a)) is not supported."
  exit 1
fi

while true; do
  if [ "$(uname)" == 'Darwin' ]; then
    NHASH=`md5 $1`
  else
    NHASH=`md5sum $1`
  fi
  if [ "$NHASH" != "$FHASH" ]; then
    ./mdp -file $1
    FHASH=$NHASH
  fi
  sleep 5
done
