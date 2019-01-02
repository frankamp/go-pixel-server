package viewerclient

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type Scene struct {
	BaseFrame *Frame `json:"b"`
	Frames []Frame `json:"f"`
}

// available commands:
// thickness: 0 is filled (default), if its above that, it will outline the shape
// color: 0,0,0 (e.g. set the paint brush color, because im lazy these are floats between 0-1, not hex)
// shape: "rectangle" (default) or "circle"
// if you set it to circle, R then defines circle center x,y, followed by ellipse radius a, b (for a real circle a and b are the same)
type Command struct {
	Name  string `json:"n"`
	Value string `json:"v"`
}

type Coords [4]int

// element should either be a command or a region coordinate set
type Element struct {
	C *Command `json:"c"`
	R *Coords  `json:"r"`
}

type Frame struct {
	Elements []Element `json:"e"`
}

func Visualize(scene Scene) {
	jsonStr, _ := json.Marshal(scene)
	req, err := http.NewRequest("POST", "http://localhost:8080/scene", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}