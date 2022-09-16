# Simple Go - Request / Response HTTP wrapper

## Usage

```
package main

import (
  "fmt"
  "github.com/kmatsoukas/gors"
)

func main() {
  client := gors.NewClient("https://api.example.com")

  client.SetDefaultHeaders(map[string]string{
    "Authorization": "Bearer my-token",
    "Content-Type":  "application/json",
  })

  req := client.NewRequest(gors.GET, "/my/endpoint")
  req.SetQuery("search", "value1")
  req.SetHeader("Another-Header", "Header-Value")

  res, err := req.Send()

  if err != nil {
    fmt.Fatal(err)
  }

  fmt.Println(string(res.Body))
}
```

## Parse JSON using Go generics

```
package main

type MyStruct struct {
	Field1    int    `json:"field1"`
	Field2    string `json:"Field2"`
}

import (
  "fmt"
  "github.com/kmatsoukas/gors"
)

func main() {
  client := gors.NewClient("https://api.example.com")

  client.SetDefaultHeaders(map[string]string{
    "Authorization": "Bearer my-token",
    "Content-Type":  "application/json",
  })

  req := client.NewRequest(gors.GET, "/my/endpoint")
  req.SetQuery("search", "value1")
  req.SetHeader("Another-Header", "Header-Value")

  res, err := gors.SendWithJSONResponse[MyStruct](req)

  if err != nil {
    fmt.Fatal(err)
  }

  fmt.Printf("%+v\n", res)
}
```