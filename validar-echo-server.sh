#!/bin/bash

TEST_MESSAGE="${TEST_MESSAGE:-test}"
result=$(docker run --network tp0_testing_net --rm alpine /bin/sh -c "echo '$TEST_MESSAGE' | nc server 12345")

if [ "$result" == "$TEST_MESSAGE" ]; then
  echo "action: test_echo_server | result: success"
else
  echo "action: test_echo_server | result: fail"
fi