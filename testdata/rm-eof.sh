#!/bin/bash

ls *{in,out}.sh | while read FILE; do
    perl -pe 'chomp if eof' "$FILE" >tmpfile
    mv tmpfile "$FILE"
done
