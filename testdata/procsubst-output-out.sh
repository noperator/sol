tee >/dev/null \
    >(
        wc -l
    ) \
    >(
        sort
    )