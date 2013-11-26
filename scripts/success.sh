#!/bin/bash

sleep 3

echo -n "SUCCESS with args: '"
for arg in "$@"; do
	echo -n " ${arg} "
done
echo "'"

true
