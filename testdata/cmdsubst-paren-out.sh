echo "Date: \
$(
    date -I
)" | tr -d '-' | wc -l