#!/bin/bash 

rm -rf ./db_integration.bolt
go install ../../cmd/todo-app.go
export I_KEY=$(uuidgen)
export APP_ENV="I"
export I_DBPATH="./db_integration.bolt"
export I_LEVEL="debug"
export I_PORT=$((RANDOM % (50000 - 5000 + 1) + 5000))
export I_ADMIN_TOKEN=$(todo-app -c)
todo-app > "test_${I_PORT}.out" 2> "test_${I_PORT}.err" &
pid=$!
go test -test.v .
TEST_RES=$?
kill $pid
rm -rf ./db_integration.bolt
