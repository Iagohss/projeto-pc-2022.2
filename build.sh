#!/bin/sh

go build -o ./src/sequencial/sequencial ./src/sequencial/sequencial.go

go build -o ./src/concurrent/concurrent ./src/concurrent/concurrent.go