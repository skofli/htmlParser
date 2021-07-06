# Html Parser

Service that can parse comments on habr. Using redis to save md5 hash of comment to avoid repeat. Saving name and comment to mongo

+ Backend: Go(goquery)
+ Database: mongo, redis

## How to start

1. Start databases with `docker-compose up --build`
2. Run main.go
3. Done
