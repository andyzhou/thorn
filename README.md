# what's this
 
 This is a basic udp game service library.

# udp service

- provide base udp service

## generate proto
cd proto

protoc --go_out=plugins=grpc:. *.proto

## how to use?
please see sub dir `example`
