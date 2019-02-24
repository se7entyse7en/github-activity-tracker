package main

import (
	"context"
	"fmt"

	"github.com/se7entyse7en/github-activity-tracker/pkg/client"
)

func main() {
	c := client.NewClient("se7entyse7en")
	activity := c.GetActivity(context.Background(), true, nil, nil)
	fmt.Printf("%s\n", activity)
}
