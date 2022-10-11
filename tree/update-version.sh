#!/usr/bin/env bash

cat > version.go <<STR
package main

var AppVersion = "$(date "+%Y-%m-%d-%H:%M:%S")"
STR
