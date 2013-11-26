#!/bin/bash

sleep 3

echo -n "FAIL with args: '"
for arg in "$@"; do
	echo -n " ${arg} "
done
echo "'"

false
