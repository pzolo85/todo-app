# todo-app-backend

## Build / install the app 
```
$ go install -C back cmd/todo-app.go
```

## Help 
```
$ todo-app -h 
env var APP_ENV is not set. Trying to load config from flags


 ____  __  ___    __        __   ___  ___ 
(_  _)/  \(   \  /  \  ___ (  ) (  ,\(  ,\
  )( ( () )) ) )( () )(___)/__\  ) _/ ) _/
 (__) \__/(___/  \__/     (_)(_)(_)  (_)  


 todo-app is the backend of a web app for creating and sharing To-Do lists
 
 Usage:

   -c   create a new admin JWT token
  -create-token
        create a new admin JWT token
  -d duration
        duration of the admin JWT token (default 15m0s)
  -duration duration
        duration of the admin JWT token (default 15m0s)
  -e string
        email address to use in the JWT token (default "admin@localhost")
  -email string
        email address to use in the JWT token (default "admin@localhost")
  -g    create a new JWT signing key (/home/user/.todo-app.key)
  -generate
        create a new JWT signing key (/home/user/.todo-app.key)
  -h    show this help
  -help
        show this help
  -k string
        file holding the signing key for JWT (default "/home/user/.todo-app.key")
  -key-path string
        file holding the signing key for JWT (default "/home/user/.todo-app.key")
```

## Generate a new JWT sign key 
```
$ todo-app -g 
env var APP_ENV is not set. Trying to load config from flags

new key generated: /home/user/.todo-app.key
```

## Generate an admin JWT token 
```
$ export ADMIN_TOKEN=$(todo-app -c -d 12h -e test@localhost)
env var APP_ENV is not set. Trying to load config from flags

```

## Check token contents
```
$ echo $ADMIN_TOKEN  | cut -d . -f 2 | base64 -d | jq 
{
  "email": "test@localhost",
  "created_at": "2024-10-06T23:43:31.969480195+01:00",
  "expires_at": "2024-10-07T11:43:31.969480195+01:00",
  "is_admin": true,
  "source_address": "127.0.0.1",
  "user_agent": "curl",
  "claim_id": "6426ce7d-f9a9-4046-a0c8-382a30a42480"
}
```

## Launch app with log level set to debug
```
$ APP_ENV=TD TD_LEVEL=debug todo-app 
```

## Create new user
```
$ curl -sH 'content-type:application/json' localhost:7777/api/v1/user/create -d '{"email":"jon@test.com", "salt":"abc123", "hashed_pass":"deadbeef"}' | jq                                                                  
{
  "email": "jon@test.com",
  "pass_hash": "deadbeef",
  "salt": "abc123",
  "role": "user",
  "created_at": "2024-10-06T23:56:49.888288772+01:00"
}
```

## Log-in with new user
```
$ USER_TOKEN=$(curl -sH 'content-type:application/json' localhost:7777/api/v1/auth/login -d '{"email":"jon@test.com", "hash":"deadbeef"}' | jq .token -r )
```

## Check content of token
```
$ echo $USER_TOKEN | cut -d . -f2 | base64 -d  | jq
{
  "email": "jon@test.com",
  "created_at": "2024-10-07T00:32:22.807572033+01:00",
  "expires_at": "0001-01-01T00:00:00Z",
  "is_admin": false,
  "source_address": "127.0.0.1",
  "user_agent": "curl/7.81.0",
  "claim_id": "817a4ed0-a0af-4d08-bc44-69d564d87e0c"
}
```

## Try to access area for users that validated their email 
```
$ curl localhost:7777/api/v1/user/info -sH "x-auth-token: $USER_TOKEN"  | jq 
{
  "message": "please validate your account"
}
```

## List emails waiting validation with admin account 
```
$ curl localhost:7777/api/v1/admin/mail/list -sH "x-auth-token: $ADMIN_TOKEN"  | jq 
{
  "mails": [
    {
      "subject": "verify your email",
      "link": "http://127.0.0.1:7777/api/v1/user/validate?email=jon@test.com&challenge=262f0a7f-db92-49fa-9879-a6aee8449a16",
      "to": "jon@test.com"
    }
  ]
}
```

## Verify email 
```
$ curl -s "http://127.0.0.1:7777/api/v1/user/validate?email=jon@test.com&challenge=262f0a7f-db92-49fa-9879-a6aee8449a16" | jq 
{
  "email": "jon@test.com",
  "pass_hash": "deadbeef",
  "salt": "abc123",
  "role": "user",
  "created_at": "2024-10-07T00:59:22.976387731+01:00",
  "valid_email": true,
  "active_jwt": [
    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImpvbkB0ZXN0LmNvbSIsImNyZWF0ZWRfYXQiOiIyMDI0LTEwLTA3VDAwOjU5OjI1LjQwOTIxNTc3MiswMTowMCIsImV4cGlyZXNfYXQiOiIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsImlzX2FkbWluIjpmYWxzZSwic291cmNlX2FkZHJlc3MiOiIxMjcuMC4wLjEiLCJ1c2VyX2FnZW50IjoiY3VybC83LjgxLjAiLCJjbGFpbV9pZCI6ImZhOGYyMmQ2LTFjMzgtNDcyMi1hMmE2LTE1YjllYTdiOThkYyJ9.K16dLycyImxMzcjbaLljN_wRgzeMN-QvcoRnUXU9Fb4"
  ]
}
```

## Try to access endpoint for validated users again 
```
$ curl localhost:7777/api/v1/user/info -sH "x-auth-token: $USER_TOKEN"  | jq 
{
  "email": "jon@test.com",
  "pass_hash": "deadbeef",
  "salt": "abc123",
  "role": "user",
  "created_at": "2024-10-07T00:59:22.976387731+01:00",
  "valid_email": true,
  "active_jwt": [
    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImpvbkB0ZXN0LmNvbSIsImNyZWF0ZWRfYXQiOiIyMDI0LTEwLTA3VDAwOjU5OjI1LjQwOTIxNTc3MiswMTowMCIsImV4cGlyZXNfYXQiOiIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsImlzX2FkbWluIjpmYWxzZSwic291cmNlX2FkZHJlc3MiOiIxMjcuMC4wLjEiLCJ1c2VyX2FnZW50IjoiY3VybC83LjgxLjAiLCJjbGFpbV9pZCI6ImZhOGYyMmQ2LTFjMzgtNDcyMi1hMmE2LTE1YjllYTdiOThkYyJ9.K16dLycyImxMzcjbaLljN_wRgzeMN-QvcoRnUXU9Fb4"
  ]
}
```

## Try to access admin endpoint 
```
$ curl localhost:7777/api/v1/admin/mail/list -sH "x-auth-token: $USER_TOKEN"  | jq 
{
  "message": "Unauthorized"
}
```

## Make user admin 
```
$ curl localhost:7777/api/v1/admin/user/make-admin -H 'content-type:application/json' -X PUT -sH "x-auth-token: $ADMIN_TOKEN" -d '{"email":"jon@test.com"}'
```

## Try to access admin endpoint 
```
$ curl localhost:7777/api/v1/admin/mail/list -sH "x-auth-token: $USER_TOKEN"  | jq 
{
  "mails": [
    {
      "subject": "verify your email",
      "link": "http://127.0.0.1:7777/api/v1/user/validate?email=mary@test.com&challenge=6382d9c1-31b4-49e5-989a-27c6914853e9",
      "to": "mary@test.com"
    }
  ]
}
```


