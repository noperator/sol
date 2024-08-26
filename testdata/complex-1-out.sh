cat urls.lst |
    parallel \
        -j 4 'curl \
            -sv \
            -d {"this":"that"} \
            -H "Content-Type: application/json" {} \
            2>>out.log |
            grep \
                -v "error"' |
    jq \
        -s \
        --arg key val 'map({ 
                key: $val
            } + 
            { 
                response: (.code | 
                    test("^2")), 
                selectors: [
                    ".this", 
                    ".that"
                ]
            })'