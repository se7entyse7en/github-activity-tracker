package main

import (
	"context"
	"fmt"
	"time"

	"github.com/se7entyse7en/github-activity-tracker/pkg/client"
)

func main() {
	c := client.NewClient("se7entyse7en")

	sinceStr := "2019-02-25T00:00:00.000Z"
	since, err := time.Parse(time.RFC3339, sinceStr)

	if err != nil {
		fmt.Println(err)
		return
	}

	toStr := "2019-02-28T00:00:00.000Z"
	to, err := time.Parse(time.RFC3339, toStr)

	if err != nil {
		fmt.Println(err)
		return
	}

	activity := c.GetActivity(context.Background(), true, &since, &to)
	fmt.Printf("%s\n", activity)
}
