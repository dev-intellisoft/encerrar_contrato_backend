package pkg

import (
	"github.com/google/uuid"
	"log"
)

//var Node *snowflake.Node

//func InitNode() {
//	var err error
//	Node, err = snowflake.NewNode(1)
//	if err != nil {
//		log.Fatalf("Failed to inistialize snowflake node: %v", err)
//	}
//}

func GenerateUUID() string {
	id, err := uuid.NewRandom()
	if err != nil {
		log.Fatalf("Failed to generate UUID: %v", err)
	}
	return id.String()
}
