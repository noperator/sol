echo \
    -n 'this string with a few words' |
    wc -l | sed -E 's/^ +//'