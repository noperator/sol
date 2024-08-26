cat file.json | jq '{ 
        one: .two, 
        three: .four, 
        five: .six
    }'