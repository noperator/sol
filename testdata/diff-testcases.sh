#!/bin/bash

cat \
    <(
        cat ../sol_test.go |
            grep '"testdata/' |
            grep -E '\{.*\},' |
            tr -d '"{}' |
            tr , '\n' |
            sed -E 's/^[[:space:]]+//' |
            grep '^testdata' |
            sed -E 's|^testdata/||' |
            sort
    ) \
    <(
        ls *{in,out}.sh |
            sort
    ) |
    sort |
    uniq -c |
    sort -rn |
    grep -v ' 2 '
