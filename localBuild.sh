#!/bin/bash
set -e

./glide.sh
cd floworch
if [ -e floworch ];
then 
	rm -f floworch
fi
go build -v .
cd ../kuborch
if [ -e kuborch ];
then 
	rm -f kuborch
fi
go build -v .
