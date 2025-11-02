#!/bin/bash

LINE="##############################\n"
LINE="$LINE$URL\n"
LINE="$LINE---------- Request Header ----------\n"
LINE="$LINE$REQ_HEADER\n"
LINE="$LINE---------- Request Body ----------\n"
LINE="$LINE$REQ_BODY\n"
LINE="$LINE---------- Response Header ----------\n"
LINE="$LINE$RES_HEADER\n"
LINE="$LINE---------- Response Body ----------"

echo -e "$LINE" >> output.txt
tee -a output.txt
