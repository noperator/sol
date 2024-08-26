find /private/etc -type f |
    xargs -I {} sh -c 'cat {} |
        wc -l'