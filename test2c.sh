#!/bin/bash

go test -v --count=1 --parallel=1 -p=1 ./raft -run 2C

go test -v --count=1 --parallel=1 -p=1 ./kv/test_raftstore -run TestOneSnapshot2C

go test -v --count=1 --parallel=1 -p=1 ./kv/test_raftstore -run TestSnapshotRecover2C

go test -v --count=1 --parallel=1 -p=1 ./kv/test_raftstore -run TestSnapshotRecoverManyClients2C

go test -v --count=1 --parallel=1 -p=1 ./kv/test_raftstore -run TestSnapshotUnreliable2C

go test -v --count=1 --parallel=1 -p=1 ./kv/test_raftstore -run TestSnapshotUnreliableRecover2C

go test -v --count=1 --parallel=1 -p=1 ./kv/test_raftstore -run TestSnapshotUnreliableRecoverConcurrentPartition2C

