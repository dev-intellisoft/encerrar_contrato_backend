package pkg

import (
	"github.com/bwmarrin/snowflake"
	"log"
)

var Node *snowflake.Node

func InitNode() {
	var err error
	Node, err = snowflake.NewNode(1)
	if err != nil {
		log.Fatalf("Failed to inistialize snowflake node: %v", err)
	}
}
