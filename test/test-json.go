package main

import "fmt"
import "encoding/json"
import "../util"

var tests = []string{
`{
 a: "string value\n2nd line. Embedded \"quotes\"."
"b"= 42
c= true # comment
  
doof
:
 Hi, this works without "..."
e: { "f":"bar", "g": 99
     h: ["bla"
     "fasel",
     "dusel"
     ]
      }}`,
`a: "string value\n2nd line. Embedded \"quotes\"."
"b"= 42
c= true
  # comment
d: Hi, this works without "..."
e: { "f":"bar", "g": 99
     h: ["bla"
     "fasel",
     "dusel"
     ]
      }`,
` `,
`  {  }`,
`a=99
foobar=true
blafasel=["1","2","3"]
goro: onan
`,
}

func main() {
  var foo interface{}
  for _, test := range tests {
    js := util.AlmostJSON(test)
    fmt.Println(string(js))
    err := json.Unmarshal(js,&foo)
    if err != nil {
      fmt.Printf("ERROR: %v\n",err) 
    } else {
      fmt.Printf("Result: %#v\n",foo)
    }
  }
}
