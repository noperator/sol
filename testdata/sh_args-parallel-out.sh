cat urls.txt | parallel \
    -j 4 \
    --res out.res 'curl \
        -s \
        -v \
        -o /dev/null {}'