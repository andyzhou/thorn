# udp service

- provide base udp service

## generate proto
cd proto
protoc --go_out=plugins=grpc:. *.proto

## how to use?
please see sub dir `example`