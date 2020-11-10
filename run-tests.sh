#!/bin/bash

go test -run TestCreateAccount_Success ./src/chats/tests/ -v -count 1
go test -run TestNewChat_Success ./src/chats/tests/ -v -count 1
