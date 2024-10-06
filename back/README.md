# todo-app-backend

## Launch the app 
```
APP_ENV=T T_KEY="e9051112-b57f-4d4c-80c4-2940b9e0e63b" T_PORT=9999 T_LEVEL=debug  go run cmd/main.go
```

## Log-in 
```
$ curl -s  localhost:9999/api/v1/auth/login -X POST -d '{"email":"test@gmail.com", "hash":"deadbeef"}' -H 'content-type: application/json'  | jq 
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3RAZ21haWwuY29tIiwiY3JlYXRlZF9hdCI6IjIwMjQtMTAtMDZUMDI6MDU6MzguNTEyMzQzMDkxKzAxOjAwIiwic291cmNlX2FkZHJlc3MiOiIxMjcuMC4wLjEiLCJ1c2VyX2FnZW50IjoiY3VybC83LjgxLjAiLCJyb2xlIjoiYWRtaW4ifQ.J6Oe8edfEQlJVBmJMzBlMWQHdoriZY91bfFzPuJTGOM"
}

TOKEN=$(curl -s  localhost:9999/api/v1/auth/login -X POST -d '{"email":"test@gmail.com", "hash":"deadbeef"}' -H 'content-type: application/json'  | jq)
```

## Verify content of token 
```
$ echo $TOKEN | jq .token -r | cut -d '.' -f 2 | base64 -d  | jq 
base64: invalid input
{
  "email": "test@gmail.com",
  "created_at": "2024-10-06T02:06:35.535496759+01:00",
  "source_address": "127.0.0.1",
  "user_agent": "curl/7.81.0",
  "role": "admin"
}
```

## Fail to access protected area 
```
$ curl -s localhost:9999/api/v1/admin/mail/list
{"message":"Unauthorized"}
```

## Access with token 
```
$ curl -s localhost:9999/api/v1/admin/mail/list -H "X-AUTH-TOKEN: $(echo $TOKEN | jq -r .token)" | jq 
{
  "mails": [
    {
      "subject": "hello",
      "link": "http://as.com",
      "to": "ptil"
    }
  ]
}
```

