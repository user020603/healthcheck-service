package client

import (
	"thanhnt208/healthcheck-service/proto/pb"

	"google.golang.org/grpc"
)

func StartGrpcClient() (pb.ContainerAdmServiceClient, error) {
	conn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	client := pb.NewContainerAdmServiceClient(conn)

	return client, nil
}
