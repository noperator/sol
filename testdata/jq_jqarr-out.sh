cat file.json | jq '[
        .one, 
        .two, 
        .three
    ]'