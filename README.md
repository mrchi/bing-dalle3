# bing-dalle3

`bing-dalle3` is a Golang implementation of [yihong0618/BingImageCreator](https://github.com/yihong0618/BingImageCreator). It allows you to access the [Image Creator from Microsoft Bing](https://www.bing.com/images/create) service through API calls.

In comparison to [yihong0618/BingImageCreator](https://github.com/yihong0618/BingImageCreator), `bing-dalle3` simplifies certain features, including:
- The addition of a random `x-forwarded-for` header to disguise the source IP of requests (useful if Bing does not correctly implement code to obtain the real IP).
- The ability to save generated images to disk.
- Implementation of a CLI command.

## Quickstart 

Please refer to the README of [yihong0618/BingImageCreator](https://github.com/yihong0618/BingImageCreator) to obtain your Bing cookie.

```golang
package main

import (
	"fmt"
	"os"

	bingdalle3 "github.com/mrchi/bing-dalle3"
)

func main() {
	prompt := "月落乌啼霜满天，江枫渔火对愁眠。"
	bingClient := bingdalle3.NewBingDalle3("Your Bing cookie")

	balance, err := bingClient.GetTokenBalance()
	if err != nil {
		panic(err)
	}
	fmt.Println("balance: ", balance)

	writingId, err := bingClient.CreateImage(prompt)
	if err != nil {
		panic(err)
	}
	fmt.Println("writingId: ", writingId)

	imageUrls, err := bingClient.QueryResult(writingId, prompt)
	if err != nil {
		panic(err)
	}
	fmt.Println("imageUrls: ", imageUrls)

	imageContent, err := bingClient.DownloadImage(imageUrls[0])
	if err != nil {
		panic(err)
	}

	if err = os.WriteFile("output.jpeg", *imageContent, 0644); err != nil {
		panic(err)
	}
}
```

## Thanks

I'd like to express my gratitude to [@yihong0618](https://github.com/yihong0618) and the original author [@acheong08](https://github.com/acheong08) for their contributions.
